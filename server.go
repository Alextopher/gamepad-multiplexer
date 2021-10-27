package main

import (
	"fmt"
	"github.com/go-gl/glfw/v3.3/glfw"
	"log"
	"net"
	"strings"
)

func controlError(conn net.Conn, msg string) {
	errMsg := (&ControlProtocol{}).Error(msg)
	conn.Write(errMsg)
	conn.Close()
}

func (c *Connection) ServerHandshake() error {
	// Create a buffer
	buf := make([]byte, 1024)

	// Read in data
	_, err := c.TCPConn.Read(buf)
	if err != nil {
		log.Printf("Failed to read from client %s with error: %s", c.TCPConn.RemoteAddr().String(), err.Error())
		return err
	}

	pkt := &ControlProtocol{}
	err = pkt.Parse(buf)
	if err != nil {
		controlError(c.TCPConn, "Invalid packet, expecting type REGISTER followed by a name")
		return err
	}
	if pkt.Type != REGISTER {
		controlError(c.TCPConn, "Invalid packet, expecting type REGISTER followed by a name")
		return err
	}

	// Get the client name
	name := string(pkt.Data)

	// See if it's a valid name
	if !namePattern.Match(pkt.Data) {
		// Invalid name, tell them that and die
		controlError(c.TCPConn, "Invalid name")
		return err
	}

	// Loop through clients to see if this name already exists
	// This is only done on connect so that the map will mostly be used efficiently by id
	// We just haven't gotten to this step yet, so we can't do that
	clientLock.Lock()
	for _, client := range clients {
		// Kill the connection if the name already exists
		if client.Name == name {
			controlError(c.TCPConn, "Name already taken, please try something else")
			// Remember to unlock
			clientLock.Unlock()
			return err
		}
	}

	// Now try to find a new valid id
	for c.Id = 0; c.Id < 255; c.Id++ {
		_, exists := clients[c.Id]
		if !exists {
			break
		}
	}

	// Create the client in the map
	clients[c.Id] = c

	// Unlock since we are done with the map
	clientLock.Unlock()

	// Tell the client of their id
	_, err = c.TCPConn.Write(pkt.SetId(c.Id))
	if err != nil {
		log.Printf(
			"Could not send packet to client %s due to error: %s\n",
			c.TCPConn.RemoteAddr().String(),
			err.Error(),
		)
		clientLock.Lock()
		delete(clients, c.Id)
		clientLock.Unlock()
		return err
	}

	// Get the client configuration
	joystickRules, exists := c.Rules[name]
	if !exists {
		// Tell the client they don't have a configuration
		controlError(c.TCPConn, "Configuration doesn't exist for name "+name)
		clientLock.Lock()
		delete(clients, c.Id)
		clientLock.Unlock()
		return err
	}

	// Send over the rules
	conf := joystickRules.Bytes()
	_, err = c.TCPConn.Write(pkt.Configure(conf))

	// Complain on error
	if err != nil {
		log.Printf(
			"Could not send packet to client %s due to error: %s\n",
			c.TCPConn.RemoteAddr().String(),
			err.Error(),
		)
		clientLock.Lock()
		delete(clients, c.Id)
		clientLock.Unlock()
		return err
	}
	return nil
}

// ControlSocket handles the handshake with the client and continues
// to listen for messages to the server
func (c *Connection) ControlSocket(port uint16) {
	// Handle the handshake
	err := c.ServerHandshake()
	if err != nil {
		return
	}

	// Spin off udp server for this client
	go joystickHandler(c, port)

	pkt := &ControlProtocol{}
	// Wait for joystick peripheral announcements
	for {
		buf := make([]byte, 512)
		_, err := c.TCPConn.Read(buf)
		if err != nil {
			log.Printf(
				"Could not read packet from client %s due to error: %s\n",
				c.TCPConn.RemoteAddr().String(),
				err.Error(),
			)
			clientLock.Lock()
			delete(clients, c.Id)
			clientLock.Unlock()
			return
		}

		// Parse the packet and complain on failure, but don't close the connection
		err = pkt.Parse(buf)
		if err != nil {
			controlError(c.TCPConn, "Invalid packet")
			continue
		}

		if pkt.Type == PERIPHERAL_CONNECT || pkt.Type == PERIPHERAL_DISCONNECT {
			// Print out what the peripheral event was
			log.Println(string(pkt.Data))
		} else if pkt.Type == DONE {
			// Close the connection, the client said they're done
			c.TCPConn.Close()
			clientLock.Lock()
			delete(clients, c.Id)
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
		// Create the client
		client := Connection{
			TCPConn: conn,
			Rules:   rules,
		}
		// Handle the controlSocket and die if it's bad
		go client.ControlSocket(port)
	}
}

func joystickHandler(conn *Connection, port uint16) {
	// Create the counter map
	counter := make(map[string]uint32)

	// Create the listener
	ip := strings.Split(conn.TCPConn.RemoteAddr().String(), ":")[0]
	servAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		log.Fatalln("Could not create server due to error:", err)
	}
	conn.UDPConn, err = net.ListenUDP("udp", servAddr)
	if err != nil {
		log.Fatalln("Could not create server due to error:", err)
	}

	// Accept connections as they come in
	for {
		// Create a buffer
		buf := make([]byte, 31)

		// Read in the data from the packet
		_, raddr, err := conn.UDPConn.ReadFromUDP(buf)
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
