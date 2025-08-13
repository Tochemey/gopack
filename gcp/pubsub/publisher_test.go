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
	"encoding/json"
	"testing"

	"cloud.google.com/go/pubsub/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"

	"github.com/tochemey/gopack/log/zapl"
)

func TestPublish(t *testing.T) {
	t.Run("successful when topic exist", func(t *testing.T) {
		// create the go context
		ctx := context.TODO()
		// set the emulator addr
		emulator := NewEmulator()

		// set the emulator env var
		t.Setenv("PUBSUB_EMULATOR_HOST", emulator.EndPoint())

		// create a pubsub client
		client, err := pubsub.NewClient(ctx, projectID)
		assert.NoError(t, err)
		assert.NotNil(t, client)

		// create an instance of the management suite
		mgmt := NewTooling(client)
		assert.NotNil(t, mgmt)

		// create the topic using the management API
		_, err = mgmt.CreateTopic(ctx, topicName)
		assert.NoError(t, err)

		// create an instance of the publisher
		pub := NewPublisher(client, zapl.DiscardLogger)
		assert.NotNil(t, pub)
		assert.NoError(t, err)

		// create an object to persist
		acct := &account{
			AccountID:   "test-account-id",
			AccountName: "test-account-name",
		}
		// let us jsonify the account
		bytea, err := json.Marshal(acct)
		assert.NoError(t, err)
		assert.NotNil(t, bytea)

		// create the message
		message := &Message{
			Key:     "some-key",
			Payload: bytea,
		}

		// create a topic
		topic := &Topic{
			Name:           topicName,
			EnableOrdering: true,
		}
		// let us publish the message
		err = pub.Publish(ctx, topic, []*Message{message})
		assert.NoError(t, err)

		// check that the message is actually published
		publishedMessages := emulator.Server().Messages()
		assert.NotEmpty(t, publishedMessages)
		assert.Equal(t, 1, len(publishedMessages))

		actual := publishedMessages[0]
		// unmarshal the actual payload
		var actualAcct *account
		err = json.Unmarshal(actual.Data, &actualAcct)
		assert.NoError(t, err)

		assert.True(t, cmp.Equal(acct, actualAcct))

		// cleanup resources
		assert.NoError(t, emulator.Cleanup())
		assert.NoError(t, client.Close())
	})
	t.Run("fails when the ordering key is not set with ordering turned on", func(t *testing.T) {
		// create the go context
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

		// create the topic using the management API
		_, err = mgmt.CreateTopic(ctx, topicName)
		assert.NoError(t, err)

		// create an instance of the publisher
		pub := NewPublisher(client, zapl.DiscardLogger)
		assert.NotNil(t, pub)
		assert.NoError(t, err)

		// create an object to persist
		acct := &account{
			AccountID:   "test-account-id",
			AccountName: "test-account-name",
		}
		// let us jsonify the account
		bytea, err := json.Marshal(acct)
		assert.NoError(t, err)
		assert.NotNil(t, bytea)

		// create the message
		message := &Message{
			Key:     "",
			Payload: bytea,
		}

		// create a topic
		topic := &Topic{
			Name:           topicName,
			EnableOrdering: true,
		}

		// let us publish the message
		err = pub.Publish(ctx, topic, []*Message{message})
		assert.EqualError(t, err, "message key is required when MessageOrdering is enabled")

		// cleanup resources
		assert.NoError(t, emulator.Cleanup())
		assert.NoError(t, client.Close())
	})
	t.Run("fails when the topic does not exist", func(t *testing.T) {
		// create the go context
		ctx := context.TODO()
		// set the emulator addr
		emulator := NewEmulator()

		t.Setenv("PUBSUB_EMULATOR_HOST", emulator.EndPoint())

		// create a pubsub client
		client, err := pubsub.NewClient(ctx, projectID)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		// create an instance of the management suite
		mgmt := NewTooling(client)
		assert.NotNil(t, mgmt)

		// create an instance of the publisher
		pub := NewPublisher(client, zapl.DiscardLogger)
		assert.NotNil(t, pub)
		assert.NoError(t, err)

		// create an object to persist
		acct := &account{
			AccountID:   "test-account-id",
			AccountName: "test-account-name",
		}
		// let us jsonify the account
		bytea, err := json.Marshal(acct)
		assert.NoError(t, err)
		assert.NotNil(t, bytea)

		// create the message
		message := &Message{
			Key:     "",
			Payload: bytea,
		}

		// let us publish the message
		topic := &Topic{
			Name: topicName,
		}

		err = pub.Publish(ctx, topic, []*Message{message})
		assert.EqualError(t, err, "unable to publish message to GCP Pub/Sub: rpc error: code = NotFound desc = topic \"projects/test/topics/test-topic\"")
		// cleanup resources
		assert.NoError(t, emulator.Cleanup())
		assert.NoError(t, client.Close())
	})
}
