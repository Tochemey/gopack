/*
 * MIT License
 *
 * Copyright (c) 2022-2024 Arsene Tochemey Gandote
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
	grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
)

// NewMetricUnaryInterceptor returns a grpc metric unary interceptor
func NewMetricUnaryInterceptor() grpc.UnaryServerInterceptor {
	// Create some standard server metrics.
	return grpcPrometheus.UnaryServerInterceptor
}

// NewMetricStreamInterceptor returns a grpc metric stream interceptor
func NewMetricStreamInterceptor() grpc.StreamServerInterceptor {
	return grpcPrometheus.StreamServerInterceptor
}

// NewClientMetricUnaryInterceptor creates a grpc client metric unary interceptor
func NewClientMetricUnaryInterceptor() grpc.UnaryClientInterceptor {
	return grpcPrometheus.UnaryClientInterceptor
}

// NewClientMetricStreamInterceptor creates a grpc client metric stream interceptor
func NewClientMetricStreamInterceptor() grpc.StreamClientInterceptor {
	return grpcPrometheus.StreamClientInterceptor
}
