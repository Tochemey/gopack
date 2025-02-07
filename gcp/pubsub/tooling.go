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

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/iterator"
)

// Tooling helps perform some management tasks via
// the PubSub client
type Tooling struct {
	remote *pubsub.Client
}

// NewTooling creates an instance of Tooling
func NewTooling(remote *pubsub.Client) *Tooling {
	return &Tooling{remote: remote}
}

// CreateTopic creates a GCP Pub/Sub topic
// The specified topic name must start with a letter, and contain only letters
// ([A-Za-z]), numbers ([0-9]), dashes (-), underscores (_), periods (.),
// tildes (~), plus (+) or percent signs (%). It must be between 3 and 255
// characters in length, and must not start with "goog". For more information,
// see: https://cloud.google.com/pubsub/docs/admin#resource_names.
func (c Tooling) CreateTopic(ctx context.Context, topicName string) (*pubsub.Topic, error) {
	// make the call to GCP
	topic, err := c.remote.CreateTopic(ctx, topicName)
	// handle the eventual error
	if err != nil {
		// return the result
		return nil, err
	}
	return topic, nil
}

// ListTopics fetches the list all PubSub topics in a given GCP project
// TODO figure out the way to perform the paginated requests
func (c Tooling) ListTopics(ctx context.Context) ([]*pubsub.Topic, error) {
	var topics []*pubsub.Topic
	it := c.remote.Topics(ctx)
	for {
		topic, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		topics = append(topics, topic)
	}
	return topics, nil
}
