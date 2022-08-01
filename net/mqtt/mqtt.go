// Package mqtt is intended to provide compatible interfaces with the
// Paho mqtt library.
package mqtt

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/eclipse/paho.mqtt.golang/packets"
	"tinygo.org/x/drivers/net"
	"tinygo.org/x/drivers/net/tls"
)

// NewClient will create an MQTT v3.1.1 client with all of the options specified
// in the provided ClientOptions. The client must have the Connect method called
// on it before it may be used. This is to make sure resources (such as a net
// connection) are created before the application is actually ready.
func NewClient(o *ClientOptions) Client {
	c := &mqttclient{opts: o, adaptor: o.Adaptor}
	c.msgRouter, c.stopRouter = newRouter()

	c.inboundPacketChan = make(chan packets.ControlPacket, 10)
	c.stopInbound = make(chan struct{})
	c.incomingPubChan = make(chan *packets.PublishPacket, 10)
	// this launches a goroutine, so only call once per client:
	c.msgRouter.matchAndDispatch(c.incomingPubChan, c.opts.Order, c)
	return c
}

type mqttclient struct {
	adaptor           net.Adapter
	conn              net.Conn
	connected         bool
	opts              *ClientOptions
	mid               uint16
	inboundPacketChan chan packets.ControlPacket
	stopInbound       chan struct{}
	msgRouter         *router
	stopRouter        chan bool
	incomingPubChan   chan *packets.PublishPacket
	// stats for keepalive
	lastReceive time.Time
	lastSend    time.Time
	// keep track of routines and signal a shutdown
	workers  sync.WaitGroup
	shutdown bool
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
	if c.IsConnected() {
		return &mqtttoken{}
	}
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

	c.mid = 1

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
	connectPkt.Keepalive = uint16(c.opts.KeepAlive)

	connectPkt.WillFlag = c.opts.WillEnabled
	connectPkt.WillTopic = c.opts.WillTopic
	connectPkt.WillMessage = c.opts.WillPayload
	connectPkt.WillQos = c.opts.WillQos
	connectPkt.WillRetain = c.opts.WillRetained

	err = connectPkt.Write(c.conn)
	if err != nil {
		return &mqtttoken{err: err}
	}
	c.lastSend = time.Now()

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

	go processInbound(c)
	go readMessages(c)
	go keepAlive(c)

	return &mqtttoken{}
}

// Disconnect will end the connection with the server, but not before waiting
// the specified number of milliseconds to wait for existing work to be
// completed. Blocks until disconnected.
func (c *mqttclient) Disconnect(quiesce uint) {
	c.shutdownRoutines()
	// block until all done
	for c.connected {
		time.Sleep(time.Millisecond * 10)
	}
	return
}

// shutdownRoutines will disconnect and shut down all processes. If you want to trigger a
// disconnect internally, make sure you call this instead of Disconnect() to avoid deadlocks
func (c *mqttclient) shutdownRoutines() {
	if c.shutdown {
		return
	}
	c.shutdown = true
	c.conn.Close()
	c.stopInbound <- struct{}{}
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
	pub.Retain = retained

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
	// update this for every control message that is sent successfully, for keepalive
	c.lastSend = time.Now()

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
	c.lastSend = time.Now()

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
PROCESS:
	for {
		select {
		case msg := <-c.inboundPacketChan:
			switch m := msg.(type) {
			case *packets.PingrespPacket:
				// println("pong")
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
		case <-c.stopInbound:
			break PROCESS
		}
	}

	// as this routine could be the last to finish (if a lot of messages are queued in the
	// channel), it is the last to turn out the lights

	c.workers.Wait()
	c.connected = false
	c.shutdown = false
}

// readMessages reads incoming messages off the wire.
// incoming messages are then send into inbound buffered channel.
func readMessages(c *mqttclient) {
	c.workers.Add(1)
	defer c.workers.Done()

	var err error
	var cp packets.ControlPacket

	for !c.shutdown {
		if cp, err = c.ReadPacket(); err != nil {
			c.shutdownRoutines()
			return
		}
		if cp != nil {
			c.inboundPacketChan <- cp
			// notify keepalive logic that we recently received a packet
			c.lastReceive = time.Now()
		}

		time.Sleep(100 * time.Millisecond)
	}
}

// keepAlive is a goroutine to handle sending ping requests according to the MQTT spec. If the keepalive time has
// been reached with no messages being sent, we will send a ping request and check back to see if we've
// had any activity by the timeout. If not, disconnect.
func keepAlive(c *mqttclient) {
	c.workers.Add(1)
	defer c.workers.Done()

	var err error
	var ping *packets.PingreqPacket
	var timeout, pingsent time.Time

	for !c.shutdown {
		// As long as we haven't reached the keepalive value...
		if time.Since(c.lastSend) < time.Duration(c.opts.KeepAlive)*time.Second {
			// ...sleep and check shutdown status again
			time.Sleep(time.Millisecond * 100)
			continue
		}

		// value has been reached, so send a ping request
		ping = packets.NewControlPacket(packets.Pingreq).(*packets.PingreqPacket)
		if err = ping.Write(c.conn); err != nil {
			// if connection is lost, report disconnect
			c.shutdownRoutines()
			return
		}
		// println("ping")

		c.lastSend = time.Now()
		pingsent = time.Now()
		timeout = pingsent.Add(c.opts.PingTimeout)

		// as long as we are still connected and haven't received anything after the ping...
		for !c.shutdown && c.lastReceive.Before(pingsent) {
			// if the timeout has passed, disconnect
			if time.Now().After(timeout) {
				c.shutdownRoutines()
				return
			}
			time.Sleep(time.Millisecond * 100)
		}
	}
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
