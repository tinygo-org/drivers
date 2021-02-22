// Package mqtt is intended to provide compatible interfaces with the
// Paho mqtt library.
package mqtt

import (
	"errors"
	"strings"
	"time"

	"github.com/Nerzal/drivers/net"
	"github.com/Nerzal/drivers/net/tls"
	"github.com/Nerzal/drivers/net/ws"
	"github.com/eclipse/paho.mqtt.golang/packets"
	"tinygo.org/x/drivers/net/ws"
)

// NewClient will create an MQTT v3.1.1 client with all of the options specified
// in the provided ClientOptions. The client must have the Connect method called
// on it before it may be used. This is to make sure resources (such as a net
// connection) are created before the application is actually ready.
func NewClient(o *ClientOptions) Client {
	c := &mqttclient{opts: o, adaptor: o.Adaptor}
	c.msgRouter, c.stopRouter = newRouter()
	return c
}

type mqttclient struct {
	adaptor         net.DeviceDriver
	conn            net.Conn
	connected       bool
	opts            *ClientOptions
	mid             uint16
	inbound         chan packets.ControlPacket
	stop            chan struct{}
	msgRouter       *router
	stopRouter      chan bool
	incomingPubChan chan *packets.PublishPacket
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

	if c == nil {
		println("client was nil")
	}

	println("make connection")
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
			println("failed to dial:", err.Error())
			return &mqtttoken{err: err}
		}
	} else if strings.Contains(c.opts.Servers, "ws://") {
		websocket := ws.New(c.opts.Servers)
		websocket.Open()
		c.conn = websocket
	} else {
		// invalid protocol
		return &mqtttoken{err: errors.New("invalid protocol")}
	}

	println("finished dialing")

	c.mid = 1
	c.inbound = make(chan packets.ControlPacket, 10)
	c.stop = make(chan struct{})
	c.incomingPubChan = make(chan *packets.PublishPacket, 10)
	c.msgRouter.matchAndDispatch(c.incomingPubChan, c.opts.Order, c)

	println("matched and dispatched")

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

	connectPkt.ClientIdentifier = c.opts.ClientID
	connectPkt.ProtocolVersion = byte(c.opts.ProtocolVersion)
	connectPkt.ProtocolName = "MQTT"
	connectPkt.Keepalive = 60

	println("sending connect message")

	err = connectPkt.Write(c.conn)
	if err != nil {
		return &mqtttoken{err: err}
	}

	println("sent connect message")

	// TODO: handle timeout as ReadPacket blocks until it gets a packet.
	// CONNECT response.
	packet, err := packets.ReadPacket(c.conn)
	if err != nil {
		return &mqtttoken{err: err}
	}
	if packet != nil {
		ack, ok := packet.(*packets.ConnackPacket)
		if ok {
			if ack.ReturnCode != 0 {
				return &mqtttoken{err: errors.New(packet.String())}
			}
			c.connected = true
		}
	}

	go readMessages(c)
	go processInbound(c)

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
	if !c.IsConnected() {
		return &mqtttoken{err: errors.New("MQTT client not connected")}
	}

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
	if err != nil {
		return &mqtttoken{err: err}
	}

	return &mqtttoken{}
}

// Subscribe starts a new subscription. Provide a MessageHandler to be executed when
// a message is published on the topic provided.
func (c *mqttclient) Subscribe(topic string, qos byte, callback MessageHandler) Token {
	if !c.IsConnected() {
		return &mqtttoken{err: errors.New("MQTT client not connected")}
	}

	sub := packets.NewControlPacket(packets.Subscribe).(*packets.SubscribePacket)
	sub.Topics = append(sub.Topics, topic)
	sub.Qoss = append(sub.Qoss, qos)

	if callback != nil {
		c.msgRouter.addRoute(topic, callback)
	}

	sub.MessageID = c.mid
	c.mid++

	// drop in the channel to send
	err := sub.Write(c.conn)
	if err != nil {
		return &mqtttoken{err: err}
	}

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

func processInbound(c *mqttclient) {
	for {
		select {
		case msg := <-c.inbound:
			switch m := msg.(type) {
			case *packets.PingrespPacket:
				// TODO: handle this
			case *packets.SubackPacket:
				// TODO: handle this
			case *packets.UnsubackPacket:
				// TODO: handle this
			case *packets.PublishPacket:
				// TODO: handle Qos
				c.incomingPubChan <- m
			case *packets.PubackPacket:
				// TODO: handle this
			case *packets.PubrecPacket:
				// TODO: handle this
			case *packets.PubrelPacket:
				// TODO: handle this
			case *packets.PubcompPacket:
				// TODO: handle this
			}
		case <-c.stop:
			return
		}
	}
}

// readMessages reads incoming messages off the wire.
// incoming messages are then send into inbound channel.
func readMessages(c *mqttclient) {
	var err error
	var cp packets.ControlPacket

PROCESS:
	for {
		if cp, err = c.ReadPacket(); err != nil {
			break PROCESS
		}
		if cp != nil {
			c.inbound <- cp
			// TODO: Notify keepalive logic that we recently received a packet
		}

		time.Sleep(100 * time.Millisecond)
	}

	// TODO: handle if we received an error on read.
	// If disconnect is in progress, swallow error and return
}

func (c *mqttclient) ackFunc(packet *packets.PublishPacket) func() {
	return func() {
		switch packet.Qos {
		case 2:
			// pr := packets.NewControlPacket(packets.Pubrec).(*packets.PubrecPacket)
			// pr.MessageID = packet.MessageID
			// DEBUG.Println(NET, "putting pubrec msg on obound")
			// select {
			// case c.oboundP <- &PacketAndToken{p: pr, t: nil}:
			// case <-c.stop:
			// }
			// DEBUG.Println(NET, "done putting pubrec msg on obound")
		case 1:
			// pa := packets.NewControlPacket(packets.Puback).(*packets.PubackPacket)
			// pa.MessageID = packet.MessageID
			// DEBUG.Println(NET, "putting puback msg on obound")
			// persistOutbound(c.persist, pa)
			// select {
			// case c.oboundP <- &PacketAndToken{p: pa, t: nil}:
			// case <-c.stop:
			// }
			// DEBUG.Println(NET, "done putting puback msg on obound")
		case 0:
			// do nothing, since there is no need to send an ack packet back
		}
	}
}

// ReadPacket tries to read the next incoming packet from the MQTT broker.
// If there is no data yet but also is no error, it returns nil for both values.
func (c *mqttclient) ReadPacket() (packets.ControlPacket, error) {
	// check for data first...
	if net.ActiveDevice.IsSocketDataAvailable() {
		return packets.ReadPacket(c.conn)
	}
	return nil, nil
}
