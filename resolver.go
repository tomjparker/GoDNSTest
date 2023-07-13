// Error logging to still be implemented
// Current implementation responds with a hardcoded DNS answer resource record of type A (IPv4). 
// Might need to modify the response to match the queried domain name and the requested record type. 
// You could query a DNS cache or a specific DNS server to obtain accurate responses.

package dns

import(
	"fmt"
	"net"

	"golang.org/x/net/dns/dnsmessage"
)

const ROOT_SERVERS = "*Insert ip*"


const MaxDNSPacketSize = 512

func handlePacket(pc net.PacketConn, addr net.Addr, buf []byte) error { // addr needed for whitelist - not yet implemented
	// Parse the received DNS packet
	var parser dnsmessage.Parser
	err := parser.Start(buf)
	if err != nil {
		return err
	}

	// Retrieve the DNS message header
	header, err := parser.Header()
	if err != nil {
		return err
	}

	// Check if the message contains any questions
	if len(header.Questions) == 0 {
		// No questions present, nothing to respond to
		return nil
	}

	// Retrieve the first question from the DNS message
	question := header.Questions[0]

	// Print the received question for debugging purposes
	fmt.Println("Received DNS question:", question.Name.String())

	// Prepare the DNS response
	response := dnsmessage.Message{
		Header: dnsmessage.Header{
			ID:                 header.ID, // Echo the same ID from the received query
			Response:           true,     // This is a response message
			Authoritative:      false,    // Set this flag based on your specific requirements
			Truncated:          false,    // Will be updated later based on response size
			RecursionDesired:   header.RecursionDesired,
			RecursionAvailable: false,    // Set this flag based on your specific requirements
			RCode:              dnsmessage.RCodeSuccess,
		},
		Questions: header.Questions,
		Answers: []dnsmessage.Resource{
			// Add your DNS response answers here
			// Example:
			{
				Header: dnsmessage.ResourceHeader{
					Name:  question.Name,
					Type:  dnsmessage.TypeA,
					Class: dnsmessage.ClassINET,
				},
				Body: &dnsmessage.AResource{
					A: [4]byte{127, 0, 0, 1},
				},
			},
		},
	}

	// Serialize the DNS response
	responseBuf, err := response.Pack()
	if err != nil {
		return err
	}

	// Check if response exceeds maximum allowed size
	if len(responseBuf) > MaxDNSPacketSize {
		// Set the Truncated flag
		response.Header.Truncated = true

		// Truncate the response to the maximum allowed size
		responseBuf = responseBuf[:MaxDNSPacketSize]
	}

	// Send the DNS response back to the client
	_, err = pc.WriteTo(responseBuf, addr)
	if err != nil {
		return err
	}

	return nil
}

func outgoingDnsQuery(servers []net.IP, question dnsmessage.Question) (*dnsmessage.Parser, *dnsmessage.Header, error) {
	for _, server := range servers {
		// Create a UDP connection to send the DNS query
		conn, err := net.Dial("udp", server.String()+":53")
		if err != nil {
			// Failed to establish a connection to this server, try the next one
			continue
		}
		defer conn.Close()

		// Generate a unique ID for each query
		rand.Seed(time.Now().UnixNano()) // Initialize the random number generator
		id := rand.Uint16()

		// Prepare the DNS message
		msg := dnsmessage.Message{
			Header: dnsmessage.Header{
				ID:               id,
				RecursionDesired: true,
			},
			Questions: []dnsmessage.Question{question},
		}

		// Serialize the DNS message into a byte slice
		buf, err := msg.Pack()
		if err != nil {
			return nil, nil, err
		}

		// Send the DNS query
		_, err = conn.Write(buf)
		if err != nil {
			return nil, nil, err
		}

		// Read the response from the DNS server
		response := make([]byte, 512) // Allocate a buffer to store the response
		_, err = conn.Read(response)
		if err != nil {
			return nil, nil, err
		}

		// Parse the DNS response
		parser := &dnsmessage.Parser{}
		err = parser.Start(response)
		if err != nil {
			return nil, nil, err
		}

		// Parse the DNS message header
		header, err := parser.Header()
		if err != nil {
			return nil, nil, err
		}

		return parser, header, nil
	}

	// No successful connections to DNS servers
	return nil, nil, fmt.Errorf("failed to send DNS query to any server")
}