package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

var connectionCounter uint64
var targetHost string
var dialTimeout time.Duration

// ActiveConnectionCounter keeps track of the number of active connections
type ActiveConnectionCounter struct {
	count                  uint64
	countMu                sync.Mutex
	printActiveConnections bool
}

// Increment increments the number of active connections by one
func (c *ActiveConnectionCounter) Increment() {
	c.countMu.Lock()
	defer c.countMu.Unlock()
	c.count++
	c.printActives()
}

// Get returns the number of active connections
func (c *ActiveConnectionCounter) Get() uint64 {
	c.countMu.Lock()
	defer c.countMu.Unlock()
	return c.count
}

// Decrement decrements the number of active connections by one
func (c *ActiveConnectionCounter) Decrement() {
	c.countMu.Lock()
	defer c.countMu.Unlock()
	c.count--
	c.printActives()
}

// printActives prints the number of active connections if the flag is set
func (c *ActiveConnectionCounter) printActives() {
	if c.printActiveConnections {
		log.Printf("Active connections: %d", c.count)
	}
}

// handleConnection creates a new connection to the target host and relays the data between the two connections.
// client is the connection incoming connection from the client, connectionId is the id of the connection
func handleConnection(client net.Conn, connectionId uint64, counter *ActiveConnectionCounter) {
	defer counter.Decrement()
	defer client.Close()

	server, err := net.DialTimeout("tcp", targetHost, dialTimeout)
	defer server.Close()

	if err != nil {
		log.Printf("Connection %d | Cannot connect to target host %s due to error %s", connectionId, targetHost, err.Error())
		return
	}

	// wait for relaying to finish
	wg := sync.WaitGroup{}

	// start relaying from client to server
	wg.Add(1)
	go func() {
		written, err := io.Copy(server, client)
		log.Printf("Connection %d | Copied %d bytes from client to server", connectionId, written)
		if err != nil {
			log.Printf("Connection %d | Error while copying data from client to server: %s", connectionId, err.Error())
		}
		wg.Done()
	}()

	// start relaying from server to client
	wg.Add(1)
	go func() {
		written, err := io.Copy(client, server)
		log.Printf("Connection %d | Copied %d bytes from server to client", connectionId, written)
		if err != nil {
			log.Printf("Connection %d | Error while copying data from server to client: %s", connectionId, err.Error())
		}
		wg.Done()
	}()

	wg.Wait()
}

// startProxy starts the proxy server
func startProxy(listenAddress string, showActiveConnections bool) {
	ln, err := net.Listen("tcp", listenAddress)
	if err != nil {
		log.Fatalln("Cannot start proxy:", err)
	}

	log.Printf("Listening on %s\n", listenAddress)

	activeCounter := &ActiveConnectionCounter{printActiveConnections: showActiveConnections}

	for {
		conn, err := ln.Accept()
		connectionCounter++
		activeCounter.Increment()

		log.Printf("Got new connection with id %d from %s", connectionCounter, conn.RemoteAddr().String())

		if err != nil {
			log.Printf("Connection %d | Cannot accept connection due to error %s", connectionCounter, err.Error())
			conn.Close()
			activeCounter.Decrement()
			continue
		}

		go handleConnection(conn, connectionCounter, activeCounter)
	}
}

// usage prints the command line usage.
func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [OPTIONS] [bind_address]:bind_port [target_address]:target_port\n", os.Args[0])
	fmt.Fprintf(flag.CommandLine.Output(), "\nDescription:\n")
	fmt.Fprintf(flag.CommandLine.Output(), `Proxy relays TCP connections. For every incoming
connection, a new connection is established to the target 
host and the data is relayed between the two connections.`)
	fmt.Fprintf(flag.CommandLine.Output(), "\n\nOPTIONS:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage

	timeout := flag.Uint("timeout", 10, "Timeout in seconds for dialing to target host.")
	showActive := flag.Bool("show-active", true, "When true, prints the number of active connections.")

	flag.Parse()

	dialTimeout = time.Second * time.Duration(*timeout)

	tail := flag.Args()

	if len(tail) != 2 {
		flag.Usage()
		os.Exit(1)
	}

	listenAddress := tail[0]
	targetHost = tail[1]

	startProxy(listenAddress, *showActive)
}
