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
	"sync"

	"cloud.google.com/go/pubsub"

	"github.com/tochemey/gopack/log"
)

// Publisher implements the Publisher interface
type Publisher struct {
	Remote *pubsub.Client
	mutex  sync.Mutex
	logger log.Logger
}

// NewPublisher creates an instance of publisher
func NewPublisher(remote *pubsub.Client, logger log.Logger) *Publisher {
	return &Publisher{
		Remote: remote,
		mutex:  sync.Mutex{},
		logger: logger,
	}
}

// Publish will persist a batch of messages to pubsub
func (p *Publisher) Publish(ctx context.Context, topic *Topic, messages []*Message) error {
	// publish when connected
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// get the context logger
	log := p.logger.WithContext(ctx)
	// add some debug logging
	log.Debugf("publishing to GCP Pub/Sub topic=%s", topic.Name)
	// reference the given topic
	t := p.Remote.Topic(topic.Name)
	// set the message ordering to true to enable ordering publication
	t.EnableMessageOrdering = topic.EnableOrdering
	if topic.PublishSettings != nil {
		t.PublishSettings = *topic.PublishSettings
	}

	var results []*pubsub.PublishResult
	for _, message := range messages {
		// let us create the message to publish
		pubsubMessage := &pubsub.Message{
			OrderingKey: message.Key,
			Data:        message.Payload,
		}
		// if ordering is required then set the key
		if t.EnableMessageOrdering {
			// ignore that message when ordering is required
			// and the given message to publish does not have the required key
			if message.Key == "" {
				return errors.New("message key is required when MessageOrdering is enabled")
			}
		}

		// let us publish the message
		result := t.Publish(ctx, pubsubMessage)
		// append the result to the results list
		results = append(results, result)
	}

	resultErrors := make([]error, 0, len(results))
	// block and wait for the results of published messages
	for _, res := range results {
		// Block until the result is returned and a server-generated
		// ID is returned for the published message.
		_, err := res.Get(ctx)
		// handle the eventual error
		if err != nil {
			// wraps the error
			e := fmt.Errorf("unable to publish message to GCP Pub/Sub: %w", err)
			// log the error
			log.Error(e.Error())
			// append the errors
			resultErrors = append(resultErrors, e)
			continue
		}
	}
	// in case of an error return an error
	if len(resultErrors) != 0 {
		return fmt.Errorf("%v", resultErrors[len(resultErrors)-1])
	}
	log.Debugf("successfully published %d messages to GCP Pub/Sub topic=%s", len(messages), topic.Name)
	return nil
}
