package rfc

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestMessageParsing(t *testing.T) {
	// Create a test DNS message with a query for google.com
	message := createTestMessage()

	// Parse the message
	parsed, err := parseMessage(message)
	if err != nil {
		t.Fatalf("Failed to parse message: %v", err)
	}

	// Verify header fields
	if parsed.Header.ID != 1234 {
		t.Errorf("Expected ID 1234, got %d", parsed.Header.ID)
	}
	if parsed.Header.QR != false {
		t.Errorf("Expected QR false (query), got true")
	}
	if parsed.Header.RD != true {
		t.Errorf("Expected RD true, got false")
	}

	// Verify question section
	if len(parsed.Question) != 1 {
		t.Fatalf("Expected 1 question, got %d", len(parsed.Question))
	}
	q := parsed.Question[0]
	if q.QName != "google.com." {
		t.Errorf("Expected name google.com., got %s", q.QName)
	}
	if q.QType != QTypeA {
		t.Errorf("Expected type A, got %d", q.QType)
	}
	if q.QClass != QClassIN {
		t.Errorf("Expected class IN, got %d", q.QClass)
	}
}

func TestDomainNameCompression(t *testing.T) {
	// Create a message with compressed domain names
	message := createCompressedMessage()

	parsed, err := parseMessage(message)
	if err != nil {
		t.Fatalf("Failed to parse compressed message: %v", err)
	}

	// Verify the compressed names were resolved correctly
	if len(parsed.Answer) != 2 {
		t.Fatalf("Expected 2 answers, got %d", len(parsed.Answer))
	}

	ans1 := parsed.Answer[0]
	if ans1.Name != "www.google.com." {
		t.Errorf("Expected www.google.com., got %s", ans1.Name)
	}

	ans2 := parsed.Answer[1]
	if ans2.Name != "www.google.com." {
		t.Errorf("Expected www.google.com., got %s", ans2.Name)
	}
}

// Helper function to create a test DNS message
func createTestMessage() []byte {
	var buf bytes.Buffer

	// Header (12 bytes)
	header := []byte{
		0x04, 0xd2, // ID = 1234
		0x01, 0x00, // Flags: RD=1
		0x00, 0x01, // QDCOUNT = 1
		0x00, 0x00, // ANCOUNT = 0
		0x00, 0x00, // NSCOUNT = 0
		0x00, 0x00, // ARCOUNT = 0
	}
	buf.Write(header)

	// Question section for google.com
	writeDomainName(&buf, "google.com")

	// QTYPE = A (1)
	binary.Write(&buf, binary.BigEndian, uint16(QTypeA))
	// QCLASS = IN (1)
	binary.Write(&buf, binary.BigEndian, uint16(QClassIN))

	return buf.Bytes()
}

// Helper function to create a message with domain name compression
func createCompressedMessage() []byte {
	var buf bytes.Buffer

	// Header (12 bytes)
	header := []byte{
		0x04, 0xd2, // ID = 1234
		0x81, 0x80, // Flags: QR=1, RD=1, RA=1
		0x00, 0x01, // QDCOUNT = 1
		0x00, 0x02, // ANCOUNT = 2
		0x00, 0x00, // NSCOUNT = 0
		0x00, 0x00, // ARCOUNT = 0
	}
	buf.Write(header)

	// Question
	startPos := buf.Len()
	writeDomainName(&buf, "www.google.com")
	binary.Write(&buf, binary.BigEndian, uint16(QTypeA))
	binary.Write(&buf, binary.BigEndian, uint16(QClassIN))

	// Answer 1 with compression pointer to question
	writeCompressedName(&buf, uint16(startPos))
	binary.Write(&buf, binary.BigEndian, uint16(TypeA))
	binary.Write(&buf, binary.BigEndian, uint16(ClassIN))
	binary.Write(&buf, binary.BigEndian, uint32(300)) // TTL
	binary.Write(&buf, binary.BigEndian, uint16(4))   // RDLENGTH
	buf.Write([]byte{192, 0, 2, 1})                   // IPv4 address

	// Answer 2 with compression pointer to same name
	writeCompressedName(&buf, uint16(startPos))
	binary.Write(&buf, binary.BigEndian, uint16(TypeA))
	binary.Write(&buf, binary.BigEndian, uint16(ClassIN))
	binary.Write(&buf, binary.BigEndian, uint32(300)) // TTL
	binary.Write(&buf, binary.BigEndian, uint16(4))   // RDLENGTH
	buf.Write([]byte{192, 0, 2, 2})                   // Different IPv4 address

	return buf.Bytes()
}

// Helper to write domain name in DNS wire format
func writeDomainName(buf *bytes.Buffer, name string) {
	parts := bytes.Split([]byte(name), []byte("."))
	for _, part := range parts {
		buf.WriteByte(byte(len(part)))
		buf.Write(part)
	}
	buf.WriteByte(0)
}

// Helper to write compressed domain name
func writeCompressedName(buf *bytes.Buffer, pointer uint16) {
	// Set compression bits (top 2 bits) and pointer value
	compressed := pointer | 0xC000
	binary.Write(buf, binary.BigEndian, compressed)
}
