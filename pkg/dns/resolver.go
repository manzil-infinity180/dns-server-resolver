package dns

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"math/big"
	"net"
	"strings"

	"golang.org/x/net/dns/dnsmessage"
)

const ROOT_SERVERS = "198.41.0.4,199.9.14.201,192.33.4.12,199.7.91.13,192.203.230.10,192.5.5.241,192.112.36.4,198.97.190.53"

func HandlePacket(pc net.PacketConn, addr net.Addr, buf []byte) {
	if err := handlePacket(pc, addr, buf); err != nil {
		fmt.Printf("read error from %s: %s", addr.String(), err)
	}
}
func handlePacket(pc net.PacketConn, addr net.Addr, buf []byte) error {
	p := dnsmessage.Parser{}
	header, err := p.Start(buf)
	if err != nil {
		return err
	}
	question, err := p.Question()
	if err != nil {
		return err
	}
	response, err := dnsQuery(getRootServers(), question)
	if err != nil {
		return err
	}
	response.Header.ID = header.ID
	responseBuff, err := response.Pack()
	if err != nil {
		return err
	}
	_, err = pc.WriteTo(responseBuff, addr)
	if err != nil {
		return err
	}
	return nil
}
func getRootServers() []net.IP {
	rootservers := []net.IP{}
	for _, rootserver := range strings.Split(ROOT_SERVERS, ",") {
		rootservers = append(rootservers, net.ParseIP(rootserver))
	}
	return rootservers
}

