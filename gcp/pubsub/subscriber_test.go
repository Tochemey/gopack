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
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/stretchr/testify/assert"
	pubpb "google.golang.org/genproto/googleapis/pubsub/v1"

	"github.com/tochemey/gopack/log/zapl"
)

func TestNewSubscriberClient(t *testing.T) {
	t.Run("successful with non existing subscription", func(t *testing.T) {
		// create the go context
		ctx := context.TODO()
		emulator := NewEmulator()

		// set the emulator addr
		assert.NoError(t, os.Setenv("PUBSUB_EMULATOR_HOST", emulator.EndPoint()))

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

		// create the topic to use
		topic := client.Topic(topicName)

		// let us start consuming the messages
		subCfg := &pubsub.SubscriptionConfig{
			Topic: topic,
			RetryPolicy: &pubsub.RetryPolicy{
				MinimumBackoff: time.Second * 100,
				MaximumBackoff: time.Second * 1000,
			},
		}

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
			SubscriptionID:     subscriberID,
			SubscriptionConfig: subCfg,
			ReceiveSettings:    receiveCfg,
			Logger:             zapl.DiscardLogger,
		}
		// create an instance of the subscriber
		subClient, err := NewSubscriber(ctx, client, subscriberConfig)
		assert.NotNil(t, subClient)
		assert.NoError(t, err)

		// fetch the list of subscriber
		resp, err := emulator.Server().GServer.ListSubscriptions(ctx, &pubpb.ListSubscriptionsRequest{
			Project:   fmt.Sprintf("projects/%s", projectID),
			PageSize:  10,
			PageToken: "",
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 1, len(resp.GetSubscriptions()))
		actualSub := resp.GetSubscriptions()[0]
		expected := fmt.Sprintf("projects/%s/subscriptions/%s", projectID, subscriberID)
		assert.Equal(t, expected, actualSub.GetName())
		// asserts values were set
		assert.EqualValues(t, 1000, actualSub.GetRetryPolicy().GetMaximumBackoff().Seconds)
		assert.EqualValues(t, 100, actualSub.GetRetryPolicy().GetMinimumBackoff().Seconds)

		// cleanup resources
		assert.NoError(t, emulator.Cleanup())
		assert.NoError(t, client.Close())
		err = os.Unsetenv("PUBSUB_EMULATOR_HOST")
		assert.NoError(t, err)
	})
	t.Run("successful with existing subscription", func(t *testing.T) {
		// create the go context
		ctx := context.TODO()
		emulator := NewEmulator()

		// set the emulator addr
		assert.NoError(t, os.Setenv("PUBSUB_EMULATOR_HOST", emulator.EndPoint()))
		client, err := pubsub.NewClient(ctx, projectID)
		assert.NoError(t, err)
		assert.NotNil(t, client)

		// create an instance of the management suite
		mgmt := NewTooling(client)
		assert.NotNil(t, mgmt)

		// create the topic using the management API
		_, err = mgmt.CreateTopic(ctx, topicName)
		assert.NoError(t, err)

		// create the topic to use
		topic := client.Topic(topicName)

		// let us start consuming the messages
		subCfg := &pubsub.SubscriptionConfig{
			Topic: topic,
			RetryPolicy: &pubsub.RetryPolicy{
				MinimumBackoff: time.Second * 100,
				MaximumBackoff: time.Second * 1000,
			},
			AckDeadline: 0,
		}

		receiveCfg := &pubsub.ReceiveSettings{
			MaxExtension:           1,
			MaxExtensionPeriod:     time.Millisecond * 50,
			MinExtensionPeriod:     time.Millisecond * 100,
			MaxOutstandingMessages: 1,
			MaxOutstandingBytes:    1048576, // 1mb
			UseLegacyFlowControl:   false,
			NumGoroutines:          1,
		}

		// creates the initial subscription
		_, err = client.CreateSubscription(ctx, subscriberID, pubsub.SubscriptionConfig{
			Topic: topic,
		})
		assert.NoError(t, err)

		// create an instance of the subscriber
		subClient, err := NewSubscriber(ctx, client, &SubscriberConfig{
			SubscriptionID:     subscriberID,
			SubscriptionConfig: subCfg,
			ReceiveSettings:    receiveCfg,
			Logger:             zapl.DiscardLogger,
		})
		assert.NotNil(t, subClient)
		assert.NoError(t, err)

		// fetch the list of subscriber
		resp, err := emulator.Server().GServer.ListSubscriptions(ctx, &pubpb.ListSubscriptionsRequest{
			Project:   fmt.Sprintf("projects/%s", projectID),
			PageSize:  1,
			PageToken: "",
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 1, len(resp.GetSubscriptions()))
		actualSub := resp.GetSubscriptions()[0]
		expected := fmt.Sprintf("projects/%s/subscriptions/%s", projectID, subscriberID)
		assert.Equal(t, expected, actualSub.GetName())
		// asserts values were updated
		assert.EqualValues(t, 1000, actualSub.GetRetryPolicy().GetMaximumBackoff().Seconds)
		assert.EqualValues(t, 100, actualSub.GetRetryPolicy().GetMinimumBackoff().Seconds)

		// cleanup resources
		assert.NoError(t, emulator.Cleanup())
		assert.NoError(t, client.Close())
		err = os.Unsetenv("PUBSUB_EMULATOR_HOST")
		assert.NoError(t, err)
	})
	t.Run("topic does not exist", func(t *testing.T) {
		// create the go context
		ctx := context.TODO()
		emulator := NewEmulator()

		// set the emulator addr
		assert.NoError(t, os.Setenv("PUBSUB_EMULATOR_HOST", emulator.EndPoint()))

		client, err := pubsub.NewClient(ctx, projectID)
		assert.NoError(t, err)
		assert.NotNil(t, client)

		// create an instance of the management suite
		mgmt := NewTooling(client)
		assert.NotNil(t, mgmt)

		// create the topic to use
		topic := client.Topic(topicName)

		// let us start consuming the messages
		subCfg := &pubsub.SubscriptionConfig{
			Topic: topic,
			RetryPolicy: &pubsub.RetryPolicy{
				MinimumBackoff: time.Second * 100,
				MaximumBackoff: time.Second * 1000,
			},
		}

		receiveCfg := &pubsub.ReceiveSettings{
			MaxExtension:           1,
			MaxExtensionPeriod:     time.Millisecond * 50,
			MinExtensionPeriod:     time.Millisecond * 100,
			MaxOutstandingMessages: 1,
			MaxOutstandingBytes:    1048576, // 1mb
			UseLegacyFlowControl:   false,
			NumGoroutines:          1,
		}

		// create an instance of the subscriber
		subClient, err := NewSubscriber(ctx, client, &SubscriberConfig{
			SubscriptionID:     subscriberID,
			SubscriptionConfig: subCfg,
			ReceiveSettings:    receiveCfg,
			Logger:             zapl.DiscardLogger,
		})
		assert.Nil(t, subClient)
		assert.EqualError(t, err, "topic test-topic does not exist")

		// cleanup resources
		assert.NoError(t, emulator.Cleanup())
		assert.NoError(t, client.Close())
		err = os.Unsetenv("PUBSUB_EMULATOR_HOST")
		assert.NoError(t, err)
	})
	t.Run("with config not set", func(t *testing.T) {
		ctx := context.TODO()
		emulator := NewEmulator()

		// set the emulator addr
		assert.NoError(t, os.Setenv("PUBSUB_EMULATOR_HOST", emulator.EndPoint()))

		projectID := "test"
		// create a pubsub client
		client, err := pubsub.NewClient(ctx, projectID)
		assert.NoError(t, err)
		assert.NotNil(t, client)

		// create an instance of the subscriber
		subClient, err := NewSubscriber(ctx, client, nil)
		assert.Nil(t, subClient)
		assert.Error(t, err)
		assert.EqualError(t, err, "config is not set")

		assert.NoError(t, emulator.Cleanup())
		assert.NoError(t, client.Close())
		err = os.Unsetenv("PUBSUB_EMULATOR_HOST")
		assert.NoError(t, err)
	})
	t.Run("with invalid config", func(t *testing.T) {
		ctx := context.TODO()
		emulator := NewEmulator()

		// set the emulator addr
		assert.NoError(t, os.Setenv("PUBSUB_EMULATOR_HOST", emulator.EndPoint()))

		projectID := "test"
		// create a pubsub client
		client, err := pubsub.NewClient(ctx, projectID)
		assert.NoError(t, err)
		assert.NotNil(t, client)

		subscriberID := "some-subscription"
		subscriberConfig := &SubscriberConfig{
			SubscriptionID: subscriberID,
			Logger:         zapl.DiscardLogger,
		}
		// create an instance of the subscriber
		subClient, err := NewSubscriber(ctx, client, subscriberConfig)
		assert.Nil(t, subClient)
		assert.Error(t, err)

		assert.NoError(t, emulator.Cleanup())
		assert.NoError(t, client.Close())
		err = os.Unsetenv("PUBSUB_EMULATOR_HOST")
		assert.NoError(t, err)
	})
}

func TestConsume(t *testing.T) {
	t.Run("consume a published message with successful handler", func(t *testing.T) {
		// create the go context
		ctx := context.TODO()
		emulator := NewEmulator()

		// set the emulator addr
		assert.NoError(t, os.Setenv("PUBSUB_EMULATOR_HOST", emulator.EndPoint()))

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
		// create the topic to use
		topic := client.Topic(topicName)

		// let us start consuming the messages
		subCfg := &pubsub.SubscriptionConfig{Topic: topic}

		// create an instance of the subscriber
		subClient, err := NewSubscriber(ctx, client, &SubscriberConfig{
			SubscriptionID:     subscriberID,
			SubscriptionConfig: subCfg,
			Logger:             zapl.DiscardLogger,
		})
		assert.NotNil(t, subClient)
		assert.NoError(t, err)

		// fetch the list of subscriber
		resp, err := emulator.Server().GServer.ListSubscriptions(ctx, &pubpb.ListSubscriptionsRequest{
			Project:   fmt.Sprintf("projects/%s", projectID),
			PageSize:  10,
			PageToken: "",
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 1, len(resp.GetSubscriptions()))
		actualSub := resp.GetSubscriptions()[0]
		expected := fmt.Sprintf("projects/%s/subscriptions/%s", projectID, subscriberID)
		assert.Equal(t, expected, actualSub.GetName())

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

		publicationTopic := &Topic{Name: topicName, EnableOrdering: true}
		// let us publish the message
		err = pub.Publish(ctx, publicationTopic, []*Message{message, message, message})
		assert.NoError(t, err)

		// consume some messages for 2 seconds
		cancelCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		var actualAcct *account
		mu := sync.Mutex{}
		decode := func(data []byte) error {
			mu.Lock()
			defer mu.Unlock()
			err := json.Unmarshal(data, &actualAcct)
			if err != nil {
				return err
			}
			return nil
		}
		// create a handler
		handler := func(ctx context.Context, data []byte) error {
			return decode(data)
		}
		errChan := make(chan error, 1)
		go subClient.Consume(cancelCtx, handler, errChan)
		for e := range errChan {
			assert.NoError(t, e)
		}

		actualAccountName := actualAcct.AccountName
		assert.Equal(t, acct.AccountName, actualAccountName)

		// cleanup resources
		assert.NoError(t, emulator.Cleanup())
		assert.NoError(t, client.Close())
		err = os.Unsetenv("PUBSUB_EMULATOR_HOST")
		assert.NoError(t, err)
	})
	t.Run("consume a published message with failure handler", func(t *testing.T) {
		// create the go context
		ctx := context.TODO()
		emulator := NewEmulator()

		// set the emulator env var
		assert.NoError(t, os.Setenv("PUBSUB_EMULATOR_HOST", emulator.EndPoint()))

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

		// create the topic to use
		topic := client.Topic(topicName)

		// let us start consuming the messages
		subCfg := &pubsub.SubscriptionConfig{Topic: topic}

		// create an instance of the subscriber
		subClient, err := NewSubscriber(ctx, client, &SubscriberConfig{
			SubscriptionID:     subscriberID,
			SubscriptionConfig: subCfg,
			Logger:             zapl.DiscardLogger,
		})
		assert.NotNil(t, subClient)
		assert.NoError(t, err)

		// fetch the list of subscriber
		resp, err := emulator.Server().GServer.ListSubscriptions(ctx, &pubpb.ListSubscriptionsRequest{
			Project:   fmt.Sprintf("projects/%s", projectID),
			PageSize:  10,
			PageToken: "",
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 1, len(resp.GetSubscriptions()))
		actualSub := resp.GetSubscriptions()[0]
		expected := fmt.Sprintf("projects/%s/subscriptions/%s", projectID, subscriberID)
		assert.Equal(t, expected, actualSub.GetName())

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

		pubTopic := &Topic{Name: topicName, EnableOrdering: true}
		// let us publish the message
		err = pub.Publish(ctx, pubTopic, []*Message{message})
		assert.NoError(t, err)

		// consume some messages for 2 seconds
		cancelCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		var count int32
		handler := func(context.Context, []byte) error {
			atomic.AddInt32(&count, 1)
			// successful when even iteration
			if count%2 == 0 {
				return nil
			}
			return errors.New("failure")
		}
		errChan := make(chan error, 10)
		go subClient.Consume(cancelCtx, handler, errChan)
		for e := range errChan {
			assert.Error(t, e)
		}

		// cleanup resources
		assert.NoError(t, emulator.Cleanup())
		assert.NoError(t, client.Close())
		err = os.Unsetenv("PUBSUB_EMULATOR_HOST")
		assert.NoError(t, err)
	})
	t.Run("consume many messages", func(t *testing.T) {
		// create the go context
		ctx := context.TODO()
		emulator := NewEmulator()

		// set the emulator addr
		// set the emulator env var
		assert.NoError(t, os.Setenv("PUBSUB_EMULATOR_HOST", emulator.EndPoint()))

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

		// create an instance of the subscriber
		subscriber, err := NewSubscriberWithDefaults(ctx, client, subscriberID, topicName)
		assert.NotNil(t, subscriber)
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

		// let us create one thousand messages to publish
		messages := make([]*Message, 1000)
		for i := 0; i < 1000; i++ {
			messages[i] = &Message{
				Key:     "some-key",
				Payload: bytea,
			}
		}

		pubTopic := &Topic{Name: topicName, EnableOrdering: true}
		// let us publish the message
		err = pub.Publish(ctx, pubTopic, messages)
		assert.NoError(t, err)

		// consume some messages for 2 seconds
		cancelCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		var count int32
		handler := func(context.Context, []byte) error {
			atomic.AddInt32(&count, 1)
			return nil
		}
		errChan := make(chan error, 1000)
		go subscriber.Consume(cancelCtx, handler, errChan)
		for e := range errChan {
			assert.NoError(t, e)
		}

		assert.EqualValues(t, subscriber.messagesProcessedCount, count)
		// cleanup resources
		assert.NoError(t, emulator.Cleanup())
		assert.NoError(t, client.Close())
		err = os.Unsetenv("PUBSUB_EMULATOR_HOST")
		assert.NoError(t, err)
	})
}

func Test_subscriptionConfigToUpdate(t *testing.T) {
	ackDeadline := time.Millisecond * 1
	retentionDuration := time.Millisecond * 2
	expirationPolicy := time.Millisecond * 3
	deadLetterPolicy := &pubsub.DeadLetterPolicy{
		DeadLetterTopic:     "some-topic",
		MaxDeliveryAttempts: 1,
	}
	labels := map[string]string{"baz": "bar"}
	retryPolicy := &pubsub.RetryPolicy{
		MinimumBackoff: time.Millisecond * 4,
		MaximumBackoff: time.Millisecond * 5,
	}

	t.Run("happy path", func(t *testing.T) {
		pushConfig := &pubsub.PushConfig{
			Endpoint:             "some-endpoint",
			Attributes:           map[string]string{"foo": "bar"},
			AuthenticationMethod: &pubsub.OIDCToken{},
		}

		bigQueryConfig := &pubsub.BigQueryConfig{
			Table:             "some-table",
			UseTopicSchema:    true,
			WriteMetadata:     true,
			DropUnknownFields: true,
			State:             pubsub.BigQueryConfigActive,
		}

		input := &pubsub.SubscriptionConfig{
			PushConfig:                    *pushConfig,
			BigQueryConfig:                *bigQueryConfig,
			AckDeadline:                   ackDeadline,
			RetainAckedMessages:           true,
			RetentionDuration:             retentionDuration,
			ExpirationPolicy:              expirationPolicy,
			Labels:                        labels,
			DeadLetterPolicy:              deadLetterPolicy,
			RetryPolicy:                   retryPolicy,
			TopicMessageRetentionDuration: retentionDuration,
		}

		expected := pubsub.SubscriptionConfigToUpdate{
			PushConfig:          pushConfig,
			BigQueryConfig:      bigQueryConfig,
			AckDeadline:         ackDeadline,
			RetainAckedMessages: true,
			RetentionDuration:   retentionDuration,
			ExpirationPolicy:    expirationPolicy,
			DeadLetterPolicy:    deadLetterPolicy,
			Labels:              labels,
			RetryPolicy:         retryPolicy,
		}

		actual := subscriptionConfigToUpdate(input)
		assert.Equal(t, expected, actual)
	})
	t.Run("without push config or big query config", func(t *testing.T) {
		pushConfig := &pubsub.PushConfig{}
		bigQueryConfig := &pubsub.BigQueryConfig{}
		input := &pubsub.SubscriptionConfig{
			PushConfig:                    *pushConfig,
			BigQueryConfig:                *bigQueryConfig,
			AckDeadline:                   ackDeadline,
			RetainAckedMessages:           true,
			RetentionDuration:             retentionDuration,
			ExpirationPolicy:              expirationPolicy,
			Labels:                        labels,
			DeadLetterPolicy:              deadLetterPolicy,
			RetryPolicy:                   retryPolicy,
			TopicMessageRetentionDuration: retentionDuration,
		}

		expected := pubsub.SubscriptionConfigToUpdate{
			PushConfig:          nil,
			BigQueryConfig:      nil,
			AckDeadline:         ackDeadline,
			RetainAckedMessages: true,
			RetentionDuration:   retentionDuration,
			ExpirationPolicy:    expirationPolicy,
			DeadLetterPolicy:    deadLetterPolicy,
			Labels:              labels,
			RetryPolicy:         retryPolicy,
		}

		actual := subscriptionConfigToUpdate(input)
		assert.Equal(t, expected, actual)
	})
}
