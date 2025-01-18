package my_dns

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net"
	"strings"

	"golang.org/x/net/dns/dnsmessage"
)

const ROOT_SERVERS string = "198.41.0.4,199.9.14.201,192.33.4.12,199.7.91.13,192.203.230.10,192.5.5.241,192.112.36.4,198.97.190.53"

func get_root_servers() []net.IP {
	var ip_addies []net.IP
	for _, root := range strings.Split(ROOT_SERVERS, ",") {
		ip_addies = append(ip_addies, net.ParseIP(root))
	}
	return ip_addies
}

func HandlePacket(pc net.PacketConn, addr net.Addr, buf []byte) error {
	return handlePacket(pc, addr, buf)
}

func handlePacket(pc net.PacketConn, addr net.Addr, buf []byte) error {
	return fmt.Errorf("Pass")
}

func outgoingDnsQuery(servers []net.IP, question dnsmessage.Question) (*dnsmessage.Parser, *dnsmessage.Header, error) {
	var (
		max_val                uint16
		header                 dnsmessage.Header
		message                dnsmessage.Message
		resp_ques              []dnsmessage.Question
		packed_message, answer []byte
		queryID                *big.Int
		conn                   net.Conn
		p                      dnsmessage.Parser
		n                      int
		err                    error
	)

	fmt.Printf("New outgoing dns query for %s, servers: %+v\n", question.Name.String(), servers)

	max_val = ^uint16(0)

	queryID, err = rand.Int(rand.Reader, big.NewInt(int64(max_val)))
	if err != nil {
		log.Println("usage: rand.Int()")
		return nil, nil, err
	}

	header = dnsmessage.Header{
		ID:       uint16(queryID.Uint64()),
		Response: false,
		OpCode:   dnsmessage.OpCode(0),
	}

	message = dnsmessage.Message{
		Header:    header,
		Questions: []dnsmessage.Question{question},
	}

	packed_message, err = message.Pack()
	if err != nil {
		log.Println("usage message.Pack()")
		return nil, nil, err
	}

	for _, server := range servers {
		conn, err = net.Dial("udp", server.String())
		if err != nil {
			log.Printf("conn fail for %s", server.String())
			continue
		}
	}

	if conn == nil {
		return nil, nil, fmt.Errorf("no connections found for \n%+v", servers)
	}
	defer conn.Close()

	if n, err = conn.Write(packed_message); err != nil {
		return nil, nil, fmt.Errorf("usage conn.Write() --> %v", err)
	}

	if n != len(packed_message) {
		log.Println("write unsuccessful")
		return nil, nil, nil
	}

	answer = make([]byte, 512)
	if n, err = bufio.NewReader(conn).Read(answer); err != nil {
		return nil, nil, err
	}

	if header, err = p.Start(answer[:n]); err != nil {
		log.Println("usage parser.Start()")
		return nil, nil, err
	}

	if resp_ques, err = p.AllQuestions(); err != nil {
		log.Println("p.AllQuestions() error")
		return nil, nil, err
	}

	if len(message.Questions) != len(resp_ques) {
		return nil, nil, fmt.Errorf("response question length not equal to request length")
	}

	if err = p.SkipAllQuestions(); err != nil {
		return nil, nil, fmt.Errorf("usage p.SkipAllQuestions() error %v", err)
	}
	return &p, &header, nil
}
