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
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type testJob struct {
	id string
	wg *sync.WaitGroup
}

func (j *testJob) Run(context.Context) error {
	j.wg.Done()
	return nil
}

func (j *testJob) ID() string {
	return j.id
}

type testLongRunningJob struct {
	id string
}

func (j *testLongRunningJob) ID() string {
	return j.id
}

func (j *testLongRunningJob) Run(context.Context) error {
	time.Sleep(2 * time.Second)
	return nil
}

type fastJob struct {
	id string
}

func (j *fastJob) ID() string {
	return j.id
}

func (j *fastJob) Run(context.Context) error {
	return nil
}

type schedulerTestSuite struct {
	suite.Suite
}

// tests schedule a job for every second, and then wait at most a second
// for it to run.  This amount is just slightly larger than 1 second to
// compensate for a few milliseconds of runtime.
const oneSecond = 1*time.Second + 50*time.Millisecond // nolint

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestSchedulerTestSuite(t *testing.T) {
	suite.Run(t, new(schedulerTestSuite))
}

func (s *schedulerTestSuite) TestNewScheduler() {
	// create a new instance of Scheduler
	scheduler := NewJobsScheduler()
	s.Assert().NotNil(scheduler)
}

func (s *schedulerTestSuite) TestStart() {
	// create the text context
	ctx := context.TODO()
	// create a new instance of Scheduler
	scheduler := NewJobsScheduler()
	s.Assert().NotNil(scheduler)
	// start the scheduler
	scheduler.Start(ctx)
	// start the scheduler
	err := scheduler.Stop(ctx)
	s.Assert().NoError(err)
}

func (s *schedulerTestSuite) TestStop() {
	s.Run("with no jobs", func() {
		// create the text context
		ctx := context.TODO() // create a new instance of Scheduler
		scheduler := NewJobsScheduler()
		s.Assert().NotNil(scheduler)
		// start the scheduler
		scheduler.Start(ctx)
		// stop the scheduler
		err := scheduler.Stop(ctx)
		s.Assert().NoError(err)
	})
	s.Run("with scheduler started and stopped then an added job should error", func() {
		var err error
		var scheduler Scheduler
		wg := &sync.WaitGroup{}
		wg.Add(1)
		// create the text context
		ctx := context.TODO()
		// set cron expression and grace period
		const expr = "* * * * * ?"
		// create a new instance of Scheduler
		scheduler = NewJobsScheduler()
		s.Assert().NotNil(scheduler)
		// start the scheduler
		scheduler.Start(ctx)
		s.Assert().NoError(err)

		// stop the scheduler
		err = scheduler.Stop(ctx)
		s.Assert().NoError(err)
		// add a job
		job := &testJob{wg: wg, id: "Job-x"}
		err = scheduler.Schedule(ctx, expr, job)
		s.Assert().Error(err)

		select {
		case <-time.After(oneSecond):
			// No job ran!
		case <-wait(wg):
			s.T().Fatal("expected stopped scheduler does not run any job")
		}
	})
	s.Run("with long running job", func() {
		var err error
		var scheduler Scheduler
		// create the text context
		ctx := context.TODO()
		// set cron expression and grace period
		const expr = "* * * * * *"

		// create a new instance of Scheduler
		scheduler = NewJobsScheduler()
		s.Assert().NotNil(scheduler)

		// start the scheduler
		scheduler.Start(ctx)

		// add a fast job before start
		job := &fastJob{id: "fastJob-X"}
		err = scheduler.Schedule(ctx, expr, job)
		s.Assert().NoError(err)

		// add some jobs
		slowJob := &testLongRunningJob{id: "slowJob"}
		err = scheduler.Schedule(ctx, expr, slowJob)
		s.Assert().NoError(err)

		anotherFastJob := &fastJob{"fastJob-Y"}
		err = scheduler.Schedule(ctx, expr, anotherFastJob)
		s.Assert().NoError(err)
		// sleep for second
		time.Sleep(time.Second)

		// stop the scheduler
		err = scheduler.Stop(ctx)
		s.Assert().NoError(err)
	})
}

func (s *schedulerTestSuite) TestScheduleJob() {
	s.Run("with job scheduled", func() {
		wg := &sync.WaitGroup{}
		wg.Add(1)
		// create the text context
		ctx := context.TODO()
		// set cron expression and grace period
		const expr = "* * * * * ?"
		// create a new instance of Scheduler
		scheduler := NewJobsScheduler()
		s.Assert().NotNil(scheduler)

		// start the scheduler
		scheduler.Start(ctx)

		// add a job
		job := &testJob{wg: wg, id: "Job-X"}
		err := scheduler.Schedule(ctx, expr, job)
		s.Assert().NoError(err)

		select {
		case <-time.After(oneSecond):
			s.T().Fatal("expected job runs")
		case <-wait(wg):
		}

		s.Assert().NoError(scheduler.Stop(ctx))
	})
	s.Run("with duplicate job added and expect only the first job to run", func() {
		wg := &sync.WaitGroup{}
		wg.Add(1)
		// create the text context
		ctx := context.TODO()
		// set cron expression and grace period
		const expr = "* * * * * ?"

		// create a new instance of Scheduler
		scheduler := NewJobsScheduler()
		s.Assert().NotNil(scheduler)

		// start the scheduler
		scheduler.Start(ctx)

		// add a jobs
		job := &testJob{id: "Job-X", wg: wg}
		err := scheduler.Schedule(ctx, expr, job)
		s.Assert().NoError(err)
		err = scheduler.Schedule(ctx, expr, job)
		s.Assert().Error(err)

		select {
		case <-time.After(oneSecond):
			s.T().Fatal("expected job runs")
		case <-wait(wg):
		}

		s.Assert().NoError(scheduler.Stop(ctx))
	})
}

// utility function
func wait(wg *sync.WaitGroup) chan bool {
	ch := make(chan bool)
	go func() {
		wg.Wait()
		ch <- true
	}()
	return ch
}
