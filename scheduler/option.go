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

package scheduler

import (
	"time"

	"github.com/tochemey/gopack/log"
)

// Option defines a configuration option that can be applied to a JobsScheduler.
//
// Implementations of this interface modify the scheduler's configuration when applied.
type Option interface {
	// Apply applies the configuration option to the given JobsScheduler instance.
	Apply(*JobsScheduler)
}

// enforce compilation error if OptionFunc does not implement Option
var _ Option = OptionFunc(nil)

// OptionFunc is a function type that implements the Option interface.
//
// It allows functions to be used as configuration options for JobsScheduler.
type OptionFunc func(*JobsScheduler)

// Apply applies the OptionFunc to the given JobsScheduler.
//
// This enables the use of functions as dynamic configuration options.
func (f OptionFunc) Apply(scheduler *JobsScheduler) {
	f(scheduler)
}

// WithLogger configures the scheduler to use a custom logger.
//
// Parameters:
//   - logger: An instance of log.Logger used for logging scheduler events.
//
// Returns:
//   - An Option that applies the custom logger to the JobsScheduler.
//
// Usage:
//
//	scheduler := NewJobsScheduler(WithLogger(myLogger))
func WithLogger(logger log.Logger) Option {
	return OptionFunc(
		func(scheduler *JobsScheduler) {
			scheduler.logger = logger
		},
	)
}

// WithStopTimeout configures a custom timeout duration for stopping the scheduler.
//
// Parameters:
//   - timeout: A time.Duration value defining how long the scheduler should wait
//     for running jobs to complete before shutting down.
//
// Returns:
//   - An Option that applies the stop timeout to the JobsScheduler.
//
// Usage:
//
//	scheduler := NewJobsScheduler(WithStopTimeout(5 * time.Second))
func WithStopTimeout(timeout time.Duration) Option {
	return OptionFunc(func(scheduler *JobsScheduler) {
		scheduler.stopTimeout = timeout
	})
}
