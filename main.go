package main

import (
	"fmt"
	"github.com/manzil-infinity180/dns-server-resolver/pkg/dns"
	"net"
)

func main() {

	fmt.Printf("Starting DNS Server...\n")
	packetConn, err := net.ListenPacket("udp", ":53")
	if err != nil {
		panic(err)
	}
	defer packetConn.Close()

	for {
		buf := make([]byte, 512)
		bytesRead, addr, err := packetConn.ReadFrom(buf)
		if err != nil {
			fmt.Printf("read error from %s: %s", addr.String(), err)
			continue
		}
		go dns.HandlePacket(packetConn, addr, buf[:bytesRead])
	}

}
