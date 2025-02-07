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
	"os"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/stretchr/testify/assert"

	"github.com/tochemey/gopack/log/zapl"
)

func TestSubscriberConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		ctx := context.TODO()
		emulator := NewEmulator()

		// set the emulator env var
		assert.NoError(t, os.Setenv("PUBSUB_EMULATOR_HOST", emulator.EndPoint()))

		// create a pubsub client
		client, err := pubsub.NewClient(ctx, projectID)
		assert.NoError(t, err)
		assert.NotNil(t, client)

		topic := client.Topic(topicName)
		subscriptionConfig := &pubsub.SubscriptionConfig{
			Topic: topic,
			RetryPolicy: &pubsub.RetryPolicy{
				MinimumBackoff: time.Second * 100,
				MaximumBackoff: time.Second * 1000,
			},
		}

		receiveSettings := &pubsub.ReceiveSettings{
			MaxExtension:           1,
			MaxExtensionPeriod:     time.Millisecond * 50,
			MinExtensionPeriod:     time.Millisecond * 100,
			MaxOutstandingMessages: 1,
			MaxOutstandingBytes:    1048576, // 1mb
			UseLegacyFlowControl:   false,
			NumGoroutines:          1,
		}

		subscriberConfig := &SubscriberConfig{
			SubscriptionID:     subscriberID,
			SubscriptionConfig: subscriptionConfig,
			ReceiveSettings:    receiveSettings,
			Logger:             zapl.DiscardLogger,
		}

		err = subscriberConfig.Validate()
		assert.NoError(t, err)
		assert.NoError(t, emulator.Cleanup())
		assert.NoError(t, client.Close())
		err = os.Unsetenv("PUBSUB_EMULATOR_HOST")
		assert.NoError(t, err)
	})
	t.Run("subscription config not set", func(t *testing.T) {
		receiveCfg := &pubsub.ReceiveSettings{
			MaxExtension:           1,
			MaxExtensionPeriod:     time.Millisecond * 50,
			MinExtensionPeriod:     time.Millisecond * 100,
			MaxOutstandingMessages: 1,
			MaxOutstandingBytes:    1048576, // 1mb
			UseLegacyFlowControl:   false,
			NumGoroutines:          1,
		}

		subscriberConfig := &SubscriberConfig{
			SubscriptionID:  subscriberID,
			ReceiveSettings: receiveCfg,
			Logger:          zapl.DiscardLogger,
		}

		err := subscriberConfig.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "subscription config is not set")
	})
	t.Run("topic not set", func(t *testing.T) {
		subscriptionConfig := &pubsub.SubscriptionConfig{
			RetryPolicy: &pubsub.RetryPolicy{
				MinimumBackoff: time.Second * 100,
				MaximumBackoff: time.Second * 1000,
			},
		}

		receiveSettings := &pubsub.ReceiveSettings{
			MaxExtension:           1,
			MaxExtensionPeriod:     time.Millisecond * 50,
			MinExtensionPeriod:     time.Millisecond * 100,
			MaxOutstandingMessages: 1,
			MaxOutstandingBytes:    1048576, // 1mb
			UseLegacyFlowControl:   false,
			NumGoroutines:          1,
		}

		subscriberConfig := &SubscriberConfig{
			SubscriptionID:     subscriberID,
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

		topic := client.Topic(topicName)
		subscriptionConfig := &pubsub.SubscriptionConfig{
			Topic: topic,
			RetryPolicy: &pubsub.RetryPolicy{
				MinimumBackoff: time.Second * 100,
				MaximumBackoff: time.Second * 1000,
			},
		}

		receiveSettings := &pubsub.ReceiveSettings{
			MaxExtension:           1,
			MaxExtensionPeriod:     time.Millisecond * 50,
			MinExtensionPeriod:     time.Millisecond * 100,
			MaxOutstandingMessages: 1,
			MaxOutstandingBytes:    1048576, // 1mb
			UseLegacyFlowControl:   false,
			NumGoroutines:          1,
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
