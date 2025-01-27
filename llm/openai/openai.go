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

import (
	"context"
	"errors"
	"net/http"

	"github.com/cenkalti/backoff/v4"
	openai "github.com/sashabaranov/go-openai"
	"golang.org/x/time/rate"
)

// API defines the OpenAI LLM integration
type API interface {
	// Query sends messages to OpenAI APIs and retrieves responses.
	//
	// This function interacts with OpenAI APIs to process a sequence of messages and
	// returns responses based on the specified `ResponseType`.
	//
	// Parameters:
	//   - ctx: A `context.Context` used to control the lifecycle of the request. It
	//     allows cancellation, timeouts, and deadlines.
	//   - requests: A list of `*Request` objects,
	//     where each request contains the input message, prompt, or query to send
	//     to the API.
	//   - responseType: Specifies the type of response expected from the OpenAI API.
	//     It determines how the API should process and format its output.
	//
	// Returns:
	//   - responses: A slice of `Response` objects representing the output generated
	//     by OpenAI APIs. Each response corresponds to an input request in `messages`.
	//   - err: An error if the request fails, such as due to network issues, invalid
	//     parameters, or API-specific errors.
	//
	// Usage Notes:
	//   - Ensure `ctx` has an appropriate timeout or deadline to prevent long-running
	//     requests from blocking your application.
	//   - The `ResponseType` should align with the API's expected output format. The following are supported: JSON and Text
	//   - Input messages in `Request` must follow the format expected by the OpenAI
	//     API. For instance, when working with chat models, include roles (e.g.,
	//     "user", "assistant") and content.
	//
	// Error Handling:
	//   - Returns a non-nil `err` if there is an issue with the API request or
	//     response parsing.
	//   - For successful calls, `err` will be nil, and `responses` will contain the
	//     API's output.
	//
	// Example:
	//   // ctx with timeout
	//   ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	//   defer cancel()
	//
	//   messages := []*Request{
	//       {Type: UserMessage, Content: "Hello, OpenAI!"},
	//   }
	//
	//   responses, err := Query(ctx, JSONResponseType, messages...)
	//   if err != nil {
	//       log.Fatalf("Query failed: %v", err)
	//   }
	//
	//   for _, response := range responses {
	//       fmt.Println("Response:", response.Content)
	//   }
	Query(ctx context.Context, requests []*Request, responseType ResponseType) (responses []*Response, err error)
	// VisionQuery sends image query requests to OpenAI and retrieves responses.
	//
	// This function interacts with OpenAI APIs to handle image-related requests
	// and returns the corresponding responses based on the provided input messages.
	//
	// Parameters:
	//   - ctx: A `context.Context` used to manage the lifecycle of the request. It
	//     supports cancellation, timeouts, and deadlines.
	//   - requests: A list of `*VisionRequest` objects.
	//     Each request contains the data required to query OpenAI APIs for image
	//     generation or processing.
	//
	// Returns:
	//   - responses: A slice of `Response` objects containing the results of the
	//     image queries. Each response corresponds to an input message in the `messages`
	//     parameter.
	//   - err: An error value indicating the success or failure of the request. If the
	//     operation fails, this will contain details about the issue.
	//
	// Usage Notes:
	//   - Ensure `ctx` is properly configured with a timeout or cancellation mechanism
	//     to prevent excessive blocking during API calls.
	//   - Each `VisionRequest` in `messages` must conform to the expected input
	//     format defined by OpenAI's image-related APIs. This includes specifying
	//     required fields such as image description, parameters, or any relevant metadata.
	//   - The function processes multiple image requests in a single call, returning
	//     a separate response for each input.
	//
	// Error Handling:
	//   - If the API call fails due to network issues, invalid parameters, or server
	//     errors, the function returns a non-nil `err`.
	//   - In case of a partial failure (e.g., one of several requests fails), the function
	//     may still return valid responses for the successful requests, depending on the
	//     API's behavior.
	//
	// Performance Considerations:
	//   - When querying with multiple `VisionRequest` objects, be mindful of the API's
	//     rate limits and response times.
	//   - For large or complex image queries, ensure the client application can handle
	//     the potentially high payload size of the responses.
	VisionQuery(ctx context.Context, messages ...*VisionRequest) (responses []*Response, err error)
}

type api struct {
	config      *Config
	remote      *openai.Client
	temperature float32 // temp for calls
	frequency   float32 // frequency penalty
	presence    float32 // presence penalty
	rateLimit   *rate.Limiter
	httpClient  *http.Client
}

// enforce compilation error
var _ API = (*api)(nil)

