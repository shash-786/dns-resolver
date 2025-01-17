package main

import (
	"fmt"
	"log"
	"net"

	my_dns "githib.com/shash-786/dns-resolver/pkg/dns"
)

func main() {
	fmt.Println("Starting DNS Server!!")
	p, err := net.ListenPacket("udp", ":53")
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()

	for {
		buf := make([]byte, 512)
		n, addr, err := p.ReadFrom(buf)
		if err != nil {
			fmt.Printf("conn error for addr %s\n%v", addr.String(), err)
			continue
		}
		fmt.Printf("recived %s from %s", buf[:n], addr.String())
		go my_dns.HandlePacket(p, addr, buf[:n])
	}
}
