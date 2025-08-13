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
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/tochemey/gopack/log"
	"github.com/tochemey/gopack/log/zapl"
)

const (
	MinimumBackoff = 200 * time.Millisecond
	MaximumBackoff = 600 * time.Second
)

// SubscriptionHandler Handle processes the received message. When the handler fails to process the message
// then the message is replayed back to the handler based upon the retry policy on best effort basis
// ref: https://cloud.google.com/pubsub/docs/reference/rest/v1/projects.subscriptions#retrypolicy
type SubscriptionHandler func(ctx context.Context, data []byte) error

// Subscriber implements the Subscriber interface
type Subscriber struct {
	underlying *pubsub.Subscriber

	// internal components
	mutex sync.Mutex
	// useful to hook in metrics
	// TODO add more later on
	messagesReceivedCount  int32
	messagesProcessedCount int32

	logger log.Logger
}

// NewSubscriber creates an instance of Subscriber
func NewSubscriber(ctx context.Context, client *pubsub.Client, cfg *SubscriberConfig) (*Subscriber, error) {
	if cfg == nil {
		return nil, errors.New("config is not set")
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Ensure topic exists
	if err := ensureTopic(ctx, client, cfg.SubscriptionConfig.Topic); err != nil {
		return nil, err
	}

	// Apply defaults if not provided
	applyDefaults(cfg)

	// Ensure subscription exists (or update if it already exists)
	sub, err := ensureSubscription(ctx, client, cfg.SubscriptionConfig)
	if err != nil {
		return nil, err
	}

	// Configure subscriber
	subscriber := client.Subscriber(sub.GetName())
	subscriber.ReceiveSettings = pubsub.DefaultReceiveSettings
	if cfg.ReceiveSettings != nil {
		subscriber.ReceiveSettings = *cfg.ReceiveSettings
	}

	return &Subscriber{
		underlying: subscriber,
		logger:     cfg.Logger,
	}, nil
}

// NewSubscriberWithDefaults creates an instance of Subscriber with the default settings
func NewSubscriberWithDefaults(ctx context.Context, client *pubsub.Client, subscriptionID, topicName string) (*Subscriber, error) {
	subscriberConfig := &SubscriberConfig{
		SubscriptionID: subscriptionID,
		SubscriptionConfig: &pubsubpb.Subscription{
			Topic:                 topicName,
			AckDeadlineSeconds:    10,
			EnableMessageOrdering: true,
		},
		Logger: zapl.New(log.DebugLevel, os.Stdout),
	}

	subscriber, err := NewSubscriber(ctx, client, subscriberConfig)
	if err != nil {
		return nil, err
	}

	return subscriber, nil
}

// Consume receives messages from the topic and pass it to the
// message handler and the buffered channel to keep track of errors
// ref: https://cloud.google.com/go/docs/reference/cloud.google.com/go/pubsub/latest#receiving
func (s *Subscriber) Consume(ctx context.Context, handler SubscriptionHandler, errChan chan error) {
	// make sure to close the channel when done
	defer close(errChan)
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// set logging just for debug purpose
	logger := s.logger.WithContext(ctx)
	logger.Debug("start consuming messages")
	message := make(chan *pubsub.Message, 1)

	// consume messages
	go func() {
		err := s.underlying.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
			message <- msg
		})
		if err != nil {
			select {
			// return when the channel is closed
			case <-errChan:
				return
			default:
				// send the error to channel and return
				errChan <- err
				return
			}
		}
	}()

	// handle the message consumed
	for {
		select {
		case msg := <-message:
			// add some debug logging
			logger.Debugf("received message=%s", msg.ID)
			// set the messages received counter
			atomic.AddInt32(&s.messagesReceivedCount, 1)
			// pass the consumed message to the handler
			if err := handler(ctx, msg.Data); err != nil {
				// set the errChan return
				errChan <- err
				// we don't acknowledge the message and allow a quick redelivery rather
				// than awaiting the message expiration
				msg.Nack()
				return
			}
			// acknowledge that message has been processed
			atomic.AddInt32(&s.messagesProcessedCount, 1)
			msg.Ack()
		case <-ctx.Done():
			// add some debug messaging
			logger.Debugf("Total messages received=%d", s.messagesReceivedCount)
			return
		}
	}
}

// ensureTopic checks if a topic exists, and creates it if missing.
func ensureTopic(ctx context.Context, client *pubsub.Client, topicName string) error {
	topic, err := client.TopicAdminClient.GetTopic(ctx, &pubsubpb.GetTopicRequest{Topic: topicName})
	if err != nil {
		return err
	}
	if isEmptyTopic(topic) {
		topic, err = client.TopicAdminClient.CreateTopic(ctx, &pubsubpb.Topic{Name: topicName})
		if err != nil {
			return err
		}
		if isEmptyTopic(topic) {
			return fmt.Errorf("topic %s does not exist", topicName)
		}
	}
	return nil
}

func isEmptyTopic(t *pubsubpb.Topic) bool {
	return t == nil || proto.Equal(t, new(pubsubpb.Topic))
}

// ensureSubscription creates a subscription or updates it if it already exists.
func ensureSubscription(ctx context.Context, client *pubsub.Client, cfg *pubsubpb.Subscription) (*pubsubpb.Subscription, error) {
	sub, err := client.SubscriptionAdminClient.CreateSubscription(ctx, cfg)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.AlreadyExists:
				return client.SubscriptionAdminClient.UpdateSubscription(ctx,
					&pubsubpb.UpdateSubscriptionRequest{Subscription: cfg})
			case codes.NotFound:
				return nil, fmt.Errorf("topic %s not found", cfg.Topic)
			}
		}
		return nil, err
	}
	return sub, nil
}

// applyDefaults ensures RetryPolicy is set if missing.
func applyDefaults(cfg *SubscriberConfig) {
	if cfg.SubscriptionConfig.RetryPolicy == nil {
		cfg.SubscriptionConfig.RetryPolicy = &pubsubpb.RetryPolicy{
			MinimumBackoff: durationpb.New(MinimumBackoff),
			MaximumBackoff: durationpb.New(MaximumBackoff),
		}
	}
}
