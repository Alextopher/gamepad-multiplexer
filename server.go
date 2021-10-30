package main

import (
	"fmt"
	"github.com/go-gl/glfw/v3.3/glfw"
	"log"
	"net"
)

func controlError(conn net.Conn, msg string) {
	errMsg := (&ControlProtocol{}).Error(msg)
	conn.Write(errMsg)
	conn.Close()
}

func (c *ServerConn) Handshake() error {
	// Create a buffer
	buf := make([]byte, 1024)

	// Read in data
	_, err := c.Conn.Read(buf)
	if err != nil {
		log.Printf("Failed to read from client %s with error: %s", c.Conn.RemoteAddr().String(),
			err.Error())
		return err
	}

	pkt := &ControlProtocol{}
	err = pkt.Parse(buf)
	if err != nil {
		controlError(c.Conn, "Invalid packet, expecting type REGISTER followed by a name")
		return err
	}
	if pkt.Type != REGISTER {
		controlError(c.Conn, "Invalid packet, expecting type REGISTER followed by a name")
		return err
	}

	// Get the client name
	name := string(pkt.Data)

	// See if it's a valid name
	if !namePattern.Match(pkt.Data) {
		// Invalid name, tell them that and die
		controlError(c.Conn, "Invalid name")
		return err
	}

	// Loop through clients to see if this name already exists
	// This is only done on connect so that the map will mostly be used efficiently by id
	// We just haven't gotten to this step yet, so we can't do that
	clientLock.Lock()
	for _, client := range clients {
		// Kill the connection if the name already exists
		if client.Name == name {
			controlError(c.Conn, "Name already taken, please try something else")
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
	_, err = c.Conn.Write(pkt.SetId(c.Id))
	if err != nil {
		log.Printf(
			"Could not send packet to client %s due to error: %s\n",
			c.Conn.RemoteAddr().String(),
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
		controlError(c.Conn, "Configuration doesn't exist for name "+name)
		clientLock.Lock()
		delete(clients, c.Id)
		clientLock.Unlock()
		return err
	}

	// Send over the rules
	conf := joystickRules.Bytes()
	_, err = c.Conn.Write(pkt.Configure(conf))

	// Complain on error
	if err != nil {
		log.Printf(
			"Could not send packet to client %s due to error: %s\n",
			c.Conn.RemoteAddr().String(),
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
func (c *ServerConn) ControlSocket(port uint16) {
	// Handle the handshake
	err := c.Handshake()
	if err != nil {
		return
	}

	pkt := &ControlProtocol{}
	// Wait for joystick peripheral announcements
	for {
		buf := make([]byte, 512)
		_, err := c.Conn.Read(buf)
		if err != nil {
			log.Printf(
				"Could not read packet from client %s due to error: %s\n",
				c.Conn.RemoteAddr().String(),
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
			controlError(c.Conn, "Invalid packet")
			continue
		}

		if pkt.Type == PERIPHERAL_CONNECT || pkt.Type == PERIPHERAL_DISCONNECT {
			// Print out what the peripheral event was
			log.Println(string(pkt.Data))
		} else if pkt.Type == DONE {
			// Close the connection, the client said they're done
			c.Conn.Close()
			clientLock.Lock()
			delete(clients, c.Id)
			clientLock.Unlock()
			return
		}
	}
}

func udpListener(serv net.PacketConn) {
	counter := make(map[string]uint32)
	// Make a buffer for the size of the packet we expect
	for {
		buf := make([]byte, GamestatePacketLen)
		// Read in the data
		_, raddr, err := serv.ReadFrom(buf)
		if err != nil {
			log.Fatalf("Failed to read from %s socket due to error: %s", raddr.Network(), err)
		}

		ip := raddr.String()

		// Parse the packet
		pkt := &GamestateProtocol{}
		err = pkt.Parse(buf)

		// If the packet is bad or old just ignore it
		if err != nil {
			continue
		}

		// TODO actually check to make sure this is a real client
		// See if this client doesn't exist or it's an old packet
		// if c, exists := counter[ip]; !exists || pkt.PacketId < c {
		// 	continue
		// }

		// Make sure the packet isn't old
		if pkt.PacketId < counter[ip] {
			continue
		}

		// Up the packet counter
		counter[ip] += 1

		// TODO Check client rules to validate client
		// Multiplex the rules
		gamestateLock.Lock()
		gamepadStates[glfw.Joystick(pkt.JoystickId)] = pkt.GamepadState
		gamestateLock.Unlock()
	}
}

func listen(host string, port uint16, rules ClientsMap) {
	// Create the global UDP listener to handle all clients
	udpServ, err := net.ListenPacket("udp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Fatalf("Failed to open socket %s:%d due to error: %s\n", host, port, err.Error())
	}

	// Handle UDP connections
	go udpListener(udpServ)

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
		client := ServerConn{
			Conn:  conn,
			Rules: rules,
		}
		// Handle the controlSocket and die if it's bad
		go client.ControlSocket(port)
	}
}
