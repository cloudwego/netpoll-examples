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

	conn.AddCloseCallback(func(connection netpoll.Connection) error {
		fmt.Printf("[%v] connection closed\n", connection.RemoteAddr())
		return nil
	})

	// write & send message
	writer := conn.Writer()
	message := "hello world"
	writer.WriteString(message)
	writer.Flush()

	reader := conn.Reader()
	echoMsg, _ := reader.ReadString(len(message))
	fmt.Printf("[recv msg] %v\n", echoMsg)
}
