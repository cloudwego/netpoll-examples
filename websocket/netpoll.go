package main

import (
	"context"
	"net"

	"log"

	"github.com/cloudwego/netpoll"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

var (
	_ netpoll.OnPrepare = prepare
	_ netpoll.OnConnect = connect
	_ netpoll.OnRequest = handle
)

func prepare(conn netpoll.Connection) context.Context {
	log.Printf("prepare")
	return context.Background()
}

func connect(ctx context.Context, conn netpoll.Connection) context.Context {
	log.Printf("connect")
	_, err := ws.Upgrade(conn)
	if err != nil {
		log.Printf("upgrade error: %v", err)
		conn.Close()
		return ctx
	}
	log.Printf("%s: established websocket connection", nameConn(conn))
	return ctx
}

func nameConn(conn net.Conn) string {
	return conn.LocalAddr().String() + " > " + conn.RemoteAddr().String()
}

func handle(ctx context.Context, conn netpoll.Connection) error {
	header, err := ws.ReadHeader(conn)
	if err != nil {
		log.Printf("server ReadHeader failed,err=%+v", err)
		return err
	}
	log.Printf("server recv header is %+v", header)

	if header.OpCode.IsControl() {
		if err := wsutil.ControlFrameHandler(conn, ws.StateServerSide)(header, conn); err != nil {
			log.Printf("server ControlFrameHandler err is %+v", err)
		}
		return nil
	}

	msg, err := conn.Reader().Next(int(header.Length))
	if err != nil {
		log.Printf("Reader().Next failed,err=%+v", err)
		return err
	}
	if header.Masked {
		ws.Cipher(msg, header.Mask, 0)
	}

	if err := wsutil.WriteServerText(conn, msg); err != nil {
		log.Printf("WriteServerText failed,err=%+v", err)
		conn.Close()
		return nil
	}
	return nil
}
