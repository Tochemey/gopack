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
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	pstestpb "cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	pubsubv2pb "cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"cloud.google.com/go/pubsub/v2/pstest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"
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
	t.Run("consume a published message with panic handler", func(t *testing.T) {
		// create the go context
		ctx := context.TODO()
		emulator := NewEmulator()

		// set the emulator addr
		t.Setenv("PUBSUB_EMULATOR_HOST", emulator.EndPoint())

		client, err := pubsub.NewClient(ctx, projectID)
		require.NoError(t, err)
		require.NotNil(t, client)

		// create an instance of the management suite
		mgmt := NewTooling(client)
		require.NotNil(t, mgmt)

		// create the topic using the management API
		_, err = mgmt.CreateTopic(ctx, topicName)
		require.NoError(t, err)

		// create an instance of the publisher
		pub := NewPublisher(client, zapl.DiscardLogger)
		require.NotNil(t, pub)

		// let us start consuming the messages
		subCfg := &pubsubpb.Subscription{Topic: TopicFullName(projectID, topicName), Name: SubscriptionFullName(projectID, subscriberID)}

		// create an instance of the subscriber
		subClient, err := NewSubscriber(ctx, client, &SubscriberConfig{
			SubscriptionConfig: subCfg,
			Logger:             zapl.DiscardLogger,
		})
		require.NotNil(t, subClient)
		require.NoError(t, err)

		// create an object to persist
		acct := &account{
			AccountID:   "test-account-id",
			AccountName: "test-account-name",
		}
		// let us jsonify the account
		bytea, err := json.Marshal(acct)
		require.NoError(t, err)
		require.NotNil(t, bytea)

		// create the message
		message := &Message{
			Key:     "some-key",
			Payload: bytea,
		}

		pubTopic := &Topic{Name: topicName, EnableOrdering: true}
		// let us publish the message
		err = pub.Publish(ctx, pubTopic, []*Message{message})
		require.NoError(t, err)

		// consume some messages for 2 seconds
		cancelCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		handler := func(context.Context, []byte) error {
			panic("boom")
		}
		errChan := make(chan error, 1)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			subClient.Consume(cancelCtx, handler, errChan)
		}()

		select {
		case err := <-errChan:
			require.Error(t, err)
			require.Contains(t, err.Error(), "panic in subscription handler")
		case <-time.After(2 * time.Second):
			t.Fatal("expected handler panic error")
		}

		cancel()
		wg.Wait()
		for range errChan {
		}

		// cleanup resources
		require.NoError(t, emulator.Cleanup())
		require.NoError(t, client.Close())
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

func TestEnsureTopic(t *testing.T) {
	t.Run("returns error when get topic fails", func(t *testing.T) {
		_, client := newTestClient(t, pstest.WithErrorInjection("GetTopic", codes.Unknown, "boom"))
		err := ensureTopic(context.Background(), client, "bad-topic")
		require.Error(t, err)
	})

	t.Run("returns nil when topic exists", func(t *testing.T) {
		ctx := context.Background()
		_, client := newTestClient(t)

		_, err := client.TopicAdminClient.CreateTopic(ctx, &pubsubv2pb.Topic{
			Name: TopicFullName(projectID, "existing-topic"),
		})
		require.NoError(t, err)

		require.NoError(t, ensureTopic(ctx, client, "existing-topic"))
	})

	t.Run("returns error when topic creation fails", func(t *testing.T) {
		_, client := newTestClient(t, pstest.WithErrorInjection("CreateTopic", codes.Internal, "boom"))
		err := ensureTopic(context.Background(), client, "broken-topic")
		require.Error(t, err)
	})

	t.Run("returns error when topic creation returns empty topic", func(t *testing.T) {
		reactor := pstest.ServerReactorOption{
			FuncName: "CreateTopic",
			Reactor:  emptyTopicReactor{},
		}
		_, client := newTestClient(t, reactor)

		err := ensureTopic(context.Background(), client, "empty-topic")
		require.EqualError(t, err, "failed to create topic projects/test/topics/empty-topic: topic projects/test/topics/empty-topic creation returned empty topic")
	})
}

func TestEnsureSubscription(t *testing.T) {
	ctx := context.Background()

	t.Run("updates subscription when it already exists", func(t *testing.T) {
		updateReactor := pstest.ServerReactorOption{
			FuncName: "UpdateSubscription",
			Reactor:  updateSubscriptionReactor{},
		}
		_, client := newTestClient(t, updateReactor)
		topicFullName := TopicFullName(projectID, "ensure-sub-topic")
		_, err := client.TopicAdminClient.CreateTopic(ctx, &pubsubv2pb.Topic{Name: topicFullName})
		require.NoError(t, err)

		subscription := &pubsubv2pb.Subscription{
			Name:  SubscriptionFullName(projectID, "ensure-sub"),
			Topic: topicFullName,
		}
		_, err = client.SubscriptionAdminClient.CreateSubscription(ctx, subscription)
		require.NoError(t, err)

		subscription.AckDeadlineSeconds = 20
		subscription.RetryPolicy = &pubsubv2pb.RetryPolicy{
			MinimumBackoff: durationpb.New(time.Second),
			MaximumBackoff: durationpb.New(2 * time.Second),
		}

		updated, err := ensureSubscription(ctx, client, subscription)
		require.NoError(t, err)
		require.EqualValues(t, subscription.AckDeadlineSeconds, updated.AckDeadlineSeconds)
		require.True(t, proto.Equal(subscription.RetryPolicy, updated.RetryPolicy))
	})

	t.Run("returns not found when topic does not exist", func(t *testing.T) {
		_, client := newTestClient(t)
		subscription := &pubsubv2pb.Subscription{
			Name:  SubscriptionFullName(projectID, "missing-topic-sub"),
			Topic: TopicFullName(projectID, "missing-topic"),
		}

		_, err := ensureSubscription(ctx, client, subscription)
		require.EqualError(t, err, "topic projects/test/topics/missing-topic not found")
	})

	t.Run("returns underlying subscription creation error", func(t *testing.T) {
		_, client := newTestClient(t, pstest.WithErrorInjection("CreateSubscription", codes.PermissionDenied, "denied"))
		subscription := &pubsubv2pb.Subscription{
			Name:  SubscriptionFullName(projectID, "denied-sub"),
			Topic: TopicFullName(projectID, "denied-topic"),
		}

		_, err := ensureSubscription(ctx, client, subscription)
		require.Error(t, err)
	})
}

func TestNewSubscriberErrors(t *testing.T) {
	t.Run("returns error when ensureTopic fails", func(t *testing.T) {
		_, client := newTestClient(t, pstest.WithErrorInjection("GetTopic", codes.Unknown, "boom"))
		cfg := &SubscriberConfig{
			SubscriptionConfig: &pubsubv2pb.Subscription{
				Topic: "bad-topic",
				Name:  SubscriptionFullName(projectID, "bad-sub"),
			},
			Logger: zapl.DiscardLogger,
		}

		subscriber, err := NewSubscriber(context.Background(), client, cfg)
		require.Nil(t, subscriber)
		require.Error(t, err)
	})

	t.Run("returns error when ensureSubscription fails", func(t *testing.T) {
		_, client := newTestClient(t, pstest.WithErrorInjection("CreateSubscription", codes.InvalidArgument, "bad-sub"))
		cfg := &SubscriberConfig{
			SubscriptionConfig: &pubsubv2pb.Subscription{
				Topic: "create-subscription-error-topic",
				Name:  SubscriptionFullName(projectID, "create-subscription-error"),
			},
			Logger: zapl.DiscardLogger,
		}

		subscriber, err := NewSubscriber(context.Background(), client, cfg)
		require.Nil(t, subscriber)
		require.Error(t, err)
	})

	t.Run("NewSubscriberWithDefaults propagates creation errors", func(t *testing.T) {
		_, client := newTestClient(t, pstest.WithErrorInjection("CreateTopic", codes.Internal, "boom"))

		subscriber, err := NewSubscriberWithDefaults(context.Background(), client, "defaults-sub", "defaults-topic")
		require.Nil(t, subscriber)
		require.Error(t, err)
	})
}

func TestConsumeErrors(t *testing.T) {
	t.Run("sends receive errors to channel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		_, client := newTestClient(t)
		subscriber := &Subscriber{
			underlying: client.Subscriber("missing-subscription"),
			logger:     zapl.DiscardLogger,
		}

		errChan := make(chan error, 1)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			subscriber.Consume(ctx, func(context.Context, []byte) error { return nil }, errChan)
		}()

		select {
		case err := <-errChan:
			require.Error(t, err)
		case <-time.After(2 * time.Second):
			t.Fatal("expected error on errChan")
		}

		cancel()
		wg.Wait()
		for range errChan {
		}
	})

	t.Run("skips sending error when errChan already has a value", func(t *testing.T) {
		_, client := newTestClient(t)
		subscriber := &Subscriber{
			underlying: client.Subscriber("missing-subscription"),
			logger:     zapl.DiscardLogger,
		}

		errChan := make(chan error, 1)
		errChan <- errors.New("prefilled")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			subscriber.Consume(ctx, func(context.Context, []byte) error { return nil }, errChan)
		}()

		time.Sleep(50 * time.Millisecond)
		cancel()
		wg.Wait()
		for range errChan {
		}
	})
}

type emptyTopicReactor struct{}

func (emptyTopicReactor) React(_ interface{}) (handled bool, ret interface{}, err error) {
	return true, &pstestpb.Topic{}, nil
}

type updateSubscriptionReactor struct{}

func (updateSubscriptionReactor) React(req interface{}) (handled bool, ret interface{}, err error) {
	updateReq := req.(*pstestpb.UpdateSubscriptionRequest)
	return true, updateReq.Subscription, nil
}
