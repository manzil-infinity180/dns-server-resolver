package rfc

import (
	"encoding/binary"
	"fmt"
	"net"
)

type OpCode uint16

// RCode represents DNS response codes
type RCode uint8

// DNS Response Codes
const (
	RCodeNoError        RCode = 0 // No error condition
	RCodeFormatError    RCode = 1 // Format error - server unable to interpret request
	RCodeServerFailure  RCode = 2 // Server failure
	RCodeNameError      RCode = 3 // Name Error - domain name does not exist
	RCodeNotImplemented RCode = 4 // Not Implemented - server does not support requested kind of query
	RCodeRefused        RCode = 5 // Refused - server refuses to perform operation
)

// RecordType represents the type of DNS resource record
type RecordType uint16

// DNS Record Types
const (
	TypeA     RecordType = 1  // Host address
	TypeNS    RecordType = 2  // Authoritative name server
	TypeCNAME RecordType = 5  // Canonical name for an alias
	TypeSOA   RecordType = 6  // Start of authority zone
	TypePTR   RecordType = 12 // Domain name pointer
	TypeMX    RecordType = 15 // Mail exchange
	TypeTXT   RecordType = 16 // Text strings
	TypeAAAA  RecordType = 28 // IPv6 host address
)

// RecordClass represents the class of DNS resource record
type RecordClass uint16

// DNS Record Classes
const (
	ClassIN RecordClass = 1 // Internet
	ClassCS RecordClass = 2 // CSNET (Obsolete)
	ClassCH RecordClass = 3 // CHAOS
	ClassHS RecordClass = 4 // Hesiod
)

// QType represents the type in a DNS question
type QType uint16

// DNS Question Types (same values as RecordType)
const (
	QTypeA     QType = 1
	QTypeNS    QType = 2
	QTypeCNAME QType = 5
	QTypeSOA   QType = 6
	QTypePTR   QType = 12
	QTypeMX    QType = 15
	QTypeTXT   QType = 16
	QTypeAAAA  QType = 28
)

// QClass represents the class in a DNS question
type QClass uint16

// DNS Question Classes (same values as RecordClass)
const (
	QClassIN QClass = 1
	QClassCS QClass = 2
	QClassCH QClass = 3
	QClassHS QClass = 4
)

// QuestionEntry represents a DNS question
type QuestionEntry struct {
	QName  string // domain name being queried
	QType  QType  // type of query
	QClass QClass // class of query
}

// DNS Operation Codes
const (
	OpCodeQuery  OpCode = 0 // Standard query
	OpCodeIQuery OpCode = 1 // Inverse query
	OpCodeStatus OpCode = 2 // Server status request
	OpCodeNotify OpCode = 4 // Notify
	OpCodeUpdate OpCode = 5 // Update
)

type RDataCNAME struct {
	CName string // <domain-name> which specifies canonical or primary name for owner, owner name is alias
}

func (d RDataCNAME) String() string {
	return d.CName
}

type RDataMX struct {
	Preference uint16 // lower values are preferred
	Exchange   string // <domain-name> of host willing to act as mail exchange
}

func (d RDataMX) String() string {
	return fmt.Sprintf("%d %s", d.Preference, d.Exchange)
}

type RDataNS struct {
	NsdName string // <domain-name> specifying host which should be authoritative for specified class
}

func (d RDataNS) String() string {
	return d.NsdName
}

type RDataTXT struct {
	TxtData string // one or more character strings
}

func (d RDataTXT) String() string {
	return d.TxtData
}

type RDataA struct {
	Address string // host internet address
}

func (d RDataA) String() string {
	return d.Address
}

type RDataAAAA struct { // RFC 3596
	Address string // 128-bit IPv6 address
}

func (d RDataAAAA) String() string {
	return d.Address
}

type Message struct {
	Header     MessageHeader    // always present
	Question   []QuestionEntry  // question for name server
	Answer     []ResourceRecord // RRs answering question
	Authority  []ResourceRecord // RRs pointing towards authority
	Additional []ResourceRecord // RRs holding additional info
}

type MessageHeader struct {
	ID      uint16   // identifier assigned by program that generated query, copied to reply
	QR      bool     // one-bit field that specifies if message is query (0, false) or response (1, true)
	OPCode  OpCode   // kind of query
	AA      bool     // only in response, specifies that responding server is authority for domain name (corresponds to name matching query or first owner name in answer)
	TC      bool     // TrunCation, specifies message was truncated due to length greater than permitted on the transmission channel
	RD      bool     // Recursion Desired: set in query, copied to response, directs name server to pursue query recursively
	RA      bool     // Recursion Available, set (1) or cleared (0) in response, denotes whether name server supports recursive queries
	Z       struct{} // reserved, must be 0
	RCode   RCode    // response code
	QDCount uint16   // number of entries in questions section
	ANCount uint16   // number of RRs in answer section
	NSCount uint16   // number of RRs in authority records section
	ARCount uint16   // number of RRs in additional records section
}

