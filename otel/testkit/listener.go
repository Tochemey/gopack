/*
 * MIT License
 *
 * Copyright (c) 2022-2023 Tochemey
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

package testkit

import (
	"net"
	"runtime"
	"strings"
	"sync"
)

// Listener wraps the net.Listener
type Listener struct {
	closeOnce sync.Once
	wrapped   net.Listener
	C         chan struct{}
}

// NewListener creates an instance of Listener
func NewListener(wrapped net.Listener) *Listener {
	return &Listener{
		wrapped: wrapped,
		C:       make(chan struct{}, 1),
	}
}

// Close closes the underlying listener
func (l *Listener) Close() error { return l.wrapped.Close() }

// Addr returns the listener address
func (l *Listener) Addr() net.Addr { return l.wrapped.Addr() }

// Accept waits for and returns the next connection to the listener. It will
// send a signal on l.C that a connection has been made before returning.
func (l *Listener) Accept() (net.Conn, error) {
	conn, err := l.wrapped.Accept()
	if err != nil {
		// Go 1.16 exported net.ErrClosed that could clean up this check, but to
		// remain backwards compatible with previous versions of Go that we
		// support the following string evaluation is used instead to keep in line
		// with the previously recommended way to check this:
		// https://github.com/golang/go/issues/4373#issuecomment-353076799
		if strings.Contains(err.Error(), "use of closed network connection") {
			// If the listener has been closed, do not allow callers of
			// WaitForConn to wait for a connection that will never come.
			l.closeOnce.Do(func() { close(l.C) })
		}
		return conn, err
	}

	select {
	case l.C <- struct{}{}:
	default:
		// If C is full, assume nobody is listening and move on.
	}
	return conn, nil
}

// WaitForConn will wait indefinitely for a connection to be established with
// the listener before returning.
func (l *Listener) WaitForConn() {
	for {
		select {
		case <-l.C:
			return
		default:
			runtime.Gosched()
		}
	}
}
