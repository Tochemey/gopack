package postgres

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
)

// QueryBuilder interface generalizes the sql execution implementations
type QueryBuilder interface {
	// BuildQuery return the SQL statement and arguments to use to run the query
	// The args are used to handle prepared statement. Therefore, they must be provided in the order
	// of the various placeholder for smooth substitution.
	BuildQuery() (sqlStatement string, args []any, err error)
}

// TxRunner helps run database queries in a safe database transaction.
// In case of errors the underlying database transaction is rolled back.
// When there is no errors the underlying database transaction is committed.
type TxRunner struct {
	tx       *sql.Tx
	builders []QueryBuilder

	ctx context.Context
}

// NewTxRunner creates an instance of TxRunner
func NewTxRunner(ctx context.Context, db Postgres) (*TxRunner, error) {
	// create a db transaction
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}

	// create an instance of TxRunner
	txRunner := &TxRunner{
		tx:       tx,
		ctx:      ctx,
		builders: make([]QueryBuilder, 0),
	}

	// create an instance of TxRunner and returns
	return txRunner, nil
}

// AddQueryBuilder adds an QueryBuilder to the db transaction runner.
// The Builders queries will be executed in the database transaction in the order
// the builders have been added. Therefore, add builders according to the order of execution of the queries
func (r *TxRunner) AddQueryBuilder(v QueryBuilder) *TxRunner {
	r.builders = append(r.builders, v)
	return r
}

// Execute executes the database queries returns resulting error(s).
// In case of errors the underlying database transaction is rolled back.
// When there is no errors the underlying database transaction is committed
func (r *TxRunner) Execute() error {
	// create a type to hold the query and arguments
	type queryArgs struct {
		statement string
		args      []any
	}

	// let us build the query and args
	queries := make([]queryArgs, 0, len(r.builders))
	for _, builder := range r.builders {
		// build the query
		query, args, err := builder.BuildQuery()
		if err != nil {
			// rollback the transaction
			if rollbackErr := r.tx.Rollback(); rollbackErr != nil {
				return errors.Wrap(err, rollbackErr.Error())
			}

			return err
		}

		// add to the queries
		queries = append(queries, queryArgs{
			statement: query,
			args:      args,
		})
	}

	for _, query := range queries {
		if _, execErr := r.tx.ExecContext(r.ctx, query.statement, query.args...); execErr != nil {
			// rollback the transaction
			if rollbackErr := r.tx.Rollback(); rollbackErr != nil {
				return errors.Wrap(execErr, rollbackErr.Error())
			}

			return execErr
		}
	}

	// commit the database transaction
	return r.tx.Commit()
}