type ResourceRecord struct {
	Name     string       // domain name associated with record
	Type     RecordType   // two octets containing RR TYPE code
	Class    RecordClass  // two octets containing RR class code
	TTL      uint32       // time in seconds until record should be refreshed in cache
	RDLength uint16       // length of octets in RData
	RData    fmt.Stringer // variable-length string of octets describing resource, format varies by TYPE and CLASS
}

func parseDomainName(fullMessage []byte, bytes []byte) (string, []byte) {
	domainName := ""
	for {
		length := bytes[0]

		isCompressed := length&0b1100_0000 > 0
		if isCompressed {
			pointer := (bytes[0] << 2) + bytes[1]

			resolvedPointer, _ := parseDomainName(fullMessage, fullMessage[pointer:])
			domainName += resolvedPointer

			bytes = bytes[2:]

			break
		}

		bytes = bytes[1:]

		if length == 0 {
			break
		}

		label := bytes[0:length]

		domainName += string(label)
		domainName += "."

		bytes = bytes[length:]
	}
	fmt.Printf("== domain == %v", domainName)
	fmt.Printf("== bytes == %v", bytes)
	return domainName, bytes
}

// transparently handles empty sections
func parseResourceRecords(fullMessage []byte, messageBytes []byte, numRecords uint16) ([]ResourceRecord, []byte, error) {
	resourceRecords := make([]ResourceRecord, numRecords)

	for i := range resourceRecords {

		var domainName string
		domainName, messageBytes = parseDomainName(fullMessage, messageBytes)

		resourceRecords[i].Name = domainName

		resourceRecords[i].Type = RecordType(binary.BigEndian.Uint16(messageBytes[0:2]))
		messageBytes = messageBytes[2:]

		resourceRecords[i].Class = RecordClass(binary.BigEndian.Uint16(messageBytes[0:2]))
		messageBytes = messageBytes[2:]

		resourceRecords[i].TTL = binary.BigEndian.Uint32(messageBytes[0:4])
		messageBytes = messageBytes[4:]

		resourceRecords[i].RDLength = binary.BigEndian.Uint16(messageBytes[0:2])
		messageBytes = messageBytes[2:]

		switch resourceRecords[i].Type {
		case TypeA:
			// read first 32bit/4 byte
			ipv4 := net.IPAddr{
				IP: messageBytes[0:4],
			}
			messageBytes = messageBytes[4:]

			resourceRecords[i].RData = RDataA{
				Address: ipv4.String(),
			}
		case TypeAAAA:
			// read first 128bit (https://datatracker.ietf.org/doc/html/rfc3596#section-2.2)
			ipv6 := net.IPAddr{
				IP: messageBytes[0:16],
			}
			messageBytes = messageBytes[16:]

			resourceRecords[i].RData = RDataAAAA{
				Address: ipv6.String(),
			}
		case TypeCNAME:
			var canonicalDomainName string
			canonicalDomainName, messageBytes = parseDomainName(fullMessage, messageBytes)

			resourceRecords[i].RData = RDataCNAME{
				CName: canonicalDomainName,
			}
		case TypeMX:
			rDataStr := string(messageBytes[0:resourceRecords[i].RDLength])
			messageBytes = messageBytes[resourceRecords[i].RDLength:]

			preference := binary.BigEndian.Uint16([]byte(rDataStr)[0:2])
			var exchangeDomainName string
			exchangeDomainName, messageBytes = parseDomainName(fullMessage, []byte(rDataStr[2:]))

			resourceRecords[i].RData = RDataMX{
				Preference: preference,
				Exchange:   exchangeDomainName,
			}
		case TypeNS:
			rDataStr := string(messageBytes[0:resourceRecords[i].RDLength])
			messageBytes = messageBytes[resourceRecords[i].RDLength:]

			var nsDomainName string
			nsDomainName, messageBytes = parseDomainName(fullMessage, []byte(rDataStr))

			resourceRecords[i].RData = RDataNS{
				NsdName: nsDomainName,
			}
		case TypeTXT:
			rDataStr := string(messageBytes[0:resourceRecords[i].RDLength])
			messageBytes = messageBytes[resourceRecords[i].RDLength:]

			resourceRecords[i].RData = RDataTXT{
				TxtData: rDataStr,
			}
		}
	}
	fmt.Printf("== resourceRecords == %v", resourceRecords)
	fmt.Printf("== bytes == %v", messageBytes)
	return resourceRecords, messageBytes, nil
}

