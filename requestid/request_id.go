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

package requestid

import (
	"context"

	"github.com/google/uuid"
)

// XRequestIDKey is used to store the x-request-id
// into the grpc context
type XRequestIDKey struct{}

const (
	XRequestIDMetadataKey = "x-request-id"
)

// FromContext return the request ID set in context
func FromContext(ctx context.Context) string {
	id, ok := ctx.Value(XRequestIDKey{}).(string)
	if !ok {
		return ""
	}
	return id
}

// Context sets a requestID into the parent context and return the new
// context that can be used in case there is no request ID. In case the parent context contains a requestID then it is returned
// as the newly created context, otherwise a new context is created with a requestID
func Context(ctx context.Context) context.Context {
	// here the given context contains a request ID no need to set one
	// just return the context
	if requestID := FromContext(ctx); requestID != "" {
		return ctx
	}

	// create a requestID
	requestID := uuid.NewString()
	// set the requestID and return the new context
	return context.WithValue(ctx, XRequestIDKey{}, requestID)
}
