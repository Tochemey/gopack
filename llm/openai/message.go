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

package openai

import "image"

// RequestType defines the query message type
type RequestType int

const (
	// UserMessage defines a user message when calling the OpenAI apis
	UserMessage RequestType = iota
	// SystemMessage defines a system message when calling the OpenAI apis
	SystemMessage
	// AssistantMessage defines an assistant message when calling the OpenAI apis
	AssistantMessage
)

// ResponseType defines the query response type
type ResponseType int

const (
	// JSONResponseType defines the OpenAI query JSON response type
	JSONResponseType ResponseType = iota
	// TextResponseType defines the OpenAI query TEXT response type
	TextResponseType
)

// Request defines the query message sent to OpenAI
type Request struct {
	// Type specifies the message type
	Type RequestType
	// Content specifies the message content
	Content string
}

// VisionRequest defines an image message request sent to OpenAI
type VisionRequest struct {
	// Type specifies the message type
	Type RequestType
	// Content specifies the message content
	Content string
	// Image specifies the image content
	Image image.Image
}

// Response defines the OpenAI response
type Response struct {
	// Content specifies the response content
	Content          string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}
