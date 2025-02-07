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

package pubsub

import (
	"cloud.google.com/go/pubsub/pstest"
)

// Emulator helps creates a GCP PubSub emulator to
// run unit tests
type Emulator struct {
	endPoint string
	srv      *pstest.Server
}

// NewEmulator create a PubSub test server useful for unit and integration tests
// This function will exit when there is an error.Call this function inside your SetupTest to create the container before each test.
func NewEmulator() *Emulator {
	// Start a fake server running locally.
	srv := pstest.NewServer()
	// create the instance of the Emulator
	emulator := &Emulator{
		endPoint: srv.Addr,
		srv:      srv,
	}

	return emulator
}

// Cleanup frees the resource by removing a container and linked volumes from docker.
// Call this function inside your TearDownSuite to clean-up resources after each test
func (c Emulator) Cleanup() error {
	return c.srv.Close()
}

// EndPoint return the endpoint of the Emulator
func (c Emulator) EndPoint() string {
	return c.endPoint
}

// Server return the server object of the Emulator
func (c Emulator) Server() *pstest.Server {
	return c.srv
}
