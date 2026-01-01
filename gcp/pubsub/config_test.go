// MIT License
//
// Copyright (c) 2022-2026 Arsene Tochemey Gandote
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package pubsub

import (
	"context"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/tochemey/gopack/log/zapl"
)

func TestSubscriberConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		ctx := context.TODO()
		emulator := NewEmulator()

		// set the emulator env var
		require.NoError(t, os.Setenv("PUBSUB_EMULATOR_HOST", emulator.EndPoint()))

		// create a pubsub client
		client, err := pubsub.NewClient(ctx, projectID)
		require.NoError(t, err)
		require.NotNil(t, client)

		topic, err := client.TopicAdminClient.CreateTopic(ctx, &pubsubpb.Topic{
			Name: topicName,
		})
		require.NoError(t, err)
		require.NotNil(t, topic)

		subscriptionConfig := &pubsubpb.Subscription{
			Topic: topic.GetName(),
			Name:  subscriberID,
			RetryPolicy: &pubsubpb.RetryPolicy{
				MinimumBackoff: durationpb.New(time.Second * 100),
				MaximumBackoff: durationpb.New(time.Second * 1000),
			},
		}

		receiveSettings := &pubsub.ReceiveSettings{
			MaxExtension:               1,
			MaxDurationPerAckExtension: time.Millisecond * 50,
			MinDurationPerAckExtension: time.Millisecond * 100,
			MaxOutstandingMessages:     1,
			MaxOutstandingBytes:        1048576, // 1mb
			NumGoroutines:              1,
		}

		subscriberConfig := &SubscriberConfig{
			SubscriptionConfig: subscriptionConfig,
			ReceiveSettings:    receiveSettings,
			Logger:             zapl.DiscardLogger,
		}

		err = subscriberConfig.Validate()
		require.NoError(t, err)
		require.NoError(t, emulator.Cleanup())
		require.NoError(t, client.Close())
		err = os.Unsetenv("PUBSUB_EMULATOR_HOST")
		require.NoError(t, err)
	})
	t.Run("subscription config not set", func(t *testing.T) {
		receiveCfg := &pubsub.ReceiveSettings{
			MaxExtension:               1,
			MaxDurationPerAckExtension: time.Millisecond * 50,
			MinDurationPerAckExtension: time.Millisecond * 100,
			MaxOutstandingMessages:     1,
			MaxOutstandingBytes:        1048576, // 1mb
			NumGoroutines:              1,
		}

		subscriberConfig := &SubscriberConfig{
			ReceiveSettings: receiveCfg,
			Logger:          zapl.DiscardLogger,
		}

		err := subscriberConfig.Validate()
		require.Error(t, err)
		assert.EqualError(t, err, "subscription config is not set")
	})
	t.Run("topic not set", func(t *testing.T) {
		subscriptionConfig := &pubsubpb.Subscription{
			Name: subscriberID,
			RetryPolicy: &pubsubpb.RetryPolicy{
				MinimumBackoff: durationpb.New(time.Second * 100),
				MaximumBackoff: durationpb.New(time.Second * 1000),
			},
		}

		receiveSettings := &pubsub.ReceiveSettings{
			MaxExtension:               1,
			MaxDurationPerAckExtension: time.Millisecond * 50,
			MinDurationPerAckExtension: time.Millisecond * 100,
			MaxOutstandingMessages:     1,
			MaxOutstandingBytes:        1048576, // 1mb
			NumGoroutines:              1,
		}

		subscriberConfig := &SubscriberConfig{
			ReceiveSettings:    receiveSettings,
			SubscriptionConfig: subscriptionConfig,
			Logger:             zapl.DiscardLogger,
		}

		err := subscriberConfig.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "subscription topic is not set")
	})
	t.Run("subscription ID not set", func(t *testing.T) {
		ctx := context.TODO()
		emulator := NewEmulator()

		// set the emulator env var
		assert.NoError(t, os.Setenv("PUBSUB_EMULATOR_HOST", emulator.EndPoint()))

		// create a pubsub client
		client, err := pubsub.NewClient(ctx, projectID)
		assert.NoError(t, err)
		assert.NotNil(t, client)

		topic, err := client.TopicAdminClient.CreateTopic(ctx, &pubsubpb.Topic{
			Name: topicName,
		})
		require.NoError(t, err)
		require.NotNil(t, topic)

		subscriptionConfig := &pubsubpb.Subscription{
			Topic: topic.GetName(),
			RetryPolicy: &pubsubpb.RetryPolicy{
				MinimumBackoff: durationpb.New(time.Second * 100),
				MaximumBackoff: durationpb.New(time.Second * 1000),
			},
		}

		receiveSettings := &pubsub.ReceiveSettings{
			MaxExtension:               1,
			MaxDurationPerAckExtension: time.Millisecond * 50,
			MinDurationPerAckExtension: time.Millisecond * 100,
			MaxOutstandingMessages:     1,
			MaxOutstandingBytes:        1048576, // 1mb
			NumGoroutines:              1,
		}

		subscriberConfig := &SubscriberConfig{
			SubscriptionConfig: subscriptionConfig,
			ReceiveSettings:    receiveSettings,
			Logger:             zapl.DiscardLogger,
		}

		err = subscriberConfig.Validate()
		assert.Error(t, err)
		assert.NoError(t, emulator.Cleanup())
		assert.NoError(t, client.Close())
		err = os.Unsetenv("PUBSUB_EMULATOR_HOST")
		assert.NoError(t, err)
	})
}
