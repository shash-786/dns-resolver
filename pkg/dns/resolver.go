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

func HandlePacket(pc net.PacketConn, addr net.Addr, buf []byte) {
	if err := handlePacket(pc, addr, buf); err != nil {
		fmt.Printf("HandlePacket error: %s", err)
	}
}

func handlePacket(pc net.PacketConn, addr net.Addr, buf []byte) error {
	var (
		err             error
		p               dnsmessage.Parser
		q               dnsmessage.Question
		response        *dnsmessage.Message
		header          dnsmessage.Header
		packed_response []byte
	)

	if header, err = p.Start(buf); err != nil {
		log.Println("parser.Start() error")
		return err
	}

	if q, err = p.Question(); err != nil {
		log.Println("p.Question() error")
		return err
	}

	if response, err = dnsQuery(get_root_servers(), q); err != nil {
		return err
	}

	response.Header.ID = header.ID
	if packed_response, err = response.Pack(); err != nil {
		return err
	}

	if _, err := pc.WriteTo(packed_response, addr); err != nil {
		return err
	}
	return nil
}

func dnsQuery(servers []net.IP, question dnsmessage.Question) (*dnsmessage.Message, error) {
	for i := 0; i < 3; i++ {
		var (
			p                        *dnsmessage.Parser
			h                        *dnsmessage.Header
			authorities, additionals []dnsmessage.Resource
			err                      error
		)

		if p, h, err = outgoingDnsQuery(servers, question); err != nil {
			fmt.Printf("outgoingDnsQuery() error --> %v", err)
			continue
		}

		if h.Authoritative {
			var answers []dnsmessage.Resource
			if answers, err = p.AllAnswers(); err != nil {
				return &dnsmessage.Message{
					Header: dnsmessage.Header{
						RCode: dnsmessage.RCodeServerFailure,
					},
				}, err
			}

			return &dnsmessage.Message{
				Header: dnsmessage.Header{
					Response: true,
				},
				Answers: answers,
			}, nil
		}

		if authorities, err = p.AllAuthorities(); err != nil {
			fmt.Printf("p.AllAuthorities() error %v", err)
			return nil, err
		}

		if len(authorities) == 0 {
			return &dnsmessage.Message{
				Header: dnsmessage.Header{
					RCode: dnsmessage.RCodeNameError,
				},
			}, nil
		}

		var nameservers []string
		for _, name := range authorities {
			nameservers = append(nameservers, name.Header.Name.String())
		}

		if additionals, err = p.AllAdditionals(); err != nil && err != dnsmessage.ErrSectionDone {
			fmt.Printf("p.AllAdditionals() error %v", err)
			continue
		}

    var has_ip bool = false
    var new_servers []net.IP

    for _, nameserver := range nameservers {
      for _, packet := range additionals {
        if(nameserver == packet.Header.Name.String()) {
          has_ip = true
          new_servers = append(new_servers, packet.Body.)
        }
      }
    }

		}

		// NOTE: Case when the additionals will have no ips
	}
	return dnsmessage.Message{}, nil
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
		conn, err = net.Dial("udp", server.String()+":53")
		if err != nil {
			log.Printf("conn fail for %s", server.String())
			continue
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
	return nil, nil, fmt.Errorf("No connection found in %+v", servers)
}
