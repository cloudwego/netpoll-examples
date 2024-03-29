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
	network, address := "tcp", ":8082"
	listener, _ := netpoll.CreateListener(network, address)
	eventLoop, _ := netpoll.NewEventLoop(
		nil,
		netpoll.WithOnConnect(onConnect),
	)

	// start listen loop ...
	eventLoop.Serve(listener)
}

var _ netpoll.OnConnect = onConnect

func onConnect(ctx context.Context, connection netpoll.Connection) context.Context {
	go func() {
		for range time.Tick(time.Second) {
			connection.Writer().WriteString(fmt.Sprintf("%s\n", time.Now().Format(time.RFC3339)))
			connection.Writer().Flush()
		}
	}()
	return context.Background()
}
