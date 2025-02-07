/*
 * MIT License
 *
 * Copyright (c) 2022-2025 Arsene Tochemey Gandote
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
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

type txRunnerSuite struct {
	suite.Suite
	container *TestContainer
}

// SetupSuite starts the Postgres database engine and set the container
// host and port to use in the tests
func (s *txRunnerSuite) SetupSuite() {
	s.container = NewTestContainer("testdb", "test", "test", "public")
}

func (s *txRunnerSuite) TearDownSuite() {
	s.container.Cleanup()
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestTxRunnerSuite(t *testing.T) {
	suite.Run(t, new(txRunnerSuite))
}

func (s *txRunnerSuite) TestAddSQLBuilder() {
	s.T().Skip("")
	ctx := context.TODO()
	db := s.container.Testkit()

	err := db.Connect(ctx)
	s.Assert().NoError(err)

	txRunner, err := NewTxRunner(ctx, db)
	s.Assert().NoError(err)
	s.Assert().NotNil(txRunner)
	txRunner.
		AddSQLBuilder(new(mangoesInsertBuilder)).
		AddSQLBuilder(new(carsInsertBuilder))

	s.Assert().NotEmpty(txRunner.builders)
	s.Assert().Equal(2, len(txRunner.builders))

	err = db.Disconnect(ctx)
	s.Assert().NoError(err)
}

func (s *txRunnerSuite) TestRun() {
	s.Run("happy path", func() {
		ctx := context.TODO()
		db := s.container.Testkit()

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
			AddSQLBuilder(new(mangoesInsertBuilder)).
			AddSQLBuilder(new(carsInsertBuilder)).
			Run()
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
		db := s.container.Testkit()

		err := db.Connect(ctx)
		s.Assert().NoError(err)

		txRunner, err := NewTxRunner(ctx, db)
		s.Assert().NoError(err)
		s.Assert().NotNil(txRunner)

		// execute the transaction
		err = txRunner.
			AddSQLBuilder(new(failureBuilder)).
			AddSQLBuilder(new(carsInsertBuilder)).
			Run()
		s.Assert().Error(err)

		err = db.Disconnect(ctx)
		s.Assert().NoError(err)
	})
	s.Run("with all builders error", func() {
		ctx := context.TODO()
		db := s.container.Testkit()

		err := db.Connect(ctx)
		s.Assert().NoError(err)

		txRunner, err := NewTxRunner(ctx, db)
		s.Assert().NoError(err)
		s.Assert().NotNil(txRunner)

		// execute the transaction
		err = txRunner.
			AddSQLBuilder(new(errorSQLBuilder)).
			AddSQLBuilder(new(errorSQLBuilder)).
			Run()
		s.Assert().Error(err)

		err = db.Disconnect(ctx)
		s.Assert().NoError(err)
	})
}

type mangoesInsertBuilder struct{}

func (i *mangoesInsertBuilder) ToSQL() (sqlStatement string, args []any, err error) {
	args = []any{10, "succulent"}
	sqlStatement = `insert into mangoes(id, taste) values($1, $2);`
	return
}

type carsInsertBuilder struct{}

func (i *carsInsertBuilder) ToSQL() (sqlStatement string, args []any, err error) {
	args = []any{10, "black"}
	sqlStatement = `insert into cars(id, color) values($1, $2);`
	return
}

type failureBuilder struct{}

func (i *failureBuilder) ToSQL() (sqlStatement string, args []any, err error) {
	err = errors.New("failed to build query")
	return
}

type errorSQLBuilder struct{}

func (i *errorSQLBuilder) ToSQL() (sqlStatement string, args []any, err error) {
	args = []any{10, "black"}
	// this is an intended wrong sql statement
	sqlStatement = `insert into table cars(id, color) values($1, $2);`
	return
}
