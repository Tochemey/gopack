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

package llm

import (
	"context"
	"fmt"

	"github.com/invopop/jsonschema"
)

type ToolBox struct {
	tools []Tool
}

func (t *ToolBox) Add(tools ...Tool) {
	t.tools = append(t.tools, tools...)
}

func (t *ToolBox) Get(id string) (Tool, bool) {
	for _, tool := range t.tools {
		if id == tool.Name() {
			return tool, true
		}
	}
	return nil, false
}

func (t *ToolBox) List() []Tool {
	out := make([]Tool, 0, len(t.tools))
	out = append(out, t.tools...)
	return out
}

type Tool interface {
	Name() string
	Description() string
	Arguments() *jsonschema.Schema
	Run(context.Context, string) (string, error)
}

// GetJSONSchema returns the json schema for a given struct in this repo
func GetJSONSchema(path string, v any) (*jsonschema.Schema, error) {
	r := new(jsonschema.Reflector)
	if err := r.AddGoComments(path, "./"); err != nil {
		return nil, err
	}
	// schema := jsonschema.Reflect(v)
	schema := r.Reflect(v)
	for _, v := range schema.Definitions {
		return v, nil
	}

	return nil, fmt.Errorf("unable to find jsonschema for %s", path)
}
