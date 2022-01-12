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
	"log"
	"time"

	"github.com/cloudwego/netpoll"
)

var (
	downstreamAddr = "127.0.0.1:8080"
	downstreamKey  = "downstream"
)

func main() {
	network, address := "tcp", ":8081"
	listener, _ := netpoll.CreateListener(network, address)
	eventLoop, _ := netpoll.NewEventLoop(
		onRequest,
		netpoll.WithOnConnect(onConnect),
		netpoll.WithReadTimeout(time.Second),
	)

	// start listen loop ...
	eventLoop.Serve(listener)
}

var _ netpoll.OnConnect = onConnect
var _ netpoll.OnRequest = onRequest

func onConnect(ctx context.Context, upstream netpoll.Connection) context.Context {
	downstream, err := netpoll.DialConnection("tcp", downstreamAddr, time.Second)
	if err != nil {
		log.Printf("connect downstream failed: %v", err)
	}
	return context.WithValue(ctx, downstreamKey, downstream)
}

func onRequest(ctx context.Context, upstream netpoll.Connection) error {
	// read request
	req, _ := upstream.Reader().ReadString(upstream.Reader().Len())

	// send request to downstream
	downstream := ctx.Value(downstreamKey).(netpoll.Connection)
	_, _ = downstream.Writer().WriteString(req)
	downstream.Writer().Flush()

	// receive response from downstream
	resp, _ := downstream.Reader().ReadString(len(req))

	// send response to upstream
	upstream.Writer().WriteString(resp)
	upstream.Writer().Flush()
	return nil
}
