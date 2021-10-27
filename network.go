package main

import (
	"net"
	"sync"
)

var clients map[uint8]*Connection = make(map[uint8]*Connection)
var clientLock *sync.Mutex = &sync.Mutex{}

type Connection struct {
	Id      uint8
	Name    string
	TCPConn net.Conn
	UDPConn *net.UDPConn
	Rules   ClientsMap
}
