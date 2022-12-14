package udp

import (
	"golang.org/x/xerrors"
	"math"
	"net"
	"os"
	"sync"
	"time"

	"go.dedis.ch/cs438/transport"
)

const bufSize = 65000

// NewUDP returns a new udp transport implementation.
func NewUDP() transport.Transport {
	return &UDP{}
}

// UDP implements a transport layer using UDP
//
// - implements transport.Transport
type UDP struct{}

// CreateSocket implements transport.Transport
func (n *UDP) CreateSocket(address string) (transport.ClosableSocket, error) {
	// Resolve the address to the form of IP:port
	udpAddr, err := net.ResolveUDPAddr("udp4", address)
	if err != nil {
		return nil, xerrors.Errorf("UDP CreateSocket ResolveUDPAddr error: %v", err)
	}

	// Listen to the specified UDP address and create the connection
	conn, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		return nil, xerrors.Errorf("UDP CreateSocket ListenUDP error: %v", err)
	}

	return &Socket{
		UDP:  n,
		conn: conn,
		ins:  packets{},
		outs: packets{},
	}, nil
}

// Socket implements a network socket using UDP.
//
// - implements transport.Socket
// - implements transport.ClosableSocket
type Socket struct {
	*UDP              // shared UDP object
	conn *net.UDPConn // connection per Socket
	ins  packets      // Receiving packets
	outs packets      // Sending packets
}

// Close implements transport.Socket. It returns an error if already closed.
func (s *Socket) Close() error {
	return s.conn.Close()
}

// Send implements transport.Socket
func (s *Socket) Send(dest string, pkt transport.Packet, timeout time.Duration) error {
	if timeout == 0 {
		timeout = math.MaxInt64
	}

	// Create the binary data that could be sent through the network
	data, err := pkt.Marshal()
	if err != nil {
		return err
	}

	// Resolve the UDP address
	udpAddr, err := net.ResolveUDPAddr("udp4", dest)
	if err != nil {
		return err
	}

	// Send the packet out with the specified timeout
	err = s.conn.SetWriteDeadline(time.Now().Add(timeout))
	if err != nil {
		return err
	}

	_, err = s.conn.WriteTo(data, udpAddr)
	if err != nil {
		if !os.IsTimeout(err) {
			// It is not a timeout error
			return err
		}
		return transport.TimeoutError(timeout)
	}

	// The packet is successfully write within the deadline
	s.outs.add(pkt)
	return nil
}

// Recv implements transport.Socket. It blocks until a packet is received, or
// the timeout is reached. In the case the timeout is reached, return a
// TimeoutErr.
func (s *Socket) Recv(timeout time.Duration) (transport.Packet, error) {
	if timeout == 0 {
		timeout = math.MaxInt64
	}

	// buffer to store the content receiving from the socket
	buffer := make([]byte, bufSize)

	// Receive the packet with the specified timeout
	err := s.conn.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		return transport.Packet{}, err
	}

	n, _, err := s.conn.ReadFromUDP(buffer)
	if err != nil {
		if !os.IsTimeout(err) {
			// It is not a timeout error
			return transport.Packet{}, err
		}
		return transport.Packet{}, transport.TimeoutError(timeout)
	}

	// Decode the binary data to Packet
	var pkt transport.Packet
	err = pkt.Unmarshal(buffer[:n])
	if err != nil {
		return transport.Packet{}, err
	}

	// The packet is successfully write within the deadline
	s.ins.add(pkt)
	return pkt, nil

}

// GetAddress implements transport.Socket. It returns the address assigned. Can
// be useful in the case one provided a :0 address, which makes the system use a
// random free port.
func (s *Socket) GetAddress() string {
	return s.conn.LocalAddr().String()
}

// GetIns implements transport.Socket
func (s *Socket) GetIns() []transport.Packet {
	return s.ins.getAll()
}

// GetOuts implements transport.Socket
func (s *Socket) GetOuts() []transport.Packet {
	return s.outs.getAll()
}

type packets struct {
	sync.Mutex
	data []transport.Packet
}

func (p *packets) add(pkt transport.Packet) {
	p.Lock()
	defer p.Unlock()

	p.data = append(p.data, pkt.Copy())
}

func (p *packets) getAll() []transport.Packet {
	p.Lock()
	defer p.Unlock()

	res := make([]transport.Packet, len(p.data))

	for i, pkt := range p.data {
		res[i] = pkt.Copy()
	}

	return res
}
