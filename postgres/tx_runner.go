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
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

// SQLBuilder (like squirrel builders) implements ToSQL that
// returns the SQL statement to run against the database
type SQLBuilder interface {
	// ToSQL returns the SQL statement and arguments to run
	// The args are used to handle prepared statements. Therefore, they must be provided
	// in the order of the various placeholder for smooth substitution
	ToSQL() (stmt string, args []any, err error)
}

// squirrelAdapter transform a squirrel Sqlizer into a SQLBuilder
// that can be run in a transaction
type squirrelAdapter struct {
	s sq.Sqlizer
}

func (s squirrelAdapter) ToSQL() (string, []any, error) {
	return s.s.ToSql()
}

// TxRunner helps run SQL statements in a safe database transaction.
// In case of errors the underlying transaction is rolled back
// When there are no errors the underlying transaction is automatically committed
type TxRunner struct {
	tx       pgx.Tx
	builders []SQLBuilder
	ctx      context.Context
}

// NewTxRunner creates an instance of TxRunner
func NewTxRunner(ctx context.Context, db Postgres) (*TxRunner, error) {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return &TxRunner{
		tx:       tx,
		builders: nil,
		ctx:      ctx,
	}, nil
}

// AddSQLBuilder adds a SQL builder to the TxRunner
func (runner *TxRunner) AddSQLBuilder(builder SQLBuilder) *TxRunner {
	runner.builders = append(runner.builders, builder)
	return runner
}

// AddSQLBuilders an array of SQL builder to the TxRunner
func (runner *TxRunner) AddSQLBuilders(builders ...SQLBuilder) *TxRunner {
	runner.builders = append(runner.builders, builders...)
	return runner
}

// AddSqlizer adds a squirrel builder to the TxRunner
func (runner *TxRunner) AddSqlizer(s sq.Sqlizer) *TxRunner {
	runner.builders = append(runner.builders, squirrelAdapter{s: s})
	return runner
}

// Run executes the database transaction and returns the resulting error.
// In case of errors the underlying transaction is rolled back
// When there are no errors the underlying transaction is automatically committed
func (runner *TxRunner) Run() error {
	type stmt struct {
		query string
		args  []any
	}

	// build the SQL statements to execute with the database transaction
	// rollback the transaction when there is an error
	stmts := make([]stmt, 0, len(runner.builders))
	for _, builder := range runner.builders {
		q, args, err := builder.ToSQL()
		if err != nil {
			// rollback the transaction
			if rollbackErr := runner.tx.Rollback(runner.ctx); rollbackErr != nil {
				return fmt.Errorf("failed to rollback transaction: %w", rollbackErr)
			}
			return fmt.Errorf("failed to build query: %w", err)
		}

		stmts = append(stmts, stmt{
			query: q,
			args:  args,
		})
	}

	// execute the SQL statements build with the database transaction
	for _, stmt := range stmts {
		if _, err := runner.tx.Exec(runner.ctx, stmt.query, stmt.args...); err != nil {
			// rollback the transaction
			if rollbackErr := runner.tx.Rollback(runner.ctx); rollbackErr != nil {
				return fmt.Errorf("failed to rollback transaction: %w", rollbackErr)
			}
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return runner.tx.Commit(runner.ctx)
}
