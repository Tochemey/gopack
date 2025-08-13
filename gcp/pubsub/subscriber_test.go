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
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/tochemey/gopack/log/zapl"
)

func TestNewSubscriberClient(t *testing.T) {
	t.Run("successful with non existing subscription", func(t *testing.T) {
		// create the go context
		ctx := context.TODO()
		emulator := NewEmulator()

		// set the emulator addr
		t.Setenv("PUBSUB_EMULATOR_HOST", emulator.EndPoint())

		// create a pubsub client
		client, err := pubsub.NewClient(ctx, projectID)
		require.NoError(t, err)
		require.NotNil(t, client)

		// create an instance of the management suite
		mgmt := NewTooling(client)
		require.NotNil(t, mgmt)

		// create the topic using the management API
		_, err = mgmt.CreateTopic(ctx, topicName)
		require.NoError(t, err)

		// let us start consuming the messages
		subCfg := &pubsubpb.Subscription{
			Name:  SubscriptionFullName(projectID, subscriberID),
			Topic: TopicFullName(projectID, topicName),
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
			SubscriptionConfig: subCfg,
			ReceiveSettings:    receiveSettings,
			Logger:             zapl.DiscardLogger,
		}
		// create an instance of the subscriber
		subscriber, err := NewSubscriber(ctx, client, subscriberConfig)
		require.NotNil(t, subscriber)
		require.NoError(t, err)

		// fetch the list of subscriber
		resp, err := emulator.Server().GServer.ListSubscriptions(ctx, &pubsubpb.ListSubscriptionsRequest{
			Project:   fmt.Sprintf("projects/%s", projectID),
			PageSize:  10,
			PageToken: "",
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, 1, len(resp.GetSubscriptions()))
		actualSub := resp.GetSubscriptions()[0]
		expected := fmt.Sprintf("projects/%s/subscriptions/%s", projectID, subscriberID)
		require.Equal(t, expected, actualSub.GetName())
		// asserts values were set
		require.EqualValues(t, 1000, actualSub.GetRetryPolicy().GetMaximumBackoff().Seconds)
		require.EqualValues(t, 100, actualSub.GetRetryPolicy().GetMinimumBackoff().Seconds)

		// cleanup resources
		require.NoError(t, emulator.Cleanup())
		require.NoError(t, client.Close())
	})
	t.Run("successful with existing subscription", func(t *testing.T) {
		// create the go context
		ctx := context.TODO()
		emulator := NewEmulator()

		// set the emulator addr
		t.Setenv("PUBSUB_EMULATOR_HOST", emulator.EndPoint())

		// create a pubsub client
		client, err := pubsub.NewClient(ctx, projectID)
		require.NoError(t, err)
		require.NotNil(t, client)

		// create an instance of the management suite
		mgmt := NewTooling(client)
		require.NotNil(t, mgmt)

		// create the topic using the management API
		_, err = mgmt.CreateTopic(ctx, topicName)
		require.NoError(t, err)

		// create the topic to use
		topic, err := client.TopicAdminClient.GetTopic(ctx, &pubsubpb.GetTopicRequest{Topic: TopicFullName(projectID, topicName)})
		require.NoError(t, err)
		require.NotNil(t, topic)

		// let us start consuming the messages
		subCfg := &pubsubpb.Subscription{
			Topic: topic.GetName(),
			Name:  SubscriptionFullName(projectID, subscriberID),
			RetryPolicy: &pubsubpb.RetryPolicy{
				MinimumBackoff: durationpb.New(time.Second * 100),
				MaximumBackoff: durationpb.New(time.Second * 1000),
			},
			AckDeadlineSeconds: 0,
		}

		receiveSettings := &pubsub.ReceiveSettings{
			MaxExtension:               1,
			MaxDurationPerAckExtension: time.Millisecond * 50,
			MinDurationPerAckExtension: time.Millisecond * 100,
			MaxOutstandingMessages:     1,
			MaxOutstandingBytes:        1048576, // 1mb
			NumGoroutines:              1,
		}

		// create an instance of the subscriber
		subClient, err := NewSubscriber(ctx, client, &SubscriberConfig{
			SubscriptionConfig: subCfg,
			ReceiveSettings:    receiveSettings,
			Logger:             zapl.DiscardLogger,
		})
		require.NotNil(t, subClient)
		require.NoError(t, err)

		// fetch the list of subscriber
		resp, err := emulator.Server().GServer.ListSubscriptions(ctx, &pubsubpb.ListSubscriptionsRequest{
			Project:   fmt.Sprintf("projects/%s", projectID),
			PageSize:  1,
			PageToken: "",
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, 1, len(resp.GetSubscriptions()))
		actualSub := resp.GetSubscriptions()[0]
		expected := fmt.Sprintf("projects/%s/subscriptions/%s", projectID, subscriberID)
		require.Equal(t, expected, actualSub.GetName())
		// asserts values were updated
		require.EqualValues(t, 1000, actualSub.GetRetryPolicy().GetMaximumBackoff().Seconds)
		require.EqualValues(t, 100, actualSub.GetRetryPolicy().GetMinimumBackoff().Seconds)

		// cleanup resources
		require.NoError(t, emulator.Cleanup())
		require.NoError(t, client.Close())
	})
	t.Run("with config not set", func(t *testing.T) {
		ctx := context.TODO()
		emulator := NewEmulator()

		t.Setenv("PUBSUB_EMULATOR_HOST", emulator.EndPoint())

		projectID := "test"
		// create a pubsub client
		client, err := pubsub.NewClient(ctx, projectID)
		require.NoError(t, err)
		require.NotNil(t, client)

		// create an instance of the subscriber
		subClient, err := NewSubscriber(ctx, client, nil)
		require.Nil(t, subClient)
		require.Error(t, err)
		require.EqualError(t, err, "config is not set")

		require.NoError(t, emulator.Cleanup())
		require.NoError(t, client.Close())
	})
	t.Run("with invalid config", func(t *testing.T) {
		ctx := context.TODO()
		emulator := NewEmulator()

		t.Setenv("PUBSUB_EMULATOR_HOST", emulator.EndPoint())

		projectID := "test"
		// create a pubsub client
		client, err := pubsub.NewClient(ctx, projectID)
		assert.NoError(t, err)
		assert.NotNil(t, client)

		subscriberConfig := &SubscriberConfig{
			Logger: zapl.DiscardLogger,
		}
		// create an instance of the subscriber
		subClient, err := NewSubscriber(ctx, client, subscriberConfig)
		assert.Nil(t, subClient)
		assert.Error(t, err)

		assert.NoError(t, emulator.Cleanup())
		assert.NoError(t, client.Close())
	})
}

func TestConsume(t *testing.T) {
	t.Run("consume a published message with successful handler", func(t *testing.T) {
		// create the go context
		ctx := context.TODO()
		emulator := NewEmulator()

		// set the emulator addr
		t.Setenv("PUBSUB_EMULATOR_HOST", emulator.EndPoint())

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

		// let us start consuming the messages
		subCfg := &pubsubpb.Subscription{Topic: TopicFullName(projectID, topicName), Name: SubscriptionFullName(projectID, subscriberID)}

		// create an instance of the subscriber
		subClient, err := NewSubscriber(ctx, client, &SubscriberConfig{
			SubscriptionConfig: subCfg,
			Logger:             zapl.DiscardLogger,
		})
		assert.NotNil(t, subClient)
		assert.NoError(t, err)

		// fetch the list of subscriber
		resp, err := emulator.Server().GServer.ListSubscriptions(ctx, &pubsubpb.ListSubscriptionsRequest{
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
	})
	t.Run("consume a published message with failure handler", func(t *testing.T) {
		// create the go context
		ctx := context.TODO()
		emulator := NewEmulator()

		t.Setenv("PUBSUB_EMULATOR_HOST", emulator.EndPoint())

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

		// let us start consuming the messages
		subCfg := &pubsubpb.Subscription{Topic: TopicFullName(projectID, topicName), Name: SubscriptionFullName(projectID, subscriberID)}

		// create an instance of the subscriber
		subClient, err := NewSubscriber(ctx, client, &SubscriberConfig{
			SubscriptionConfig: subCfg,
			Logger:             zapl.DiscardLogger,
		})
		assert.NotNil(t, subClient)
		assert.NoError(t, err)

		// fetch the list of subscriber
		resp, err := emulator.Server().GServer.ListSubscriptions(ctx, &pubsubpb.ListSubscriptionsRequest{
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
	})
	t.Run("consume many messages", func(t *testing.T) {
		// create the go context
		ctx := context.TODO()
		emulator := NewEmulator()

		t.Setenv("PUBSUB_EMULATOR_HOST", emulator.EndPoint())

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
	})
}
