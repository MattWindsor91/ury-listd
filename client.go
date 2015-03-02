package main

import (
	"bufio"
	"net"

	baps3 "github.com/UniversityRadioYork/baps3-go"
)

type Client struct {
	conn       net.Conn
	reader     *bufio.Reader
	responseCh chan []byte
	Outgoing   chan []byte
}

func MakeClient(conn net.Conn, respCh chan []byte) *Client {
	client := &Client{
		conn:       conn,
		reader:     bufio.NewReader(conn),
		responseCh: respCh,
		Outgoing:   make(chan []byte),
	}

	go client.Read()
	go client.write()

	ohai := MakeWelcomeMsg()
	data, _ := ohai.Pack()
	client.Outgoing <- data

	return client
}

func MakeWelcomeMsg() *baps3.Message {
	return baps3.NewMessage(baps3.RsOhai).AddArg("listd").AddArg("0.0")
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) Read() {
	for {
		data, _ := c.reader.ReadBytes('\n')
		c.responseCh <- data // Each client doesn't care what it got, that's for the server to handle
	}
}

func (c *Client) write() {
	for data := range c.Outgoing {
		_, err := c.conn.Write(data)
		if err != nil {
			continue // TODO Handle
		}
	}
}
