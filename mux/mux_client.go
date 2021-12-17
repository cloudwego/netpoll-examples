/*
 * Copyright 2021 CloudWeGo Authors
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
	"encoding/binary"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/cloudwego/netpoll"
	"github.com/cloudwego/netpoll/mux"

	"github.com/cloudwego/netpoll-examples/codec"
)

func main() {
	network, address := "tcp", "127.0.0.1:8080"

	// new mux client
	cli := newMuxClient(network, address, 4)
	req := &codec.Message{
		Message: "hello world",
	}
	resp, err := cli.Echo(req)
	if err != nil {
		panic(err)
	}
	println(resp)
}

func newMuxClient(network, address string, size int) *muxclient {
	cli := &muxclient{}
	cli.size = uint64(size)
	cli.conns = make([]*cliMuxConn, size)
	for i := range cli.conns {
		cn, err := netpoll.DialConnection(network, address, time.Second)
		if err != nil {
			panic(fmt.Errorf("mux dial conn failed: %s", err.Error()))
		}
		mc := newCliMuxConn(cn)
		cli.conns[i] = mc
	}
	return cli
}

type muxclient struct {
	conns  []*cliMuxConn
	size   uint64
	cursor uint64
}

func (cli *muxclient) Echo(req *codec.Message) (resp *codec.Message, err error) {
	// get conn & codec
	mc := cli.conns[atomic.AddUint64(&cli.cursor, 1)%cli.size]

	// encode
	writer := netpoll.NewLinkBuffer()
	err = codec.Encode(writer, req)
	if err != nil {
		return nil, err
	}
	mc.Put(func() (buf netpoll.Writer, isNil bool) {
		return writer, false
	})

	// decode
	reader := <-mc.rch
	resp = &codec.Message{}
	err = codec.Decode(reader, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func newCliMuxConn(conn netpoll.Connection) *cliMuxConn {
	mc := &cliMuxConn{}
	mc.conn = conn
	mc.rch = make(chan netpoll.Reader)
	// loop read
	conn.SetOnRequest(func(ctx context.Context, connection netpoll.Connection) error {
		reader := connection.Reader()
		// decode
		bLen, err := reader.Peek(4)
		if err != nil {
			return err
		}
		l := int(binary.BigEndian.Uint32(bLen))
		r, _ := reader.Slice(l + 4)
		mc.rch <- r
		return nil
	})
	// loop write
	mc.wqueue = mux.NewShardQueue(mux.ShardSize, conn)
	return mc
}

type cliMuxConn struct {
	conn   netpoll.Connection
	rch    chan netpoll.Reader
	wqueue *mux.ShardQueue // use for write

}

// Put puts the buffer getter back to the queue.
func (c *cliMuxConn) Put(gt mux.WriterGetter) {
	c.wqueue.Add(gt)
}