// NewAPI creates an instance of the Open API wrapper
func NewAPI(config *Config, opts ...Option) API {
	// TODO: add this configuration
	// 90k tokens per minute, halved as to not deplete other resources
	// tpm := 45000
	tpm := 1000000
	tokensPerSecond := tpm / 60

	api := &api{
		config:      config,
		temperature: 0,
		frequency:   0,
		presence:    0,
		rateLimit:   rate.NewLimiter(rate.Limit(tokensPerSecond), tpm),
		httpClient:  http.DefaultClient,
	}

	// apply the options
	for _, opt := range opts {
		opt.Apply(api)
	}

	// create the remote openai configuration
	cfg := openai.DefaultConfig(config.Token)
	cfg.HTTPClient = api.httpClient
	if config.Organization != "" {
		cfg.OrgID = config.Organization
	}

	api.remote = openai.NewClientWithConfig(cfg)
	return api
}

// Query sends messages to OpenAI APIs and retrieves responses.
//
// This function interacts with OpenAI APIs to process a sequence of messages and
// returns responses based on the specified `ResponseType`.
//
// Parameters:
//   - ctx: A `context.Context` used to control the lifecycle of the request. It
//     allows cancellation, timeouts, and deadlines.
//   - responseType: Specifies the type of response expected from the OpenAI API.
//     It determines how the API should process and format its output.
//   - messages: A variadic parameter representing a list of `*Request` objects,
//     where each request contains the input message, prompt, or query to send
//     to the API.
//
// Returns:
//   - responses: A slice of `*Response` objects representing the output generated
//     by OpenAI APIs. Each response corresponds to an input request in `messages`.
//   - err: An error if the request fails, such as due to network issues, invalid
//     parameters, or API-specific errors.
//
// Usage Notes:
//   - Ensure `ctx` has an appropriate timeout or deadline to prevent long-running
//     requests from blocking your application.
//   - The `ResponseType` should align with the API's expected output format. The following are supported: JSON and Text
//   - Input messages in `*Request` must follow the format expected by the OpenAI
//     API. For instance, when working with chat models, include roles (e.g.,
//     "user", "assistant") and content.
//
// Error Handling:
//   - Returns a non-nil `err` if there is an issue with the API request or
//     response parsing.
//   - For successful calls, `err` will be nil, and `responses` will contain the
//     API's output.
//
// Example:
//
//	// ctx with timeout
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	messages := []*Request{
//	    {Type: UserMessage, Content: "Hello, OpenAI!"},
//	}
//
//	responses, err := Query(ctx, JSONResponseType, messages...)
//	if err != nil {
//	    log.Fatalf("Query failed: %v", err)
//	}
//
//	for _, response := range responses {
//	    fmt.Println("Response:", response.Content)
//	}
func (x api) Query(ctx context.Context, requests []*Request, responseType ResponseType) (responses []*Response, err error) {
	msgs := make([]openai.ChatCompletionMessage, 0, len(requests))
	for _, message := range requests {
		msg, err := toChatCompletionMessage(message)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	}

	tokens, err := tokensCount(msgs, x.config.Model)
	if err != nil {
		return nil, err
	}

	// estimating 100 tokens of response
	// TODO: make this configurable
	tokens += 100

	if err := x.rateLimit.WaitN(ctx, tokens); err != nil {
		return nil, err
	}

	// create request
	req := openai.ChatCompletionRequest{
		Model:            x.config.Model,
		Messages:         msgs,
		Temperature:      x.temperature,
		PresencePenalty:  x.presence,
		FrequencyPenalty: x.frequency,
	}

	switch {
	case responseType == JSONResponseType:
		req.ResponseFormat = &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		}
	case responseType == TextResponseType:
		req.ResponseFormat = &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeText,
		}
	}

	var resp openai.ChatCompletionResponse
	// wrap in a function so we can backoff
	operation := func() error {
		ctx, cancel := context.WithTimeout(ctx, x.config.Timeout)
		var err error
		resp, err = x.remote.CreateChatCompletion(ctx, req)
		defer cancel()
		if err != nil {
			e := &openai.APIError{}
			switch {
			case errors.As(err, &e):
				switch e.HTTPStatusCode {
				case http.StatusUnauthorized:
					// invalid auth or key (do not retry)
					return &backoff.PermanentError{Err: err}
				case http.StatusTooManyRequests:
					// rate limiting or engine overload (wait and retry)
					return err
				case http.StatusInternalServerError:
					// openai server error (retry)
					return err
				default:
					// return &backoff.PermanentError{Err: err}
					return err
				}
			default:
				return err
			}
		}
		return nil
	}

	// implements backoff
	opt := backoff.WithMaxRetries(backoff.NewExponentialBackOff(), uint64(x.config.MaxRetries))
	if err := backoff.Retry(operation, opt); err != nil {
		return nil, err
	}

	// when we have no choices
	if len(resp.Choices) == 0 {
		return nil, errors.New("malformed llm response from openai")
	}

	responses = make([]*Response, len(resp.Choices))
	for i, choice := range resp.Choices {
		responses[i] = &Response{
			Content:          choice.Message.Content,
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		}
	}

	return responses, nil
}

