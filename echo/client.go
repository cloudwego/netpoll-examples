/*
 * Copyright 2021 CloudWeGo
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/netpoll"
)

func main() {
	network, address, timeout := "tcp", "127.0.0.1:8080", 50*time.Millisecond

	// use default
	conn, _ := netpoll.DialConnection(network, address, timeout)
	conn.Close()

	// use dialer
	dialer := netpoll.NewDialer()
	conn, _ = dialer.DialConnection(network, address, timeout)

	finish := make(chan bool)
	conn.SetOnRequest(func(ctx context.Context, connection netpoll.Connection) error {
		reader := connection.Reader()
		defer reader.Release()
		msg, _ := reader.ReadString(reader.Len())
		fmt.Printf("[recv msg] %v\n", msg)
		finish <- true
		return nil
	})

	conn.AddCloseCallback(func(connection netpoll.Connection) error {
		fmt.Printf("[%v] connection closed\n", connection.RemoteAddr())
		return nil
	})

	// write & send message
	writer := conn.Writer()
	writer.WriteString("hello world")
	writer.Flush()

	_ = <-finish
}
