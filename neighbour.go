// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package babel

import (
	"math"
	"net"
	"strings"
	"time"

	"github.com/stv0g/go-babel/internal/deadline"
	"github.com/stv0g/go-babel/internal/history"
	netx "github.com/stv0g/go-babel/internal/net"
	"github.com/stv0g/go-babel/internal/queue"
	"github.com/stv0g/go-babel/proto"
	"golang.org/x/exp/slog"
)

// 3.2.4. The Neighbour Table
// https://datatracker.ietf.org/doc/html/rfc8966#section-3.2.4

type helloState struct {
	history       history.Vector
	expectedSeqNo proto.SequenceNumber
}

func (s *helloState) Update(v *proto.Hello) {
	s.history.Update(v.Seqno, s.expectedSeqNo)

	s.expectedSeqNo = v.Seqno + 1

	// if v.Interval > 0 {
	// 	s.ticker.Reset(v.Interval * 3 / 2)
	// }
}

type Neighbour struct {
	intf *Interface

	logger *slog.Logger

	Address proto.Address

	TxCost      uint16
	RxCost      uint16
	Cost        uint16
	NominalCost uint16

	helloUnicast   helloState
	helloMulticast helloState

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
				n.logger.Error("Failed to send Hello", err)
			}

		case <-n.ihuTicker.C:
			// if err := n.sendIHU(); err != nil {
			// 	n.logger.Error("Failed to send IHU", err)
			// }

		case <-n.ihuTimeout.C:
			n.TxCost = 0xFFFF
		}
	}
}

func (n *Neighbour) onUpdate(upd *proto.Update) {
}

func (n *Neighbour) onHello(hello *proto.Hello) {
	if isUnicast := hello.Flags&proto.FlagHelloUnicast != 0; isUnicast {
		n.helloUnicast.Update(hello)
	} else {
		n.helloMulticast.Update(hello)
	}
}

func (n *Neighbour) onIHU(ihu *proto.IHU) {
	// IHU Hold Time is 3.5x the advertised interval
	n.ihuTimeout.Reset(ihu.Interval * 7 / 2)

	n.TxCost = ihu.RxCost

	n.updateCosts()
}

func (n *Neighbour) onRouteRequest(rr *proto.RouteRequest) {
}

func (n *Neighbour) onSeqnoRequest(sr *proto.SeqnoRequest) {
}

func (n *Neighbour) onAcknowledgmentRequest(ar *proto.AcknowledgmentRequest) {
	if err := n.sendAcknowledgment(ar.Opaque, ar.Interval*3/5); err != nil {
		n.logger.Error("Failed to send acknowledgement", err)
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

// TODO: Use function
//
//nolint:unused
func (n *Neighbour) sendIHU() error {
	n.queue.SendValue(&proto.IHU{
		RxCost:   n.RxCost,
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
func (n *Neighbour) updateCosts() {
	n.RxCost = n.helloUnicast.history.OutOf(12, 16, n.NominalCost)

	if n.RxCost == 0xFFFF {
		n.Cost = 0xFFFF
	} else {
		n.Cost = n.TxCost
	}
}