func parseMessage(messageBytes []byte) (Message, error) {
	fullMessage := messageBytes

	headerBytes := messageBytes[0:headerSizeBytes]
	header, err := parseMessageHeader(headerBytes)
	if err != nil {
		return Message{}, fmt.Errorf("could not parse header: %w", err)
	}

	messageBytes = messageBytes[headerSizeBytes:]
	questionEntries, messageBytes, err := parseQuestions(fullMessage, messageBytes, header.QDCount)
	if err != nil {
		return Message{}, fmt.Errorf("could not parse questions: %w", err)
	}

	answerRRs, messageBytes, err := parseResourceRecords(fullMessage, messageBytes, header.ANCount)
	if err != nil {
		return Message{}, fmt.Errorf("could not parse answer RRs: %w", err)
	}

	authorityRRs, messageBytes, err := parseResourceRecords(fullMessage, messageBytes, header.NSCount)
	if err != nil {
		return Message{}, fmt.Errorf("could not parse authority RRs: %w", err)
	}

	additionalRRs, messageBytes, err := parseResourceRecords(fullMessage, messageBytes, header.ARCount)
	if err != nil {
		return Message{}, fmt.Errorf("could not parse additional RRs: %w", err)
	}
	fmt.Printf("== fullMessage == \n %v", fullMessage)

	return Message{
		Header:     header,
		Question:   questionEntries,
		Answer:     answerRRs,
		Authority:  authorityRRs,
		Additional: additionalRRs,
	}, nil
}

const headerSizeBytes = (8 * 2 * 6) / 8 // two octets * 6

func parseMessageHeader(headerBytes []byte) (MessageHeader, error) {
	if len(headerBytes) != 12 {
		return MessageHeader{}, fmt.Errorf("expected header to be 12 bytes")
	}

	id := binary.BigEndian.Uint16(headerBytes[0:2]) // first two bytes = 16 bits = 2 octets

	secondRow := headerBytes[2:4] // second two bytes = 16 bits = 2 octets

	qr := secondRow[0]&byte(0b1000_0000) > 0

	// after consuming QR, shift entire octet to the left so that opcode takes up leading bits
	secondRow[0] = secondRow[0] << 1

	// retrieve four-bit opcode and shift right by remaining 4 bits in octet to "index" at 0 instead of 8
	opcode := OpCode(secondRow[0] & byte(0b1111_0000) >> 4)

	aa := secondRow[0]&byte(0b0000_1000) > 0
	tc := secondRow[0]&byte(0b0000_0100) > 0
	rd := secondRow[0]&byte(0b0000_0010) > 0

	ra := secondRow[1]&byte(0b1000_0000) > 0

	rcode := RCode(secondRow[1] & byte(0b0000_1111))

	qdcount := binary.BigEndian.Uint16(headerBytes[4:6])
	ancount := binary.BigEndian.Uint16(headerBytes[6:8])
	nscount := binary.BigEndian.Uint16(headerBytes[8:10])
	arcount := binary.BigEndian.Uint16(headerBytes[10:12])

	return MessageHeader{
		ID:      id,
		QR:      qr,
		OPCode:  opcode,
		AA:      aa,
		TC:      tc,
		RD:      rd,
		RA:      ra,
		Z:       struct{}{},
		RCode:   rcode,
		QDCount: qdcount,
		ANCount: ancount,
		NSCount: nscount,
		ARCount: arcount,
	}, nil
}

func parseQuestions(fullMessage []byte, messageBytes []byte, numQuestions uint16) ([]QuestionEntry, []byte, error) {
	questionEntries := make([]QuestionEntry, numQuestions)

	for i := range questionEntries {
		var domainName string
		domainName, messageBytes = parseDomainName(fullMessage, messageBytes)
		questionEntries[i].QName = domainName

		questionEntries[i].QType = QType(binary.BigEndian.Uint16(messageBytes[0:2]))
		messageBytes = messageBytes[2:]

		questionEntries[i].QClass = QClass(binary.BigEndian.Uint16(messageBytes[0:2]))
		messageBytes = messageBytes[2:]
	}

	return questionEntries, messageBytes, nil
}
