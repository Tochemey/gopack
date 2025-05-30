/*
 * MIT License
 *
 * Copyright (c) 2022-2025  Arsene Tochemey Gandote
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

package future

import (
	"context"
	"sync"
)

// Future represents a value which may or may not currently be available,
// but will be available at some point in the future, or an error if that value
// could not be made available. It provides a way to handle asynchronous
// computations and their results.
//
// The Future interface provides two main methods:
//
// 1. Await(ctx context.Context) (proto.Message, error):
//   - This method blocks until the Future is completed or the provided context
//     is canceled. It returns either the result of the computation or an error
//     if the computation failed or the context was canceled.
//
// 2. complete(value proto.Message, err error):
//   - This method completes the Future with either a value or an error. It is
//     used internally by the completable to set the result of the computation.
//
// Example usage:
//
//	task := func() (proto.Message, error) {
//	    // Perform some long-running computation
//	    result := &MyProtoMessage{...}
//	    return result, nil
//	}
//
//	future := future.New(task)
//
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
//	defer cancel()
//
//	result, err := future.Await(ctx)
//	if err != nil {
//	    log.Fatalf("Failed to get result: %v", err)
//	}
//
//	log.Printf("Received result: %v", result)
type Future[T any] interface {
	// Await blocks until the Future is completed or context is canceled and
	// returns either a result or an error.
	Await(context.Context) (T, error)

	// complete completes the Future with either a value or an error.
	// It is used by [completable] internally.
	complete(T, error)
}

// New creates a new Future that executes the given long-running task.
// The task is a function that returns a proto.Message and an error.
// The Future is completed with the value returned by the task or failed with the error.
//
// The task is executed asynchronously in a separate goroutine. The Future can be
// awaited using the Await method, which will block until the task is completed
// or the provided context is canceled.
//
// Example usage:
//
//	task := func() (proto.Message, error) {
//	    // Perform some long-running computation
//	    result := &MyProtoMessage{...}
//	    return result, nil
//	}
//
//	future := future.New(task)
//
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
//	defer cancel()
//
//	result, err := future.Await(ctx)
//	if err != nil {
//	    log.Fatalf("Failed to get result: %v", err)
//	}
//
//	log.Printf("Received result: %v", result)
func New[T any](task func() (T, error)) Future[T] {
	comp := newCompletable[T]()
	go func() {
		result, err := task()
		switch {
		case err == nil:
			comp.Success(result)
		default:
			comp.Failure(err)
		}
	}()
	return comp.Future()
}

// future implements the Future interface.
type future[T any] struct {
	acceptOnce   sync.Once
	completeOnce sync.Once
	done         chan any
	value        T
	err          error
}

// Verify future satisfies the Future interface.
var _ Future[any] = (*future[any])(nil)

// newFuture returns a new Future.
func newFuture[T any]() Future[T] {
	return &future[T]{
		done: make(chan any, 1),
	}
}

// wait blocks once, until the Future result is available or until
// the context is canceled.
func (x *future[T]) wait(ctx context.Context) {
	x.acceptOnce.Do(func() {
		select {
		case result := <-x.done:
			x.setResult(result)
		case <-ctx.Done():
			x.setResult(ctx.Err())
		}
	})
}

// setResult assigns a value to the Future instance.
func (x *future[T]) setResult(result any) {
	switch value := result.(type) {
	case error:
		x.err = value
	default:
		x.value = value.(T)
	}
}

// Await blocks until the Future is completed or context is canceled and
// returns either a result or an error.
func (x *future[T]) Await(ctx context.Context) (T, error) {
	x.wait(ctx)
	return x.value, x.err
}

// complete completes the Future with either a value or an error.
func (x *future[T]) complete(value T, err error) {
	x.completeOnce.Do(func() {
		if err != nil {
			x.done <- err
		} else {
			x.done <- value
		}
	})
}

// completable represents a writable, single-assignment container,
// which completes a Future.
type completable[T any] interface {
	// Success completes the underlying Future with a value.
	Success(T)

	// Failure fails the underlying Future with an error.
	Failure(error)

	// Future returns the underlying Future.
	Future() Future[T]
}

// completer implements the completable interface.
type completer[T any] struct {
	once   sync.Once
	future Future[T]
}

// Verify completer satisfies the completable interface.
var _ completable[any] = (*completer[any])(nil)

// newCompletable returns a new completable.
func newCompletable[T any]() completable[T] {
	return &completer[T]{
		future: newFuture[T](),
	}
}

// Success completes the underlying Future with a given value.
func (p *completer[T]) Success(value T) {
	p.once.Do(func() {
		p.future.complete(value, nil)
	})
}

// Failure fails the underlying Future with a given error.
func (p *completer[T]) Failure(err error) {
	p.once.Do(func() {
		var zero T
		p.future.complete(zero, err)
	})
}

// Future returns the underlying Future.
func (p *completer[T]) Future() Future[T] {
	return p.future
}
