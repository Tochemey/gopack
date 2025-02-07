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

package scheduler

import (
	"fmt"

	quartzlogger "github.com/reugn/go-quartz/logger"

	"github.com/tochemey/gopack/log"
)

type logWrapper struct {
	logger log.Logger
}

var _ quartzlogger.Logger = (*logWrapper)(nil)

func newLogWrapper(logger log.Logger) *logWrapper {
	return &logWrapper{
		logger: logger,
	}
}

func (l logWrapper) Trace(msg string, args ...any) {
	l.logger.Debug(fmt.Sprintf(msg, args...))
}

func (l logWrapper) Debug(msg string, args ...any) {
	l.logger.Debug(fmt.Sprintf(msg, args...))
}

func (l logWrapper) Info(msg string, args ...any) {
	l.logger.Info(fmt.Sprintf(msg, args...))
}

func (l logWrapper) Warn(msg string, args ...any) {
	l.logger.Warn(fmt.Sprintf(msg, args...))
}

func (l logWrapper) Error(msg string, args ...any) {
	l.logger.Error(fmt.Sprintf(msg, args...))
}
