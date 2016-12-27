package gopack

import "net"
import "time"
import "io"
import "encoding/binary"

// GoPack GoPack main class
type GoPack struct {
	opts  *Options
	conn  *net.TCPConn
	errCh chan error
}

// StorageInterface storage class implementation
type StorageInterface interface {
	UniqueId() uint16
	Save(*Packet)
	Unconfirmed(int) []*Packet
	Confirm(uint16) *Packet
	Receive(uint16, []byte)
	Release(uint16) []byte
}

// Options GoPack create options
type Options struct {
	address         string
	callback        func([]byte, error)
	maxPacketNumber int
	storage         *StorageInterface
	heartbeat       int
}

// NewGoPack creates and initializes a new GoPack using opts
func NewGoPack(opts *Options) *GoPack {
	var gp *GoPack
	gp.opts = opts
	gp.errCh = make(chan error)
	return gp
}

func (gp *GoPack) cb(packet *Packet, err error) {
	if gp.opts.callback != nil {
		gp.opts.callback(packet.Payload, err)
	}
}

func (gp *GoPack) cbErr(err error) {
	if gp.opts.callback != nil {
		gp.opts.callback(nil, err)
	}
}

func (gp *GoPack) readPacket() (packet *Packet, err error) {
	buffer := make([]byte, 5)
	_, err = io.ReadFull(gp.conn, buffer)
	if err != nil {
		return packet, err
	}
	num := buffer[3:]
	remainingLength := binary.BigEndian.Uint16(num)
	payload := make([]byte, remainingLength)
	_, err = io.ReadFull(gp.conn, payload)
	if err != nil {
		return packet, err
	}
	buffer = append(buffer, payload...)
	return Decode(buffer)
}

func (gp *GoPack) read() {
	for {
		packet, err := gp.readPacket()
		if err != nil {
			gp.errCh <- err
			return
		}
		gp.handle(packet)
	}
}

func (gp *GoPack) handle(packet *Packet) {
	if packet.MsgType == MsgTypeSend {
		if packet.Qos == Qos0 {
			gp.cb(packet, nil)
		} else if packet.Qos == Qos1 {
			// TODO
		}
	}
}

// Start internal connection loop
func (gp *GoPack) Start() {
	go gp.SyncStart()
}

// SyncStart internal connection loop (synchronization)
func (gp *GoPack) SyncStart() {
	for {
		conn, err := net.DialTimeout("tcp", gp.opts.address, 2*time.Second)
		if err != nil {
			gp.cbErr(err)
		} else {
			gp.conn = conn.(*net.TCPConn)
			go gp.read()
			err = <-gp.errCh
			if err != nil {
				gp.cbErr(err)
			}
		}
		conn.Close()
		gp.conn = nil
		time.Sleep(3 * time.Second)
	}
}
