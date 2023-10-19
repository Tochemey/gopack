/*
 * MIT License
 *
 * Copyright (c) 2022-2023 Tochemey
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
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestFromContext(t *testing.T) {
	// create a request ID
	requestID := uuid.NewString()
	ctx := context.Background()
	// set the context with the newly created request ID
	ctx = context.WithValue(ctx, XRequestIDKey{}, requestID)
	// get the request ID
	actual := FromContext(ctx)
	assert.Equal(t, requestID, actual)
}

func TestContextWithRequestID(t *testing.T) {
	t.Run("context with a request ID", func(t *testing.T) {
		// create a request ID
		requestID := uuid.NewString()
		ctx := context.Background()
		// set the context with the newly created request ID
		ctx = context.WithValue(ctx, XRequestIDKey{}, requestID)
		newCtx := Context(ctx)
		assert.Equal(t, ctx, newCtx)
	})
	t.Run("context without a request ID", func(t *testing.T) {
		ctx := context.Background()
		newCtx := Context(ctx)
		assert.NotEmpty(t, FromContext(newCtx))
		assert.Empty(t, FromContext(ctx))
	})
}
