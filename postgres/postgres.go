package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/XSAM/otelsql"
	"github.com/georgysavva/scany/sqlscan"
	_ "github.com/lib/pq" //nolint
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
)

// Postgres will be implemented by concrete RDBMS store
type Postgres interface {
	// Connect connects to the underlying database
	Connect(ctx context.Context) error
	// Disconnect closes the underlying opened underlying connection database
	Disconnect(ctx context.Context) error
	// Select fetches a single row from the database and automatically scanned it into the dst.
	// It returns an error in case of failure. When there is no record no errors is return.
	Select(ctx context.Context, dst any, query string, args ...any) error
	// SelectAll fetches a set of rows as defined by the query and scanned those record in the dst.
	// It returns nil when there is no records to fetch.
	SelectAll(ctx context.Context, dst any, query string, args ...any) error
	// Exec executes an SQL statement against the database and returns the appropriate result or an error.
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
	// BeginTx helps start an SQL transaction. The return transaction object is expected to be used in
	// the subsequent queries following the BeginTx.
	BeginTx(ctx context.Context, txOptions *sql.TxOptions) (*sql.Tx, error)
}

// Postgres helps interact with the Postgres database
type postgres struct {
	connStr      string
	dbConnection *sql.DB
	config       *Config
}

var _ Postgres = (*postgres)(nil)

const postgresDriver = "postgres"
const instrumentationName = "ego.events_store"

// New returns a store connecting to the given Postgres database.
func New(config *Config) Postgres {
	postgres := new(postgres)
	postgres.config = config
	postgres.connStr = createConnectionString(config.DBHost, config.DBPort, config.DBName, config.DBUser, config.DBPassword, config.DBSchema)
	return postgres
}

// Connect will connect to our Postgres database
func (p *postgres) Connect(ctx context.Context) error {
	// Register an OTel driver
	driverName, err := otelsql.Register(postgresDriver, otelsql.WithAttributes(semconv.DBSystemPostgreSQL))
	if err != nil {
		return errors.Wrap(err, "failed to hook the tracer to the database driver")
	}

	// open the connection and connect to the database
	db, err := sql.Open(driverName, p.connStr)
	if err != nil {
		return errors.Wrap(err, "failed to open connection")
	}

	// let us test the connection
	err = db.PingContext(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to ping database connection")
	}

	// set connection setting
	db.SetMaxOpenConns(p.config.MaxOpenConnections)
	db.SetMaxIdleConns(p.config.MaxIdleConnections)
	db.SetConnMaxLifetime(p.config.ConnectionMaxLifetime)

	// set the db handle
	p.dbConnection = db
	return nil
}

// createConnectionString will create the Postgres connection string from the
// supplied connection details
func createConnectionString(host string, port int, name, user string, password string, schema string) string {
	info := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", host, port, user, name)
	// The Postgres driver gets confused in cases where the user has no password
	// set but a password is passed, so only set password if its non-empty
	if password != "" {
		info += fmt.Sprintf(" password=%s", password)
	}

	if schema != "" {
		info += fmt.Sprintf(" search_path=%s", schema)
	}

	return info
}

// Exec executes a sql query without returning rows against the database
func (p *postgres) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	// Create a span
	tracer := otel.GetTracerProvider()
	spanCtx, span := tracer.Tracer(instrumentationName).Start(ctx, "Exec")
	defer span.End()
	return p.dbConnection.ExecContext(spanCtx, query, args...)
}

// BeginTx starts a new database transaction
func (p *postgres) BeginTx(ctx context.Context, txOptions *sql.TxOptions) (*sql.Tx, error) {
	// Create a span
	tracer := otel.GetTracerProvider()
	spanCtx, span := tracer.Tracer(instrumentationName).Start(ctx, "BeginTx")
	defer span.End()
	return p.dbConnection.BeginTx(spanCtx, txOptions)
}

// SelectAll fetches rows
func (p *postgres) SelectAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	// Create a span
	tracer := otel.GetTracerProvider()
	spanCtx, span := tracer.Tracer(instrumentationName).Start(ctx, "SelectAll")
	defer span.End()
	err := sqlscan.Select(spanCtx, p.dbConnection, dst, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		return err
	}
	return nil
}

// Select fetches only one row
func (p *postgres) Select(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	// Create a span
	tracer := otel.GetTracerProvider()
	spanCtx, span := tracer.Tracer(instrumentationName).Start(ctx, "Select")
	defer span.End()
	err := sqlscan.Get(spanCtx, p.dbConnection, dst, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	return nil
}

// Disconnect the database connection.
func (p *postgres) Disconnect(ctx context.Context) error {
	tracer := otel.GetTracerProvider()
	_, span := tracer.Tracer(instrumentationName).Start(ctx, "Disconnect")
	defer span.End()
	if p.dbConnection == nil {
		return nil
	}
	return p.dbConnection.Close()
}
