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

// Config represent the grpc option
type Config struct {
	ServiceName      string // ServiceName is the name given that will show in the traces
	GrpcHost         string // GrpcHost is the gRPC host
	GrpcPort         int32  // GrpcPort is the gRPC port used to received and handle gRPC requests
	TraceEnabled     bool   // TraceEnabled checks whether tracing should be enabled or not
	TraceURL         string // TraceURL is the OTLP collector url.
	EnableReflection bool   // EnableReflection this is useful or local dev testing
	MetricsEnabled   bool   // MetricsEnabled checks whether metrics should be enabled or not
	MetricsPort      int
}