// VisionQuery sends image query requests to OpenAI and retrieves responses.
//
// This function interacts with OpenAI APIs to handle image-related requests
// and returns the corresponding responses based on the provided input messages.
//
// Parameters:
//   - ctx: A `context.Context` used to manage the lifecycle of the request. It
//     supports cancellation, timeouts, and deadlines.
//   - requests: A variadic parameter representing a list of `*VisionRequest` objects.
//     Each request contains the data required to query OpenAI APIs for image
//     generation or processing.
//
// Returns:
//   - responses: A slice of `Response` objects containing the results of the
//     image queries. Each response corresponds to an input message in the `messages`
//     parameter.
//   - err: An error value indicating the success or failure of the request. If the
//     operation fails, this will contain details about the issue.
//
// Usage Notes:
//   - Ensure `ctx` is properly configured with a timeout or cancellation mechanism
//     to prevent excessive blocking during API calls.
//   - Each `VisionRequest` in `messages` must conform to the expected input
//     format defined by OpenAI's image-related APIs. This includes specifying
//     required fields such as image description, parameters, or any relevant metadata.
//   - The function processes multiple image requests in a single call, returning
//     a separate response for each input.
//
// Error Handling:
//   - If the API call fails due to network issues, invalid parameters, or server
//     errors, the function returns a non-nil `err`.
//   - In case of a partial failure (e.g., one of several requests fails), the function
//     may still return valid responses for the successful requests, depending on the
//     API's behavior.
//
// Performance Considerations:
//   - When querying with multiple `VisionRequest` objects, be mindful of the API's
//     rate limits and response times.
//   - For large or complex image queries, ensure the client application can handle
//     the potentially high payload size of the responses.
func (x api) VisionQuery(ctx context.Context, requests ...*VisionRequest) (responses []*Response, err error) {
	convertedMessages, err := transformImageRequests(requests)
	if err != nil {
		return nil, err
	}

	tokens, err := tokensCount(convertedMessages, x.config.Model)
	if err != nil {
		return nil, err
	}

	// estimating 100 tokens of response
	tokens += 400
	if err := x.rateLimit.WaitN(ctx, tokens); err != nil {
		return nil, err
	}

	// random seed
	seed := 8006
	// create request
	req := openai.ChatCompletionRequest{
		Model:            x.config.Model,
		Messages:         convertedMessages,
		Temperature:      x.temperature,
		PresencePenalty:  x.presence,
		FrequencyPenalty: x.frequency,
		// 4096 is the max tokens so take that minus the estimated amount
		MaxTokens: 4096 - tokens,
		Seed:      &seed,
	}

	var resp openai.ChatCompletionResponse
	// wrap in a function so we can backoff
	operation := func() error {
		ctx, cancel := context.WithTimeout(ctx, x.config.Timeout)
		var err error
		resp, err = x.remote.CreateChatCompletion(ctx, req)
		cancel()
		if err != nil {
			e := &openai.APIError{}
			if errors.As(err, &e) {
				switch e.HTTPStatusCode {
				case http.StatusUnauthorized:
					// invalid auth or key (do not retry)
					return &backoff.PermanentError{Err: err}
				case http.StatusTooManyRequests:
					// rate limiting or engine overload (wait and retry)
					return err
				case http.StatusInternalServerError:
					// openai server error (retry)
					return err
				default:
					// return &backoff.PermanentError{Err: err}
					return err
				}
			} else {
				// it means this is not an openai error
				// return &backoff.PermanentError{Err: err}
				return err
			}
		}
		return nil
	}

	// implements backoff
	opt := backoff.WithMaxRetries(backoff.NewExponentialBackOff(), uint64(x.config.MaxRetries))
	if err := backoff.Retry(operation, opt); err != nil {
		return nil, err
	}

	// when we have no choices
	if len(resp.Choices) == 0 {
		return nil, errors.New("malformed llm response from openai")
	}

	responses = make([]*Response, len(resp.Choices))
	for i, choice := range resp.Choices {
		responses[i] = &Response{
			Content:          choice.Message.Content,
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		}
	}

	return responses, nil
}
