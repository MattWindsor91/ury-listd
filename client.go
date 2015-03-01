package main

import (
	"bufio"
	"net"

	baps3 "github.com/UniversityRadioYork/baps3-go"
)

type Client struct {
	conn     net.Conn
	reader   *bufio.Reader
	writer   *bufio.Writer
	serverCh chan []byte
	Outgoing chan []byte
}

func MakeClient(conn net.Conn, clientCh chan []byte) *Client {
	client := &Client{
		conn:     conn,
		reader:   bufio.NewReader(conn),
		writer:   bufio.NewWriter(conn),
		serverCh: clientCh,
		Outgoing: make(chan []byte),
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
		c.serverCh <- data // Each client doesn't care what it got, that's for the server to handle
	}
}

func (c *Client) write() {
	for data := range c.Outgoing {
		_, err := c.writer.Write(data)
		if err != nil {
			continue // TODO Handle
		}
		c.writer.Flush()
	}
}
