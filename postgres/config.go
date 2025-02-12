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

import "time"

// Config defines the postgres configuration
// This configuration does not take into consideration the SSL mode
// TODO: enhance with SSL mode
type Config struct {
	DBHost                string        // DBHost represents the database host
	DBPort                int           // DBPort is the database port
	DBName                string        // DBName is the database name
	DBUser                string        // DBUser is the database user used to connect
	DBPassword            string        // DBPassword is the database password
	DBSchema              string        // DBSchema represents the database schema
	MaxConnections        int           // MaxConnections represents the number of max connections in the pool
	MinConnections        int           // MinConnections represents the number of minimum connections in the pool
	MaxConnectionLifetime time.Duration // MaxConnectionLifetime represents the duration since creation after which a connection will be automatically closed.
	MaxConnIdleTime       time.Duration // MaxConnIdleTime is the duration after which an idle connection will be automatically closed by the health check.
	HealthCheckPeriod     time.Duration // HeathCheckPeriod is the duration between checks of the health of idle connections.
}

// NewConfig creates an instance of Config
func NewConfig(host string, port int, user, password, dbName string) *Config {
	return &Config{
		DBHost:                host,
		DBPort:                port,
		DBName:                dbName,
		DBUser:                user,
		DBPassword:            password,
		MaxConnections:        4,
		MinConnections:        0,
		MaxConnectionLifetime: time.Hour,
		MaxConnIdleTime:       30 * time.Minute,
		HealthCheckPeriod:     time.Minute,
	}
}
