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

package scheduler

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"go.opentelemetry.io/otel"
)

// the cronAgent expression parser
var cronExpressionParser = cron.NewParser(
	cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
)

// Job will be implemented by any job runner
type Job interface {
	// ID returns the Job unique identifier
	ID() string
	// Run execute the job
	Run(ctx context.Context) error
}

// Scheduler will be implemented by the scheduler
type Scheduler interface {
	// Start starts the scheduler and run all the jobs in their separate go-routine
	Start(ctx context.Context)
	// Stop stops the scheduler and stop any running job
	Stop(ctx context.Context) error
	// Run runs the scheduler by executing all jobs that have been added to it.
	Run(ctx context.Context)
	// AddJob add a new job runner to the scheduler. The jobID and cronExpression is required.
	// It accepts for cronExpression
	//   - Standard crontab specs, e.g. "* * * * ?"
	//   - With optional second field, e.g. "* * * * * ?"
	//   - Descriptors, e.g. "@midnight", "@every 1h30m"
	AddJob(ctx context.Context, cronExpression string, job Job) error
}

// JobsScheduler implements Scheduler
type JobsScheduler struct {
	mu        sync.Mutex
	scheduler *gocron.Scheduler
	jobs      map[string]Job
}

// enforce a compilation error
var _ Scheduler = &JobsScheduler{}

// NewJobsScheduler creates a new instance of Scheduler.
// It accepts for cronExpression
//   - Standard crontab specs, e.g. "* * * * ?"
//   - With optional second field, e.g. "* * * * * ?"
//   - Descriptors, e.g. "@midnight", "@every 1h30m"
func NewJobsScheduler() *JobsScheduler {
	return &JobsScheduler{
		mu:        sync.Mutex{},
		scheduler: gocron.NewScheduler(time.UTC),
		jobs:      make(map[string]Job),
	}
}

// Start starts the scheduler and run all the jobs in their separate go-routine
func (s *JobsScheduler) Start(ctx context.Context) {
	// Create a span
	tracer := otel.GetTracerProvider()
	_, span := tracer.Tracer("").Start(ctx, "Start")
	defer span.End()
	// set the panic handler
	gocron.SetPanicHandler(func(jobName string, recoverData interface{}) {
		// TODO add some logging or a listener
		fmt.Printf("Panic in job: %s", jobName)
	})

	// start the cron jobs
	s.scheduler.StartAsync()
}

// Stop shutdowns the Scheduler gracefully
func (s *JobsScheduler) Stop(ctx context.Context) error {
	// Create a span
	tracer := otel.GetTracerProvider()
	_, span := tracer.Tracer("").Start(ctx, "Start")
	defer span.End()

	// stop the scheduler
	s.scheduler.Stop()
	return nil
}

// AddJob adds new Job to the scheduler. If the job already exists rejects the request.
func (s *JobsScheduler) AddJob(ctx context.Context, cronExpression string, job Job) error {
	// acquire the lock
	s.mu.Lock()
	// release lock when done
	defer s.mu.Unlock()

	// validate the cron expression
	if _, err := cronExpressionParser.Parse(cronExpression); err != nil {
		// return error
		return err
	}

	// check whether the job has been not been added already
	if _, ok := s.jobs[job.ID()]; ok {
		return fmt.Errorf("job (%s) is already added", job.ID())
	}

	// add the cron job
	_, err := s.scheduler.
		CronWithSeconds(cronExpression).
		Name(job.ID()).
		Tag(job.ID()).
		SingletonMode().Do(func() {
		// hook the job execution
		if err := job.Run(ctx); err != nil {
			// hook a recovery mechanism to the scheduler to handle the panic
			panic(errors.Wrapf(err, "job (%s) failed to run", job.ID()))
		}
	})

	// handle the error
	if err != nil {
		// return error
		return err
	}

	// let us add the job
	s.jobs[job.ID()] = job
	return nil
}

// Run runs the scheduler by executing all jobs that have been added to it.
func (s *JobsScheduler) Run(ctx context.Context) {
	// start the jobs scheduler
	s.Start(ctx)
	// await signal to shut down
	interruptSignal := make(chan os.Signal, 1)
	signal.Notify(interruptSignal, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	shutdownComplete := make(chan struct{})
	go func() {
		<-interruptSignal
		if err := s.Stop(ctx); err != nil {
			panic(errors.Wrap(err, "unable to shutdown the scheduler service"))
		}
		close(shutdownComplete)
	}()
	<-shutdownComplete
	os.Exit(0)
}
