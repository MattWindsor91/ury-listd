package main

import (
	"net"
)

type Connection struct {
	net.Conn
}

func NewConnection(uri string) (*Connection, error) {
	conn, err := net.Dial("tcp", uri)
	if err != nil {
		return nil, err
	}
	var obj Connection
	obj.Conn = conn
	return &obj, nil
}
