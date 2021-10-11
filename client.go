package main

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"log"
	"net"
)

func connect(host string, port uint16, name string) (*net.TCPConn, *net.UDPConn, RulesMap,
	uint8) {
	tcpRaddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Fatalln("Failed to resolve addr with err:", err)
	}

	// Connect to the server
	tcpConn, err := net.DialTCP("tcp", nil, tcpRaddr)
	if err != nil {
		log.Fatalln("Failed to connect with err:", err)
	}

	// Do the handshake to get the id and config
	rules, id, err := handshake(tcpConn, name)
	if err != nil {
		log.Fatalln("Handshake failed due to error:", err)
	}

	// Spin off a UDP connection to the same socket
	udpRaddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Fatalln("Failed to resolve addr with err:", err)
	}

	// Connect to the server
	udpConn, err := net.DialUDP("udp", nil, udpRaddr)
	if err != nil {
		log.Fatalln("Failed to connect with err:", err)
	}

	return tcpConn, udpConn, rules, id
}

func handshake(conn *net.TCPConn, name string) (rules RulesMap, id uint8, err error) {
	pkt := &ControlProtocol{}

	// Register a name
	_, err = conn.Write(pkt.Register(name))
	if err != nil {
		return rules, id, err
	}

	// Read in the next packet
	buf := make([]byte, 512)
	_, err = conn.Read(buf)
	if err != nil {
		return rules, id, err
	}

	// See if it's an ID
	err = pkt.Parse(buf)
	if err != nil {
		return rules, id, err
	}
	if pkt.Type == SET_ID {
		// Get the id
		id = pkt.Data[0]
	} else if pkt.Type == ERROR {
		// Close the connection since this is wonky
		conn.Close()
		return rules, id, errors.New(string(pkt.Data))
	} else {
		// Close the connection since this is wonky
		conn.Close()
		return rules, id, errors.New("server response was invalid, aborting connection")
	}

	// Get the next packet
	buf = make([]byte, 4096)
	_, err = conn.Read(buf)
	if err != nil {
		return rules, id, err
	}

	// See if it's a configuration
	err = pkt.Parse(buf)
	if err != nil {
		return rules, id, err
	}
	if pkt.Type == CONFIGURATION {
		err = yaml.Unmarshal(pkt.Data, &rules)
		if err != nil {
			return rules, id, err
		}
	} else if pkt.Type == ERROR {
		// Close the connection since this is wonky
		conn.Close()
		return rules, id, errors.New(string(pkt.Data))
	} else {
		// Close the connection since this is wonky
		conn.Close()
		return rules, 0, errors.New("server response was invalid, aborting connection")
	}

	// Handshake is complete
	return rules, id, nil
}
