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
	"sync"
	"sync/atomic"
	"time"

	quartzjob "github.com/reugn/go-quartz/job"
	"github.com/reugn/go-quartz/quartz"
	"go.opentelemetry.io/otel"

	"github.com/tochemey/gopack/log"
	"github.com/tochemey/gopack/log/zapl"
)

// Job represents a task that can be scheduled and executed by the JobsScheduler.
//
// Any struct implementing this interface can be scheduled for execution.
//
// Methods:
//   - ID(): Returns a unique identifier for the job.
//   - Run(ctx): Executes the job logic within the provided context.
//
// Implementations should ensure:
//   - The ID is unique to prevent conflicts in the scheduler.
//   - The Run method handles errors gracefully and respects context cancellation.
//
// Example:
//
//	type MyJob struct {
//	    id string
//	}
//
//	func (j *MyJob) ID() string {
//	    return j.id
//	}
//
//	func (j *MyJob) Run(ctx context.Context) error {
//	    // Job execution logic
//	    return nil
//	}
//
//	scheduler.Schedule(ctx, "* * * * *", &MyJob{id: "job-123"})
type Job interface {
	// ID returns a unique identifier for the job.
	ID() string
	// Run executes the job logic within the given context.
	//
	// The job should check ctx.Done() and handle cancellation appropriately.
	Run(ctx context.Context) error
}

// Scheduler will be implemented by the scheduler
type Scheduler interface {
	// Start initializes and starts the scheduler, executing all scheduled jobs.
	//
	// Each job runs in its own separate goroutine, ensuring non-blocking execution.
	// The scheduler continues running until the provided context is canceled.
	//
	// Parameters:
	//   - ctx: A context used to control the scheduler's lifecycle. Canceling the context stops the scheduler.
	//
	// Note:
	//   - Jobs are executed asynchronously in independent goroutines.
	//   - Ensure proper error handling and resource cleanup within jobs to prevent unexpected behavior.
	Start(ctx context.Context)
	// Stop gracefully shuts down the scheduler and stops any running jobs.
	//
	// Parameters:
	//   - ctx: A context used to control the shutdown process. If the context expires before
	//          all jobs are stopped, the function may return an error.
	//
	// Behavior:
	//   - Attempts to stop all scheduled jobs cleanly.
	//   - Running jobs may be interrupted depending on their implementation.
	//   - Ensures no new jobs are started after invocation.
	//
	// Returns:
	//   - An error if the shutdown process fails or is interrupted.
	Stop(ctx context.Context) error
	// Schedule adds a new job runner to the scheduler.
	//
	// Parameters:
	//   - ctx: A context to control execution and cancellation.
	//   - cronExpression: A string representing the scheduling expression, defining when the job should run.
	//   - job: The job instance to be scheduled.
	//
	// The cronExpression supports multiple formats:
	//   - Standard crontab syntax (minute, hour, day, month, weekday), e.g., "0 12 * * ?"
	//   - Extended format with an optional second field, e.g., "0 * * * * *"
	//   - Predefined descriptors for common schedules, e.g., "@midnight", "@every 1h30m"
	//
	// Returns:
	//   - An error if the scheduling fails due to an invalid expression or other internal issues.
	Schedule(ctx context.Context, cronExpression string, job Job) error
}

// JobsScheduler implements Scheduler
type JobsScheduler struct {
	mu *sync.Mutex
	// underlying Scheduler
	quartzScheduler quartz.Scheduler
	jobs            map[string]Job
	logger          log.Logger
	started         atomic.Bool
	stopTimeout     time.Duration
}

// enforce a compilation error
var _ Scheduler = &JobsScheduler{}

