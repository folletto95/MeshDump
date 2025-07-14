package mqtt

type ClientOptions struct {
	broker   string
	username string
	password string
}

func NewClientOptions() *ClientOptions {
	return &ClientOptions{}
}

func (o *ClientOptions) AddBroker(broker string) *ClientOptions {
	o.broker = broker
	return o
}

func (o *ClientOptions) SetUsername(u string) *ClientOptions {
	o.username = u
	return o
}

func (o *ClientOptions) SetPassword(p string) *ClientOptions {
	o.password = p
	return o
}

type Token struct {
	err error
}

func (t *Token) Wait() bool   { return true }
func (t *Token) Error() error { return t.err }

// Client is a minimal MQTT client interface
// used by the meshdump package. It is a stub
// implementation that performs no network
// operations but satisfies the required APIs.
type Client interface {
	Connect() Token
	Subscribe(topic string, qos byte, callback func(Client, Message)) Token
	Publish(topic string, qos byte, retained bool, payload interface{}) Token
	Disconnect(quiesce uint)
}

func NewClient(opts *ClientOptions) Client {
	return &client{opts: opts}
}

type client struct {
	opts *ClientOptions
}

func (c *client) Connect() Token {
	return Token{}
}

func (c *client) Subscribe(topic string, qos byte, callback func(Client, Message)) Token {
	return Token{}
}

func (c *client) Publish(topic string, qos byte, retained bool, payload interface{}) Token {
	return Token{}
}

func (c *client) Disconnect(quiesce uint) {}

// Message represents a received MQTT message.
type Message interface {
	Payload() []byte
}
