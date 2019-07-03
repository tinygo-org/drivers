package mqtt

import (
	"errors"
	"strings"
	"time"

	"github.com/eclipse/paho.mqtt.golang/packets"
	"tinygo.org/x/drivers/espat"
	"tinygo.org/x/drivers/espat/net"
	"tinygo.org/x/drivers/espat/tls"
)

// NewClient will create an MQTT v3.1.1 client with all of the options specified
// in the provided ClientOptions. The client must have the Connect method called
// on it before it may be used. This is to make sure resources (such as a net
// connection) are created before the application is actually ready.
func NewClient(o *ClientOptions) Client {
	c := &mqttclient{opts: o, adaptor: o.Adaptor}
	return c
}

type mqttclient struct {
	adaptor   *espat.Device
	conn      net.Conn
	connected bool
	opts      *ClientOptions
	mid       uint16
}

// AddRoute allows you to add a handler for messages on a specific topic
// without making a subscription. For example having a different handler
// for parts of a wildcard subscription
func (c *mqttclient) AddRoute(topic string, callback MessageHandler) {
	return
}

// IsConnected returns a bool signifying whether
// the client is connected or not.
func (c *mqttclient) IsConnected() bool {
	return c.connected
}

// IsConnectionOpen return a bool signifying whether the client has an active
// connection to mqtt broker, i.e not in disconnected or reconnect mode
func (c *mqttclient) IsConnectionOpen() bool {
	return c.connected
}

// Connect will create a connection to the message broker.
func (c *mqttclient) Connect() Token {
	var err error

	// make connection
	if strings.Contains(c.opts.Servers, "ssl://") {
		url := strings.TrimPrefix(c.opts.Servers, "ssl://")
		c.conn, err = tls.Dial("tcp", url, nil)
		if err != nil {
			return &mqtttoken{err: err}
		}
	} else if strings.Contains(c.opts.Servers, "tcp://") {
		url := strings.TrimPrefix(c.opts.Servers, "tcp://")
		c.conn, err = net.Dial("tcp", url)
		if err != nil {
			return &mqtttoken{err: err}
		}
	} else {
		// invalid protocol
		return &mqtttoken{err: errors.New("invalid protocol")}
	}

	// send the MQTT connect message
	connectPkt := packets.NewControlPacket(packets.Connect).(*packets.ConnectPacket)
	connectPkt.Qos = 0
	if c.opts.Username != "" {
		connectPkt.Username = c.opts.Username
		connectPkt.UsernameFlag = true
	}

	if c.opts.Password != "" {
		connectPkt.Password = []byte(c.opts.Password)
		connectPkt.PasswordFlag = true
	}

	connectPkt.ClientIdentifier = c.opts.ClientID //"tinygo-client-" + randomString(10)
	connectPkt.ProtocolVersion = byte(c.opts.ProtocolVersion)
	connectPkt.ProtocolName = "MQTT"
	connectPkt.Keepalive = 30

	err = connectPkt.Write(c.conn)
	if err != nil {
		return &mqtttoken{err: err}
	}

	// TODO: handle timeout
	for {
		packet, _ := packets.ReadPacket(c.conn)

		if packet != nil {
			ack, ok := packet.(*packets.ConnackPacket)
			if ok {
				if ack.ReturnCode == 0 {
					// success
					return &mqtttoken{}
				}
				// otherwise something went wrong
				return &mqtttoken{err: errors.New(packet.String())}
			}
		}

		time.Sleep(100 * time.Millisecond)
	}

	c.connected = true
	return &mqtttoken{}
}

// Disconnect will end the connection with the server, but not before waiting
// the specified number of milliseconds to wait for existing work to be
// completed.
func (c *mqttclient) Disconnect(quiesce uint) {
	c.conn.Close()
	return
}

// Publish will publish a message with the specified QoS and content
// to the specified topic.
// Returns a token to track delivery of the message to the broker
func (c *mqttclient) Publish(topic string, qos byte, retained bool, payload interface{}) Token {
	pub := packets.NewControlPacket(packets.Publish).(*packets.PublishPacket)
	pub.Qos = qos
	pub.TopicName = topic
	switch payload.(type) {
	case string:
		pub.Payload = []byte(payload.(string))
	case []byte:
		pub.Payload = payload.([]byte)
	default:
		return &mqtttoken{err: errors.New("Unknown payload type")}
	}
	pub.MessageID = c.mid
	c.mid++

	err := pub.Write(c.conn)
	return &mqtttoken{err: err}
}

// Subscribe starts a new subscription. Provide a MessageHandler to be executed when
// a message is published on the topic provided.
func (c *mqttclient) Subscribe(topic string, qos byte, callback MessageHandler) Token {
	return &mqtttoken{}
}

// SubscribeMultiple starts a new subscription for multiple topics. Provide a MessageHandler to
// be executed when a message is published on one of the topics provided.
func (c *mqttclient) SubscribeMultiple(filters map[string]byte, callback MessageHandler) Token {
	return &mqtttoken{}
}

// Unsubscribe will end the subscription from each of the topics provided.
// Messages published to those topics from other clients will no longer be
// received.
func (c *mqttclient) Unsubscribe(topics ...string) Token {
	return &mqtttoken{}
}

// OptionsReader returns a ClientOptionsReader which is a copy of the clientoptions
// in use by the client.
func (c *mqttclient) OptionsReader() ClientOptionsReader {
	r := ClientOptionsReader{}
	return r
}

type mqtttoken struct {
	err error
}

func (t *mqtttoken) Wait() bool {
	return true
}

func (t *mqtttoken) WaitTimeout(time.Duration) bool {
	return true
}

func (t *mqtttoken) Error() error {
	return t.err
}
