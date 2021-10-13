package main

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/go-gl/glfw/v3.3/glfw"
)

var clients map[uint8]Client = make(map[uint8]Client)
var clientLock *sync.Mutex = &sync.Mutex{}

type Client struct {
	Id   uint8
	Name string
}

func controlError(conn net.Conn, msg string) {
	errMsg := (&ControlProtocol{}).Error(msg)
	conn.Write(errMsg)
	conn.Close()
}

// TODO finish this
// Handle the controlSocket between the client and server
func controlSocket(conn net.Conn, rules ClientsMap) {
	// Create a buffer
	buf := make([]byte, 1024)

	// Read in data
	_, err := conn.Read(buf)
	if err != nil {
		log.Printf("Failed to read from client %s with error: %s", conn.RemoteAddr().String(), err.Error())
		return
	}

	pkt := &ControlProtocol{}
	err = pkt.Parse(buf)
	if err != nil {
		controlError(conn, "Invalid packet, expecting type REGISTER followed by a name")
		return
	}
	if pkt.Type != REGISTER {
		controlError(conn, "Invalid packet, expecting type REGISTER followed by a name")
		return
	}

	// Get the client name
	name := string(pkt.Data)

	// See if it's a valid name
	if !namePattern.Match(pkt.Data) {
		// Invalid name, tell them that and die
		controlError(conn, "Invalid name")
		return
	}

	// Loop through clients to see if this name already exists
	// This is only done on connect so that the map will mostly be used efficiently by id
	// We just haven't gotten to this step yet, so we can't do that
	clientLock.Lock()
	for _, client := range clients {
		// Kill the connection if the name already exists
		if client.Name == name {
			controlError(conn, "Name already taken, please try something else")
			// Remember to unlock
			clientLock.Unlock()
			return
		}
	}

	// Now try to find a new valid id
	var newId uint8
	for newId = 0; newId < 255; newId++ {
		_, exists := clients[newId]
		if !exists {
			break
		}
	}

	// Create the client in the map
	clients[newId] = Client{
		Id:   newId,
		Name: name,
	}

	// Unlock since we are done with the map
	clientLock.Unlock()

	// Tell the client of their id
	_, err = conn.Write(pkt.SetId(newId))
	if err != nil {
		log.Printf(
			"Could not send packet to client %s due to error: %s\n",
			conn.RemoteAddr().String(),
			err.Error(),
		)
		clientLock.Lock()
		delete(clients, newId)
		clientLock.Unlock()
		return
	}

	// Get the client configuration
	joystickRules, exists := rules[name]
	if !exists {
		// Tell the client they don't have a configuration
		controlError(conn, "Configuration doesn't exist for name "+name)
		clientLock.Lock()
		delete(clients, newId)
		clientLock.Unlock()
		return
	}

	// Send over the rules
	conf := joystickRules.Bytes()
	_, err = conn.Write(pkt.Configure(conf))

	// Complain on error
	if err != nil {
		log.Printf(
			"Could not send packet to client %s due to error: %s\n",
			conn.RemoteAddr().String(),
			err.Error(),
		)
		clientLock.Lock()
		delete(clients, newId)
		clientLock.Unlock()
		return
	}

	// Handshake is complete,
	// spin off udp server for this client
	go joystickHandler(conn.RemoteAddr().String())

	// Wait for joystick peripheral announcements
	for {
		buf := make([]byte, 512)
		_, err := conn.Read(buf)
		if err != nil {
			log.Printf(
				"Could not read packet from client %s due to error: %s\n",
				conn.RemoteAddr().String(),
				err.Error(),
			)
			clientLock.Lock()
			delete(clients, newId)
			clientLock.Unlock()
			return
		}

		// Parse the packet and complain on failure, but don't close the connection
		err = pkt.Parse(buf)
		if err != nil {
			controlError(conn, "Invalid packet")
			continue
		}

		if pkt.Type == PERIPHERAL_CONNECT || pkt.Type == PERIPHERAL_DISCONNECT {
			// Print out what the peripheral event was
			log.Println(string(pkt.Data))
		} else if pkt.Type == DONE {
			// Close the connection, the client said they're done
			conn.Close()
			clientLock.Lock()
			delete(clients, newId)
			clientLock.Unlock()
			return
		}
	}

}

func listen(host string, port uint16, rules ClientsMap) {
	// Create the TCP listener to make the controlSocket with all clients
	serv, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Fatalf("Failed to open socket %s:%d due to error: %s\n", host, port, err.Error())
	}

	for {
		// Accept connections
		conn, err := serv.Accept()
		if err != nil {
			log.Fatalln("Failed to accept client due to error:", err)
		}
		// Handle the controlSocket and die if it's bad
		go controlSocket(conn, rules)
	}
}

func joystickHandler(socket string) {
	// Create the counter map
	counter := make(map[string]uint32)

	// Create the listener
	servAddr, err := net.ResolveUDPAddr("udp", socket)
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

		pkt := GamestateProtocol{}
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
		gamepadStates[glfw.Joystick(pkt.JoystickId)] = pkt.GamepadState
	}
}
