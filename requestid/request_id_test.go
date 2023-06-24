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
