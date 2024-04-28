// The MIT License
//
// Copyright (c) 2022 Temporal Technologies Inc.  All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package devserver

import (
	"fmt"
	"net"
)

// Modified from Temporalite which itself modified from
// https://github.com/phayes/freeport/blob/95f893ade6f232a5f1511d61735d89b1ae2df543/freeport.go

func newPortProvider() *portProvider {
	return &portProvider{}
}

type portProvider struct {
	listeners []*net.TCPListener
}

// GetFreePort asks the kernel for a free open port that is ready to use.
// Returns the interface's IP and the free port.
func (p *portProvider) GetFreePort() (string, int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		if addr, err = net.ResolveTCPAddr("tcp6", "[::1]:0"); err != nil {
			return "", 0, fmt.Errorf("failed to get free port: %w", err)
		}
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return "", 0, err
	}

	p.listeners = append(p.listeners, l)
	tcpAddr := l.Addr().(*net.TCPAddr)

	return tcpAddr.IP.String(), tcpAddr.Port, nil
}

func (p *portProvider) Close() error {
	for _, l := range p.listeners {
		if err := l.Close(); err != nil {
			return err
		}
	}
	return nil
}

func getFreeHostPort() (string, error) {
	pp := newPortProvider()
	host, port, err := pp.GetFreePort()
	closeErr := pp.Close()
	if err != nil {
		return "", err
	} else if closeErr != nil {
		return "", fmt.Errorf("failed to close TCP listener: %w", closeErr)
	}
	return fmt.Sprintf("%v:%v", host, port), nil
}
