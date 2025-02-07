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

	"cloud.google.com/go/pubsub"

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
	subscription *pubsub.Subscription

	// internal components
	mutex sync.Mutex
	// useful to hook in metrics
	// TODO add more later on
	messagesReceivedCount  int32
	messagesProcessedCount int32

	logger log.Logger
}

// NewSubscriber creates an instance of Subscriber
func NewSubscriber(ctx context.Context, remote *pubsub.Client, config *SubscriberConfig) (*Subscriber, error) {
	// assert the subscription config
	if config == nil {
		return nil, errors.New("config is not set")
	}
	// validate the config
	if err := config.Validate(); err != nil {
		return nil, err
	}
	// set up the topic
	topic := remote.Topic(config.SubscriptionConfig.Topic.ID())
	// check topic existence
	exists, err := topic.Exists(ctx)
	// handle error
	if err != nil {
		// return the error
		return nil, err
	}
	// check whether the topic exists or not
	if !exists {
		// create the error and return it
		return nil, fmt.Errorf("topic %s does not exist", config.SubscriptionConfig.Topic.ID())
	}

	// set the subscription
	subscription := remote.Subscription(config.SubscriptionID)
	// check the existence of the subscription
	exists, err = subscription.Exists(ctx)
	// handle error
	if err != nil {
		// return the error
		return nil, err
	}

	// check whether the retry policy is set and set a default one
	if config.SubscriptionConfig.RetryPolicy == nil {
		// https://cloud.google.com/pubsub/docs/reference/rest/v1/projects.subscriptions#retrypolicy
		config.SubscriptionConfig.RetryPolicy = &pubsub.RetryPolicy{
			MinimumBackoff: MinimumBackoff,
			MaximumBackoff: MaximumBackoff,
		}
	}

	// check whether the subscription exists or not
	if !exists {
		// create the subscription
		subscription, err = remote.CreateSubscription(ctx, config.SubscriptionID, *config.SubscriptionConfig)
		// handle the error
		if err != nil {
			// return the error
			return nil, err
		}
	} else {
		// update the existing subscription
		_, err := subscription.Update(ctx, subscriptionConfigToUpdate(config.SubscriptionConfig))
		// handle the error
		if err != nil {
			// return the error
			return nil, err
		}
	}

	// set the default receiveSettings
	subscription.ReceiveSettings = pubsub.DefaultReceiveSettings
	// override the default settings with the one provided if exists
	if config.ReceiveSettings != nil {
		subscription.ReceiveSettings = *config.ReceiveSettings
	}

	return &Subscriber{
		subscription: subscription,
		logger:       config.Logger,
	}, nil
}

// NewSubscriberWithDefaults creates an instance of Subscriber with the default settings
func NewSubscriberWithDefaults(ctx context.Context, remote *pubsub.Client, subscriptionID, topicName string) (*Subscriber, error) {
	// set the topic
	topic := remote.Topic(topicName)

	// set up the subscriber config
	subscriberConfig := &SubscriberConfig{
		SubscriptionID: subscriptionID,
		SubscriptionConfig: &pubsub.SubscriptionConfig{
			Topic:                 topic,
			AckDeadline:           10 * time.Second,
			ExpirationPolicy:      time.Duration(0),
			EnableMessageOrdering: true,
		},
		Logger: zapl.New(log.DebugLevel, os.Stdout),
	}

	// create the subscription connection
	subscriber, err := NewSubscriber(ctx, remote, subscriberConfig)
	// handle the error
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
		err := s.subscription.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
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

// subscriptionConfigToUpdate returns a SubscriptionConfigToUpdate from SubscriptionConfig
func subscriptionConfigToUpdate(config *pubsub.SubscriptionConfig) pubsub.SubscriptionConfigToUpdate {
	var pushConfig *pubsub.PushConfig
	var bigQueryConfig *pubsub.BigQueryConfig

	if config.PushConfig.Endpoint != "" {
		pushConfig = &config.PushConfig
	}

	if config.BigQueryConfig.Table != "" {
		bigQueryConfig = &config.BigQueryConfig
	}

	return pubsub.SubscriptionConfigToUpdate{
		PushConfig:          pushConfig,
		BigQueryConfig:      bigQueryConfig,
		AckDeadline:         config.AckDeadline,
		RetainAckedMessages: config.RetainAckedMessages,
		RetentionDuration:   config.RetentionDuration,
		ExpirationPolicy:    config.ExpirationPolicy,
		DeadLetterPolicy:    config.DeadLetterPolicy,
		Labels:              config.Labels,
		RetryPolicy:         config.RetryPolicy,
	}
}
