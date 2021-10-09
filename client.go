package main

import (
	"log"
	"net"
)

func connect(addr string) *net.UDPConn {
	raddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Fatalln("Failed to resolve addr with err:", err)
	}

	laddr, err := net.ResolveUDPAddr("udp", "localhost:0")
	if err != nil {
		log.Fatalln("Failed to resolve addr with err:", err)
	}

	conn, err := net.DialUDP("udp", laddr, raddr)
	if err != nil {
		log.Fatalln("Failed to connect with err:", err)
	}
	return conn
}

func send(conn *net.UDPConn, pkt Packet) error {
	_, err := conn.WriteToUDP(pkt.Bytes(), pkt.RAddr)
	return err
}
