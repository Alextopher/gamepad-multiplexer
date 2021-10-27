package main

import (
	"errors"
	"fmt"
	"log"
	"net"
)

func connect(host string, port uint16, name string) (conn *Connection) {
	conn = &Connection{
		Name:    name,
		TCPConn: nil,
		UDPConn: nil,
		Rules:   make(map[string]RulesMap),
	}
	tcpRaddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Fatalln("Failed to resolve addr with err:", err)
	}

	// Connect to the server
	conn.TCPConn, err = net.DialTCP("tcp", nil, tcpRaddr)
	if err != nil {
		log.Fatalln("Failed to connect with err:", err)
	}

	// Do the handshake to get the id and config
	err = conn.ClientHandshake()
	if err != nil {
		log.Fatalln("Handshake failed due to error:", err)
	}

	// Spin off a UDP connection to the same socket
	udpRaddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Fatalln("Failed to resolve addr with err:", err)
	}

	// Connect to the server
	conn.UDPConn, err = net.DialUDP("udp", nil, udpRaddr)
	if err != nil {
		log.Fatalln("Failed to connect with err:", err)
	}

	return conn
}

func (c *Connection) ClientHandshake() error {
	pkt := &ControlProtocol{}

	// Register a name
	_, err := c.TCPConn.Write(pkt.Register(c.Name))
	if err != nil {
		return err
	}

	// Read in the next packet
	buf := make([]byte, 512)
	_, err = c.TCPConn.Read(buf)
	if err != nil {
		return err
	}

	// See if it's an ID
	err = pkt.Parse(buf)
	if err != nil {
		return err
	}

	if pkt.Type == SET_ID {
		// Get the id
		c.Id = pkt.Data[0]
	} else if pkt.Type == ERROR {
		// Close the connection since this is wonky
		c.TCPConn.Close()
		return errors.New(string(pkt.Data))
	} else {
		// Close the connection since this is wonky
		c.TCPConn.Close()
		return errors.New("server response was invalid, aborting connection")
	}

	// Get the next packet
	buf = make([]byte, 4096)
	_, err = c.TCPConn.Read(buf)
	if err != nil {
		return err
	}

	// See if it's a configuration
	err = pkt.Parse(buf)
	if err != nil {
		return err
	}

	if pkt.Type == CONFIGURATION {
		c.Rules[c.Name], err = ParseRulesMap(pkt.Data)
		if err != nil {
			return err
		}
	} else if pkt.Type == ERROR {
		// Close the connection since this is wonky
		c.TCPConn.Close()
		return errors.New(string(pkt.Data))
	} else {
		// Close the connection since this is wonky
		c.TCPConn.Close()
		return errors.New("server response was invalid, aborting connection")
	}

	// Handshake is complete
	return nil
}
