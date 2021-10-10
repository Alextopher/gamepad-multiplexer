package main

import (
	"fmt"
	"log"
	"net"
)

// TODO finish this
// Handle the controlSocket between the client and server
func controlSocket(conn net.Conn) {
	// Create a buffer
	buf := make([]byte, 1024)

	// Read in data
	_, err := conn.Read(buf)
	if err != nil {
		log.Printf("Failed to read from client %s with error: %s", conn.RemoteAddr().String(), err.Error())
		return
	}

	// See if it's a valid name
	if !namePattern.Match(buf) {
		// Invalid name, tell them that and die
		err := (&ControlPacket{}).Error("Invalid name")
		conn.Write(err)
		conn.Close()
		return
	}
}

func listen(host string, port uint16) {
	// Create the counter map
	counter := make(map[string]uint32)

	// // Create the TCP listener to make the controlSocket with all clients
	// serv, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	// if err != nil {
	// 	log.Fatalf("Failed to open socket %s:%d due to error: %s\n", host, port, err.Error())
	// }
	//
	// for {
	// 	// Accept connections
	// 	conn, err := serv.Accept()
	// 	if err != nil {
	// 		log.Fatalln("Failed to accept client due to error:", err)
	// 	}
	// 	// Handle the controlSocket and die if it's bad
	// 	go controlSocket(conn)
	// }

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
		buf := make([]byte, 31)

		// Read in the data from the packet
		_, raddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Println(err)
		}

		// Get the IP of the remote host
		ip := raddr.IP.String()

		pkt := GamestatePacket{}
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
