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
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"strings"

	"github.com/pkoukk/tiktoken-go"
	"github.com/sashabaranov/go-openai"
)

func transformImageRequests(imageRequests []*VisionRequest) ([]openai.ChatCompletionMessage, error) {
	out := openai.ChatCompletionMessage{
		// TODO: consider making this dyanmic based upon messages surrounding
		Role:         openai.ChatMessageRoleUser,
		MultiContent: []openai.ChatMessagePart{},
	}

	for _, msg := range imageRequests {
		switch {
		case msg.Image != nil:
			imgInput, err := toString(msg.Image)
			if err != nil {
				return []openai.ChatCompletionMessage{out}, fmt.Errorf("image failed to convert: %v", err)
			}
			out.MultiContent = append(out.MultiContent, openai.ChatMessagePart{
				Type: openai.ChatMessagePartTypeImageURL,
				ImageURL: &openai.ChatMessageImageURL{
					URL: imgInput,
				},
			})
		default:
			out.MultiContent = append(out.MultiContent, openai.ChatMessagePart{
				Type: openai.ChatMessagePartTypeText,
				Text: msg.Content,
			})
		}
	}

	arr := []openai.ChatCompletionMessage{out}
	return arr, nil
}

// toString converts an image to string
func toString(image image.Image) (string, error) {
	buff := new(bytes.Buffer)
	if err := jpeg.Encode(buff, image, &jpeg.Options{
		Quality: 100,
	}); err != nil {
		return "", fmt.Errorf("encoding for model: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(buff.Bytes())
	encodedURL := fmt.Sprintf("data:image/jpeg;base64,%s", encoded)
	return encodedURL, nil
}

// toChatCompletionMessage converts a message to an openai chat completion message
func toChatCompletionMessage(query *Request) (openai.ChatCompletionMessage, error) {
	message := openai.ChatCompletionMessage{
		Content: query.Content,
	}
	switch query.Type {
	case SystemMessage:
		message.Role = openai.ChatMessageRoleSystem
	case AssistantMessage:
		message.Role = openai.ChatMessageRoleAssistant
	case UserMessage:
		message.Role = openai.ChatMessageRoleUser
	default:
		return message, fmt.Errorf("unknown type: %T", query.Type)
	}
	return message, nil
}

// tokensCount estimates the number of tokens for a given array of messages
// https://github.com/pkoukk/tiktoken-go#counting-tokens-for-chat-api-calls
func tokensCount(messages []openai.ChatCompletionMessage, model string) (numTokens int, err error) {
	tkm, err := tiktoken.EncodingForModel(model)
	if err != nil {
		err = fmt.Errorf("encoding for model: %v", err)
		return
	}

	var tokensPerMessage, tokensPerName int
	switch model {
	case openai.GPT3Dot5Turbo0613,
		openai.GPT3Dot5Turbo16K0613,
		openai.GPT40314,
		openai.GPT432K0314,
		openai.GPT40613,
		openai.GPT432K0613:
		tokensPerMessage = 3
		tokensPerName = 1
	case openai.GPT3Dot5Turbo0301:
		tokensPerMessage = 4 // every message follows <|start|>{role/name}\n{content}<|end|>\n
		tokensPerName = -1   // if there's a name, the role is omitted
	case openai.GPT4VisionPreview:
		tokensPerMessage = 1500 // every message follows <|start|>{role/name}\n{content}<|end|>\n
		tokensPerName = -1      // if there's a name, the role is omitted
	case openai.GPT4Turbo:
		tokensPerMessage = 1000 // every message follows <|start|>{role/name}\n{content}<|end|>\n
		tokensPerName = -1      // if there's a name, the role is omitted
	default:
		switch {
		case strings.Contains(model, openai.GPT3Dot5Turbo):
			return tokensCount(messages, openai.GPT3Dot5Turbo0613)
		case strings.Contains(model, openai.GPT4):
			return tokensCount(messages, openai.GPT40613)
		default:
			err = fmt.Errorf("num_tokens_from_messages() is not implemented for model %s. See https://github.com/openai/openai-python/blob/main/chatml.md for information on how messages are converted to tokens", model)
			return
		}
	}

	for _, message := range messages {
		numTokens += tokensPerMessage
		numTokens += len(tkm.Encode(message.Content, nil, nil))
		numTokens += len(tkm.Encode(message.Role, nil, nil))
		numTokens += len(tkm.Encode(message.Name, nil, nil))
		if message.Name != "" {
			numTokens += tokensPerName
		}
	}
	numTokens += 3 // every reply is primed with <|start|>assistant<|message|>
	return numTokens, nil
}
