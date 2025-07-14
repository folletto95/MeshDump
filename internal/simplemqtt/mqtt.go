package mqtt

import (
	"bufio"
	"errors"
	"io"
	"net"
	"strings"
	"sync"
)

// A very small MQTT client implementing connect and subscribe for QoS 0 messages.
// It is not a full implementation but allows reading messages from a broker
// without external dependencies.

type ClientOptions struct {
	broker   string
	username string
	password string
	clientID string
}

func NewClientOptions() *ClientOptions { return &ClientOptions{clientID: "meshdump"} }

func (o *ClientOptions) AddBroker(b string) *ClientOptions    { o.broker = b; return o }
func (o *ClientOptions) SetUsername(u string) *ClientOptions  { o.username = u; return o }
func (o *ClientOptions) SetPassword(p string) *ClientOptions  { o.password = p; return o }
func (o *ClientOptions) SetClientID(id string) *ClientOptions { o.clientID = id; return o }

// Token mimics the Paho token type but operations complete synchronously,
// so Wait always returns true.
type Token struct{ err error }

func (t Token) Wait() bool   { return true }
func (t Token) Error() error { return t.err }

// Message represents a received MQTT message.
type Message interface{ Payload() []byte }

// Client is a minimal MQTT client that supports Connect and Subscribe.
type Client interface {
	Connect() Token
	Subscribe(topic string, qos byte, cb func(Client, Message)) Token
	Disconnect(quiesce uint)
}

func NewClient(opts *ClientOptions) Client {
	return &client{opts: opts}
}

type client struct {
	opts *ClientOptions
	conn net.Conn
	mu   sync.Mutex
}

func (c *client) Connect() Token {
	addr := strings.TrimPrefix(c.opts.broker, "tcp://")
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return Token{err}
	}
	c.conn = conn
	if err := c.sendConnect(); err != nil {
		conn.Close()
		return Token{err}
	}
	// skip CONNACK
	if _, err := c.readPacket(); err != nil {
		conn.Close()
		return Token{err}
	}
	return Token{}
}

func (c *client) sendConnect() error {
	flags := byte(0)
	if c.opts.username != "" {
		flags |= 0x80
	}
	if c.opts.password != "" {
		flags |= 0x40
	}
	payload := make([]byte, 0)
	payload = appendUTF(payload, c.opts.clientID)
	if c.opts.username != "" {
		payload = appendUTF(payload, c.opts.username)
	}
	if c.opts.password != "" {
		payload = appendUTF(payload, c.opts.password)
	}
	vh := []byte{
		0, 4, 'M', 'Q', 'T', 'T', // protocol name
		4,     // protocol level 4
		flags, // connect flags
		0, 30, // keepalive 30s
	}
	pkt := buildPacket(1<<4, append(vh, payload...))
	_, err := c.conn.Write(pkt)
	return err
}

func (c *client) Subscribe(topic string, qos byte, cb func(Client, Message)) Token {
	c.mu.Lock()
	defer c.mu.Unlock()
	pktID := []byte{0, 1}
	payload := appendUTF(nil, topic)
	payload = append(payload, qos)
	vh := pktID
	pkt := buildPacket(8<<4|2, append(vh, payload...))
	if _, err := c.conn.Write(pkt); err != nil {
		return Token{err}
	}
	// read SUBACK
	if _, err := c.readPacket(); err != nil {
		return Token{err}
	}
	go c.reader(cb)
	return Token{}
}

func (c *client) reader(cb func(Client, Message)) {
	for {
		data, err := c.readPacket()
		if err != nil {
			return
		}
		if len(data) == 0 {
			continue
		}
		t := data[0] >> 4
		if t == 3 { // PUBLISH
			// parse topic
			if len(data) < 3 {
				continue
			}
			tl := int(data[2])<<8 | int(data[3])
			if 4+tl > len(data) {
				continue
			}
			payload := data[4+tl:]
			cb(c, message{payload: payload})
		}
	}
}

func (c *client) readPacket() ([]byte, error) {
	reader := bufio.NewReader(c.conn)
	hdr1, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}
	rem, err := readRemainingLength(reader)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, rem)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return nil, err
	}
	return append([]byte{hdr1}, buf...), nil
}

func (c *client) Disconnect(quiesce uint) {
	if c.conn != nil {
		c.conn.Close()
	}
}

type message struct{ payload []byte }

func (m message) Payload() []byte { return m.payload }

// helpers
func appendUTF(b []byte, s string) []byte {
	l := len(s)
	b = append(b, byte(l>>8), byte(l))
	return append(b, s...)
}

func buildPacket(hdr byte, body []byte) []byte {
	buf := []byte{hdr}
	buf = append(buf, encodeLength(len(body))...)
	return append(buf, body...)
}

func encodeLength(l int) []byte {
	var enc []byte
	for {
		d := byte(l % 128)
		l /= 128
		if l > 0 {
			d |= 128
		}
		enc = append(enc, d)
		if l == 0 {
			break
		}
	}
	return enc
}

func readRemainingLength(r *bufio.Reader) (int, error) {
	mul := 1
	val := 0
	for i := 0; i < 4; i++ {
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		val += int(b&127) * mul
		if b&128 == 0 {
			return val, nil
		}
		mul *= 128
	}
	return 0, errors.New("malformed length")
}
