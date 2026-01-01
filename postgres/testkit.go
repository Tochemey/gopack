// MIT License
//
// Copyright (c) 2022-2026 Arsene Tochemey Gandote
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" //nolint
	"github.com/testcontainers/testcontainers-go"
	pgcontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestContainer helps creates a Postgres docker container to
// run unit tests
// nolint
type TestContainer struct {
	testcontainers.Container
	host   string
	port   int
	schema string

	// connection credentials
	dbUser   string
	dbName   string
	dbPass   string
	dbSchema string
}

// NewTestContainer create a Postgres test container useful for unit and integration tests
// This function will exit when there is an error.Call this function inside your SetupTest to create the container before each test.
func NewTestContainer(dbName, dbUser, dbPassword, dbSchema string) *TestContainer {
	ctx := context.Background()
	postgresContainer, err := pgcontainer.Run(ctx,
		"postgres:16-alpine",
		pgcontainer.WithDatabase(dbName),
		pgcontainer.WithUsername(dbUser),
		pgcontainer.WithPassword(dbPassword),
		pgcontainer.WithSQLDriver("pgx"),
		testcontainers.CustomizeRequest(testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Cmd: []string{"-c", "log_statement=all"},
			},
		}),
		testcontainers.WithWaitStrategy(
			// Then, we wait for docker to actually serve the port on localhost.
			// For non-linux OSes like Mac and Windows, Docker or Rancher Desktop will have to start a separate proxy.
			// Without this, the tests will be flaky on those OSes!
			wait.ForListeningPort("5432/tcp"),
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(120*time.Second)))

	// handle the error
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	host, err := postgresContainer.Host(ctx)
	if err != nil {
		log.Fatalf("fail to get testContainer host: %v", err)
	}

	port, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		log.Fatalf("fail to get testContainer port: %v", err)
	}

	//dbURL, err := postgresContainer.ConnectionString(ctx)

	// create an instance of TestContainer
	testContainer := new(TestContainer)
	testContainer.Container = postgresContainer

	// set the testContainer host, port and schema
	testContainer.dbName = dbName
	testContainer.dbUser = dbUser
	testContainer.dbPass = dbPassword
	testContainer.schema = dbSchema
	testContainer.host = host
	testContainer.port = port.Int()
	return testContainer
}

// Testkit returns a Postgres Testkit that can be used in the tests
// to perform some database queries
func (c TestContainer) Testkit() *Testkit {
	return &Testkit{
		New(&Config{
			DBHost:                c.host,
			DBPort:                c.port,
			DBName:                c.dbName,
			DBUser:                c.dbUser,
			DBPassword:            c.dbPass,
			DBSchema:              c.schema,
			MaxConnections:        4,
			MinConnections:        0,
			MaxConnectionLifetime: time.Hour,
			MaxConnIdleTime:       30 * time.Minute,
			HealthCheckPeriod:     time.Minute,
		}),
	}
}

// Host return the host of the test container
func (c TestContainer) Host() string {
	return c.host
}

// Port return the port of the test container
func (c TestContainer) Port() int {
	return c.port
}

// Schema return the test schema of the test container
func (c TestContainer) Schema() string {
	return c.schema
}

// Cleanup frees the resource by removing a container and linked volumes from docker.
// Call this function inside your TearDownSuite to clean-up resources after each test
func (c TestContainer) Cleanup() {
	ctx := context.Background()
	if err := c.Container.Terminate(ctx); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
}

// Testkit is used in test to perform
// some database queries
type Testkit struct {
	Postgres
}

// DropTable utility function to drop a database table
func (c Testkit) DropTable(ctx context.Context, tableName string) error {
	var dropSQL = fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE;", tableName)
	_, err := c.Exec(ctx, dropSQL)
	return err
}

// TableExists utility function to help check the existence of table in Postgres
// tableName is in the format: <schemaName.tableName>. e.g: public.users
func (c Testkit) TableExists(ctx context.Context, tableName string) error {
	var stmt = fmt.Sprintf("SELECT to_regclass('%s');", tableName)
	_, err := c.Exec(ctx, stmt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	return nil
}

// Count utility function to help count the number of rows in a Postgres table.
// tableName is in the format: <schemaName.tableName>. e.g: public.users
// It returns -1 when there is an error
func (c Testkit) Count(ctx context.Context, tableName string) (int, error) {
	var count int
	if err := c.Select(ctx, &count, fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)); err != nil {
		return -1, err
	}
	return count, nil
}

// CreateSchema helps create a test schema in a Postgres database
func (c Testkit) CreateSchema(ctx context.Context, schemaName string) error {
	stmt := fmt.Sprintf("CREATE SCHEMA %s", schemaName)
	if _, err := c.Exec(ctx, stmt); err != nil {
		return err
	}
	return nil
}

// SchemaExists helps check the existence of a Postgres schema. Very useful when implementing tests
func (c Testkit) SchemaExists(ctx context.Context, schemaName string) (bool, error) {
	stmt := fmt.Sprintf("SELECT schema_name FROM information_schema.schemata WHERE schema_name = '%s';", schemaName)
	var check string
	if err := c.Select(ctx, &check, stmt); err != nil {
		return false, err
	}

	// this redundant check is necessary
	if check == schemaName {
		return true, nil
	}

	return false, nil
}

// DropSchema utility function to drop a database schema
func (c Testkit) DropSchema(ctx context.Context, schemaName string) error {
	var dropSQL = fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE;", schemaName)
	_, err := c.Exec(ctx, dropSQL)
	return err
}
