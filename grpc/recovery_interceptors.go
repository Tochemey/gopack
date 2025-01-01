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

package grpc

import (
	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NewRecoveryUnaryInterceptor recovers from an unexpected panic
// Recovery handlers should typically be last in the chain so that other middleware
// (e.g. logging) can operate on the recovered state instead of being directly affected by any panic
func NewRecoveryUnaryInterceptor() grpc.UnaryServerInterceptor {
	// Define custom func to handle panic
	customFunc := func(p interface{}) (err error) {
		return status.Errorf(codes.Unknown, "panic triggered: %v", p)
	}

	opts := []grpcRecovery.Option{
		grpcRecovery.WithRecoveryHandler(customFunc),
	}

	return grpcRecovery.UnaryServerInterceptor(opts...)
}

// NewRecoveryStreamInterceptor recovers from an unexpected panic
// Recovery handlers should typically be last in the chain so that other middleware
// (e.g. logging) can operate on the recovered state instead of being directly affected by any panic
func NewRecoveryStreamInterceptor() grpc.StreamServerInterceptor {
	// Define custom func to handle panic
	customFunc := func(p interface{}) (err error) {
		return status.Errorf(codes.Unknown, "panic triggered: %v", p)
	}

	opts := []grpcRecovery.Option{
		grpcRecovery.WithRecoveryHandler(customFunc),
	}

	return grpcRecovery.StreamServerInterceptor(opts...)
}
