/*
 * MIT License
 *
 * Copyright (c) 2022-2024 Arsene Tochemey Gandote
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package postgres

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
)

type txRunnerSuite struct {
	suite.Suite
	container *TestContainer
}

// SetupSuite starts the Postgres database engine and set the container
// host and port to use in the tests
func (s *txRunnerSuite) SetupSuite() {
	s.container = NewTestContainer("testdb", "test", "test")
}

func (s *txRunnerSuite) TearDownSuite() {
	s.container.Cleanup()
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestTxRunnerSuite(t *testing.T) {
	suite.Run(t, new(txRunnerSuite))
}

func (s *txRunnerSuite) TestAddQueryBuilder() {
	ctx := context.TODO()
	db := s.container.GetTestDB()

	err := db.Connect(ctx)
	s.Assert().NoError(err)

	txRunner, err := NewTxRunner(ctx, db)
	s.Assert().NoError(err)
	s.Assert().NotNil(txRunner)
	txRunner.
		AddQueryBuilder(new(mangoesInsertBuilder)).
		AddQueryBuilder(new(carsInsertBuilder))
	s.Assert().NotEmpty(txRunner.builders)
	s.Assert().Equal(2, len(txRunner.builders))

	err = db.Disconnect(ctx)
	s.Assert().NoError(err)
}

func (s *txRunnerSuite) TestExecute() {
	s.Run("happy path", func() {
		ctx := context.TODO()
		db := s.container.GetTestDB()

		err := db.Connect(ctx)
		s.Assert().NoError(err)

		stmt := `create table mangoes(id integer, taste varchar(10));`
		_, err = db.Exec(ctx, stmt)
		s.Assert().NoError(err)

		err = db.TableExists(ctx, "public.mangoes")
		s.Assert().NoError(err)
		s.Assert().Nil(err)

		stmt = `create table cars(id integer, color varchar(10));`
		_, err = db.Exec(ctx, stmt)
		s.Assert().NoError(err)

		err = db.TableExists(ctx, "public.cars")
		s.Assert().NoError(err)
		s.Assert().Nil(err)

		// create an instance of TxRunner
		txRunner, err := NewTxRunner(ctx, db)
		s.Assert().NoError(err)
		s.Assert().NotNil(txRunner)

		// execute the transaction
		err = txRunner.
			AddQueryBuilder(new(mangoesInsertBuilder)).
			AddQueryBuilder(new(carsInsertBuilder)).
			Execute()
		s.Assert().NoError(err)

		count, err := db.Count(ctx, "public.mangoes")
		s.Assert().NoError(err)
		s.Assert().Equal(1, count)

		count, err = db.Count(ctx, "public.cars")
		s.Assert().NoError(err)
		s.Assert().Equal(1, count)

		err = db.DropTable(ctx, "mangoes")
		s.Assert().NoError(err)

		err = db.DropTable(ctx, "cars")
		s.Assert().NoError(err)

		err = db.Disconnect(ctx)
		s.Assert().NoError(err)
	})
	s.Run("with a builder error", func() {
		ctx := context.TODO()
		db := s.container.GetTestDB()

		err := db.Connect(ctx)
		s.Assert().NoError(err)

		txRunner, err := NewTxRunner(ctx, db)
		s.Assert().NoError(err)
		s.Assert().NotNil(txRunner)

		// execute the transaction
		err = txRunner.
			AddQueryBuilder(new(failureBuilder)).
			AddQueryBuilder(new(carsInsertBuilder)).
			Execute()
		s.Assert().Error(err)
		s.Assert().EqualError(err, "failed to build query")

		err = db.Disconnect(ctx)
		s.Assert().NoError(err)
	})
	s.Run("with all builders error", func() {
		ctx := context.TODO()
		db := s.container.GetTestDB()

		err := db.Connect(ctx)
		s.Assert().NoError(err)

		txRunner, err := NewTxRunner(ctx, db)
		s.Assert().NoError(err)
		s.Assert().NotNil(txRunner)

		// execute the transaction
		err = txRunner.
			AddQueryBuilder(new(errorSQLBuilder)).
			AddQueryBuilder(new(errorSQLBuilder)).
			Execute()
		s.Assert().Error(err)
		errMsg := `pq: syntax error at or near "table"`
		s.Assert().EqualError(err, errMsg)

		err = db.Disconnect(ctx)
		s.Assert().NoError(err)
	})
}

type mangoesInsertBuilder struct{}

func (i *mangoesInsertBuilder) BuildQuery() (sqlStatement string, args []any, err error) {
	args = []any{10, "succulent"}
	sqlStatement = `insert into mangoes(id, taste) values($1, $2);`
	return
}

type carsInsertBuilder struct{}

func (i *carsInsertBuilder) BuildQuery() (sqlStatement string, args []any, err error) {
	args = []any{10, "black"}
	sqlStatement = `insert into cars(id, color) values($1, $2);`
	return
}

type failureBuilder struct{}

func (i *failureBuilder) BuildQuery() (sqlStatement string, args []any, err error) {
	err = errors.New("failed to build query")
	return
}

type errorSQLBuilder struct{}

func (i *errorSQLBuilder) BuildQuery() (sqlStatement string, args []any, err error) {
	args = []any{10, "black"}
	// this is an intended wrong sql statement
	sqlStatement = `insert into table cars(id, color) values($1, $2);`
	return
}