/*
MESSAGES Format

All communications inside of the domain protocol are carried in a single
format called a message.  The top level format of message is divided
into 5 sections (some of which are empty in certain cases) shown below:

	+---------------------+
	|        Header       |
	+---------------------+
	|       Question      | the question for the name server
	+---------------------+
	|        Answer       | RRs answering the question
	+---------------------+
	|      Authority      | RRs pointing toward an authority
	+---------------------+
	|      Additional     | RRs holding additional information
	+---------------------+
*/
func dnsQuery(servers []net.IP, question dnsmessage.Question) (*dnsmessage.Message, error) {
	fmt.Printf("Questions %v \n", question)
	for i := 0; i < 3; i++ {
		dnsAnswer, header, err := outgoingDnsQuery(servers, question)
		if err != nil {
			return nil, err
		}
		parsedAnswers, err := dnsAnswer.AllAnswers()
		if err != nil {
			return nil, err
		}
		// take it as dns query like if we already Authoritative we will simply return from here
		if header.Authoritative {
			return &dnsmessage.Message{
				Header:  dnsmessage.Header{Response: true},
				Answers: parsedAnswers,
			}, nil
		}
		authorities, err := dnsAnswer.AllAuthorities()
		if err != nil {
			return nil, err
		}
		if len(authorities) == 0 {
			return &dnsmessage.Message{
				Header: dnsmessage.Header{
					RCode: dnsmessage.RCodeNameError,
				},
			}, nil
		}
		nameservers := make([]string, len(authorities))

		for k, authority := range authorities {
			if authority.Header.Type == dnsmessage.TypeNS {
				// ==== confusing part ====
				/*
					-> The (*dnsmessage.NSResource) part is a type assertion.
					-> It asserts that the Body field of authority is of type *dnsmessage.NSResource, which is a pointer to a dnsmessage.NSResource struct.
					-> If this type assertion fails (i.e., Body is not of type *dnsmessage.NSResource), the program will panic unless handled safely.
				*/
				nameservers[k] = authority.Body.(*dnsmessage.NSResource).NS.String()

			}
		}
		additionals, err := dnsAnswer.AllAdditionals()
		if err != nil {
			return nil, err
		}
		newResolverServersFound := false
		servers = []net.IP{}

		for _, additional := range additionals {
			if additional.Header.Type == dnsmessage.TypeA {
				for _, nameserver := range nameservers {
					if additional.Header.Name.String() == nameserver {
						newResolverServersFound = true
						servers = append(servers, additional.Body.(*dnsmessage.AResource).A[:])
					}
				}
			}
		}
		if !newResolverServersFound {
			for _, nameserver := range nameservers {
				if !newResolverServersFound {
					response, err := dnsQuery(getRootServers(), dnsmessage.Question{
						Name:  dnsmessage.MustNewName(nameserver),
						Type:  dnsmessage.TypeA,
						Class: dnsmessage.ClassINET,
					})
					if err != nil {
						fmt.Printf("warning: lookup of nameserver %s failed: %err\n", nameserver, err)
					} else {
						newResolverServersFound = true
						for _, answer := range response.Answers {
							if answer.Header.Type == dnsmessage.TypeA {
								servers = append(servers, answer.Body.(*dnsmessage.AResource).A[:])
							}
						}
					}

				}
			}
		}
	}
	return &dnsmessage.Message{
		Header: dnsmessage.Header{
			RCode: dnsmessage.RCodeSuccess,
		},
	}, nil
}
func outgoingDnsQuery(servers []net.IP, question dnsmessage.Question) (*dnsmessage.Parser, *dnsmessage.Header, error) {
	max := ^uint16(0)
	randomNumber, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return nil, nil, err
	}
	message := dnsmessage.Message{
		// Header is a representation of a DNS message header.
		Header: dnsmessage.Header{
			ID:       uint16(randomNumber.Int64()),
			Response: false,
			OpCode:   dnsmessage.OpCode(0),
		},
		/*
			Example:
			question := dnsmessage.Question{
					Name:  dnsmessage.MustNewName("www.google.com."),
					Type:  dnsmessage.TypeNS,
					Class: dnsmessage.ClassINET,
			}
		*/
		// A Question is a DNS query.
		Questions: []dnsmessage.Question{question},
	}
	// Pack packs a full Message.
	buf, err := message.Pack()
	if err != nil {
		return nil, nil, err
	}
	var conn net.Conn
	for _, server := range servers {
		// Dial connects to the address on the named network.
		// Examples:
		//
		//	Dial("tcp", "golang.org:http")
		//	Dial("tcp", "192.0.2.1:http")
		//	Dial("tcp", "198.51.100.1:80")
		//	Dial("udp", "[2001:db8::1]:domain")
		//	Dial("udp", "[fe80::1%lo0]:53") <-
		//	Dial("tcp", ":80")
		//
		conn, err = net.Dial("udp", server.String()+":53")
		if err == nil {
			break
		}
	}
	if conn == nil {
		return nil, nil, fmt.Errorf("failed to make connection to server %s", err)
	}
	// Write "writes" data to the connection.
	// Write can be made to time out and return an error after a fixed
	// time limit; see SetDeadline and SetWriteDeadline.
	_, err = conn.Write(buf)
	if err != nil {
		return nil, nil, err
	}
	// UDP messages    512 octets or less
	answer := make([]byte, 512) // size limit of udp - message packet is 512
	// NewReader returns a new [Reader] whose buffer has the default size.
	n, err := bufio.NewReader(conn).Read(answer)
	if err != nil {
		return nil, nil, err
	}

	conn.Close() // no need of connection any more

	// A Parser allows incrementally parsing a DNS message.
	var p dnsmessage.Parser
	//  Start parses the header and enables the parsing of Questions.
	header, err := p.Start(answer[:n])
	if err != nil {
		return nil, nil, fmt.Errorf("parser start error: %s", err)
	}
	questions, err := p.AllQuestions()
	if err != nil {
		return nil, nil, err
	}
	if len(questions) != len(message.Questions) {
		return nil, nil, fmt.Errorf("answer packet doesn't have the same amount of questions")
	}
	// SkipAllQuestions skips all Questions.
	err = p.SkipAllQuestions()
	if err != nil {
		return nil, nil, err
	}
	return &p, &header, nil
}