// NewJobsScheduler creates and returns a new instance of JobsScheduler.
//
// Parameters:
//   - opts: A variadic list of Option functions to customize the scheduler configuration.
//
// Behavior:
//   - Applies each provided Option to configure the scheduler.
//   - Initializes internal data structures needed for job scheduling.
//   - The scheduler does not start automatically; call Start(ctx) to begin execution.
//
// Returns:
//   - A pointer to a fully initialized JobsScheduler instance.
//
// Example:
//
//	scheduler := NewJobsScheduler(
//	    WithLogger(myLogger),
//	    WithStopTimeout(10 * time.Second),
//	)
//	scheduler.Start(ctx)
func NewJobsScheduler(opts ...Option) *JobsScheduler {
	scheduler := &JobsScheduler{
		jobs:        make(map[string]Job),
		logger:      zapl.New(log.InfoLevel, os.Stdout),
		started:     atomic.Bool{},
		stopTimeout: 3 * time.Second,
		mu:          &sync.Mutex{},
	}

	// apply options here
	// set the custom options to override the default values
	for _, opt := range opts {
		opt.Apply(scheduler)
	}

	quartzScheduler, _ := quartz.NewStdScheduler(quartz.WithLogger(newLogWrapper(scheduler.logger)))
	scheduler.quartzScheduler = quartzScheduler
	return scheduler
}

// Start starts the scheduler and run all the jobs in their separate go-routine
func (s *JobsScheduler) Start(ctx context.Context) {
	// Create a span
	tracer := otel.GetTracerProvider()
	_, span := tracer.Tracer("").Start(ctx, "Start")
	defer span.End()
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Info("starting Jobs Scheduler...")
	s.quartzScheduler.Start(ctx)
	s.started.Store(s.quartzScheduler.IsStarted())
	s.logger.Info("Jobs Scheduler started.:)")
}

// Stop shutdowns the Scheduler gracefully
func (s *JobsScheduler) Stop(ctx context.Context) error {
	// Create a span
	tracer := otel.GetTracerProvider()
	_, span := tracer.Tracer("").Start(ctx, "Stop")
	defer span.End()

	if !s.started.Load() {
		return nil
	}

	s.logger.Info("stopping Jobs Scheduler...")
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.quartzScheduler.Clear(); err != nil {
		return err
	}

	s.quartzScheduler.Stop()
	s.started.Store(s.quartzScheduler.IsStarted())

	ctx, cancel := context.WithTimeout(ctx, s.stopTimeout)
	defer cancel()
	s.quartzScheduler.Wait(ctx)

	s.logger.Info("Jobs Scheduler stopped...:)")
	return nil
}

// Schedule adds a new job runner to the scheduler.
//
// Parameters:
//   - ctx: A context to control execution and cancellation.
//   - cronExpression: A string representing the scheduling expression, defining when the job should run.
//   - job: The job instance to be scheduled.
//
// The cronExpression supports multiple formats:
//   - Standard crontab syntax (minute, hour, day, month, weekday), e.g., "0 12 * * ?"
//   - Extended format with an optional second field, e.g., "0 * * * * *"
//   - Predefined descriptors for common schedules, e.g., "@midnight", "@every 1h30m"
//
// Returns:
//   - An error if the scheduling fails due to an invalid expression or other internal issues.
func (s *JobsScheduler) Schedule(ctx context.Context, cronExpression string, job Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started.Load() {
		return ErrSchedulerNotStarted
	}

	// check whether the job has been not been added already
	if _, ok := s.jobs[job.ID()]; ok {
		return fmt.Errorf("job (%s) is already added", job.ID())
	}

	// create the actual job to run
	actualJob := quartzjob.NewFunctionJob[bool](
		func(ctx context.Context) (bool, error) {
			if err := job.Run(ctx); err != nil {
				return false, err
			}
			return true, nil
		},
	)

	// create the job details
	details := quartz.NewJobDetail(actualJob, quartz.NewJobKey(job.ID()))
	// set the location
	location := time.Now().Location()
	// create the trigger
	trigger, err := quartz.NewCronTriggerWithLoc(cronExpression, location)
	if err != nil {
		s.logger.Error(fmt.Errorf("failed to schedule message: %w", err))
		return err
	}

	// schedule the job
	if err := s.quartzScheduler.ScheduleJob(details, trigger); err != nil {
		s.logger.Error(fmt.Errorf("failed to schedule message: %w", err))
		return err
	}

	// let us add the job
	s.jobs[job.ID()] = job
	return nil
}
