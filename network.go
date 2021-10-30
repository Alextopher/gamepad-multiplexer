package main

import (
	"net"
	"sync"
)

var clients map[uint8]*ServerConn = make(map[uint8]*ServerConn)
var packetConnections map[string]*ServerConn = make(map[string]*ServerConn)

var clientLock *sync.Mutex = &sync.Mutex{}

type ClientConn struct {
	Id           uint8
	Name         string
	ControlConn  net.Conn
	DatagramConn net.Conn
	Rules        RulesMap
}

type ServerConn struct {
	Id    uint8
	Name  string
	Conn  net.Conn
	Rules ClientsMap
}
