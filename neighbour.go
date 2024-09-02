// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package babel

import (
	"log/slog"
	"math"
	"net"
	"strings"
	"time"

	"cunicu.li/go-babel/internal/deadline"
	"cunicu.li/go-babel/internal/history"
	netx "cunicu.li/go-babel/internal/net"
	"cunicu.li/go-babel/internal/queue"
	"cunicu.li/go-babel/proto"
)

// 3.2.4. The Neighbour Table
// https://datatracker.ietf.org/doc/html/rfc8966#section-3.2.4

type Neighbour struct {
	intf *Interface

	logger *slog.Logger

	Address proto.Address

	TxCost uint16

	helloUnicast   history.HelloHistory
	helloMulticast history.HelloHistory

	outgoingUnicastHelloSeqNo proto.SequenceNumber

	ihuTicker   *time.Ticker
	helloTicker *time.Ticker
	ihuTimeout  deadline.Deadline

	queue *queue.Queue
}

func (i *Interface) NewNeighbour(addr proto.Address) (*Neighbour, error) {
	neighbourAddr := &net.UDPAddr{
		IP:   addr.AsSlice(),
		Port: Port,
	}

	n := &Neighbour{
		Address: addr,

		queue: queue.NewQueue(i.MTU, &netx.PacketConnWriter{
			PacketConn: i.speaker.conn.PacketConn,
			Dest:       neighbourAddr,
		}),

		ihuTimeout: deadline.NewDeadline(),
		ihuTicker:  time.NewTicker(i.speaker.config.IHUInterval),

		intf: i,

		logger: i.logger,
	}

	// Only create unicast hello ticker, if its enabled.
	// Otherwise, create a stopped ticker.
	if interval := n.intf.speaker.config.UnicastHelloInterval; interval > 0 {
		n.helloTicker = time.NewTicker(interval)
	} else {
		n.helloTicker = time.NewTicker(math.MaxInt64)
		n.helloTicker.Stop()
	}

	go n.runTimers()

	return n, nil
}

func (n *Neighbour) runTimers() {
	for {
		select {
		case <-n.helloTicker.C:
			if err := n.sendUnicastHello(); err != nil {
				n.logger.Error("Failed to send Hello", slog.Any("error", err))
			}

		case <-n.ihuTicker.C:
			if err := n.sendIHU(); err != nil {
				n.logger.Error("Failed to send IHU", slog.Any("error", err))
			}

		case <-n.ihuTimeout.C:
			n.logger.Warn("IHU deadline missed")
			n.TxCost = 0xFFFF
		}
	}
}

func (n *Neighbour) onUpdate(upd *proto.Update) {
}

func (n *Neighbour) onHello(hello *proto.Hello) {
	if isUnicast := hello.Flags&proto.FlagHelloUnicast != 0; isUnicast {
		n.helloUnicast.Update(hello.Seqno)
	} else {
		n.helloMulticast.Update(hello.Seqno)
	}

	n.logger.Debug("Handled Hello", "rxcost", n.RxCost())
}

func (n *Neighbour) onIHU(ihu *proto.IHU) {
	n.ihuTimeout.Reset(time.Duration(n.intf.speaker.config.IHUHoldTimeFactor * float32(ihu.Interval)))

	n.TxCost = ihu.RxCost

	n.logger.Debug("Handled IHU", "txcost", n.TxCost, "rxcost", n.RxCost(), "cost", n.Cost())
}

func (n *Neighbour) onRouteRequest(rr *proto.RouteRequest) {
}

func (n *Neighbour) onSeqnoRequest(sr *proto.SeqnoRequest) {
}

func (n *Neighbour) onAcknowledgmentRequest(ar *proto.AcknowledgmentRequest) {
	if err := n.sendAcknowledgment(ar.Opaque, ar.Interval*3/5); err != nil {
		n.logger.Error("Failed to send acknowledgement", slog.Any("error", err))
	}
}

func (n *Neighbour) onAcknowledgment(a *proto.Acknowledgment) {
}

func (n *Neighbour) onPacket(pkt *proto.Packet, srcAddr, dstAddr proto.Address) error {
	for _, value := range pkt.Body {
		typ := proto.ValuesType(value).String()
		n.logger.Debug("Received value",
			slog.Any("type", typ),
			slog.Any(strings.ToLower(typ), value))

		switch value := value.(type) {
		case *proto.Update:
			n.onUpdate(value)
		case *proto.Acknowledgment:
			n.onAcknowledgment(value)
		case *proto.AcknowledgmentRequest:
			n.onAcknowledgmentRequest(value)
		case *proto.Hello:
			n.onHello(value)
		case *proto.IHU:
			n.onIHU(value)
		case *proto.RouteRequest:
			n.onRouteRequest(value)
		case *proto.SeqnoRequest:
			n.onSeqnoRequest(value)
		}
	}

	// TODO: Handle trailer

	return nil
}

func (n *Neighbour) sendUnicastHello() error {
	n.outgoingUnicastHelloSeqNo++

	n.queue.SendValue(&proto.Hello{
		Flags:    proto.FlagHelloUnicast,
		Seqno:    n.outgoingUnicastHelloSeqNo,
		Interval: n.intf.speaker.config.UnicastHelloInterval,
	}, n.intf.speaker.config.UnicastHelloInterval*3/5)

	return nil
}

// TODO: Use function
func (n *Neighbour) sendUnicastRouteRequest() error { //nolint:unused
	n.queue.SendValue(&proto.RouteRequest{}, n.intf.speaker.config.MulticastHelloInterval/2)

	return nil
}

// TODO: Use function
func (n *Neighbour) sendUnicastSeqnoRequest() error { //nolint:unused
	n.queue.SendValue(&proto.SeqnoRequest{}, n.intf.speaker.config.MulticastHelloInterval/2)

	return nil
}

func (n *Neighbour) sendIHU() error {
	n.queue.SendValue(&proto.IHU{
		RxCost:   n.RxCost(),
		Address:  n.Address,
		Interval: n.intf.speaker.config.IHUInterval,
	}, n.intf.speaker.config.IHUInterval*3/5)

	return nil
}

func (n *Neighbour) sendAcknowledgment(opaque uint16, interval time.Duration) error {
	n.queue.SendValue(&proto.Acknowledgment{
		Opaque: opaque,
	}, interval*2/3)

	return nil
}

// A.2.1. k-out-of-j
// See: https://datatracker.ietf.org/doc/html/rfc8966#section-a.2.1
func (n *Neighbour) RxCost() uint16 {
	if n.helloUnicast.OutOf(2, 3) || n.helloMulticast.OutOf(2, 3) {
		return n.intf.speaker.config.NominalLinkCost
	} else {
		return 0xFFFF
	}
}

func (n *Neighbour) Cost() uint16 {
	if n.RxCost() == 0xFFFF {
		return 0xFFFF
	} else {
		return n.TxCost
	}
}
