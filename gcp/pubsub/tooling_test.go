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

package pubsub

import (
	"context"
	"testing"

	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"github.com/stretchr/testify/assert"
)

// account is a test struct
type account struct {
	AccountID   string
	AccountName string
}

const (
	topicName    = "test-topic"
	projectID    = "test"
	subscriberID = "some-subscriber"
)

func TestCreateTopic(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		ctx := context.TODO()
		emulator := NewEmulator()

		t.Setenv("PUBSUB_EMULATOR_HOST", emulator.EndPoint())

		// create a pubsub client
		client, err := pubsub.NewClient(ctx, projectID)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		// create an instance of the management suite
		mgmt := NewTooling(client)
		assert.NotNil(t, mgmt)

		topic, err := mgmt.CreateTopic(ctx, topicName)
		assert.NoError(t, err)
		assert.NotNil(t, topic)
		assert.IsType(t, &pubsubpb.Topic{}, topic)

		// check that the topic exist
		topics, err := mgmt.ListTopics(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, topics)

		assert.Equal(t, 1, len(topics))
		assert.Equal(t, "projects/test/topics/test-topic", topics[0].GetName())

		err = emulator.Cleanup()
		assert.NoError(t, err)
	})
}
