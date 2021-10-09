package main

import (
	"fmt"
	"log"
	"net"
)

var counter map[string]uint32

func listen(host string, port uint16) {
	// Create the counter map
	counter = make(map[string]uint32)

	// Create the listener
	servAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Fatalln("Could not create server due to error:", err)
	}
	conn, err := net.ListenUDP("udp", servAddr)
	if err != nil {
		log.Fatalln("Could not create server due to error:", err)
	}

	// Accept connections as they come in
	for {
		// Create a buffer
		buf := make([]byte, 128)
		// Read in the data from the packet
		_, raddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Println(err)
		}

		// Get the IP of the remote host
		ip := raddr.IP.String()

		pkt := Packet{}
		err = pkt.Parse(buf)
		if err != nil {
			log.Printf("Bad packet received from %s with contents '%x'", ip, buf)
		}

		// See if a counter exists for this ip
		if count, exists := counter[ip]; exists {
			// Bump up the packet id since this is newer, drop the packet otherwise
			if count < pkt.PacketId {
				count = pkt.PacketId
			}
		} else {
			counter[ip] = pkt.PacketId
		}

		// Update the game states in the map
		// Possibly unsafe?
		gamepadStates[pkt.JoystickId] = pkt.GamepadState
	}
}
