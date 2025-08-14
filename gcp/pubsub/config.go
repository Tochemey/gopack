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
	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"

	"github.com/tochemey/gopack/log"
	"github.com/tochemey/gopack/validation"
)

// SubscriberConfig holds the subscriber settings
type SubscriberConfig struct {
	SubscriptionConfig *pubsubpb.Subscription
	ReceiveSettings    *pubsub.ReceiveSettings
	Logger             log.Logger
}

// Validate validates the config
func (c *SubscriberConfig) Validate() error {
	return validation.New(validation.FailFast()).
		AddAssertion(c.SubscriptionConfig != nil, "subscription config is not set").
		AddAssertion(c.Logger != nil, "subscription logger is not set").
		AddAssertion(c.SubscriptionConfig != nil && c.SubscriptionConfig.Topic != "", "subscription topic is not set").
		AddAssertion(c.SubscriptionConfig != nil && c.SubscriptionConfig.Name != "", "subscription id is not set").
		Validate()
}
