// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package babel

import (
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/stv0g/go-babel/proto"
	"golang.org/x/exp/slog"
)

// 3.2.4. The Neighbour Table
// https://datatracker.ietf.org/doc/html/rfc8966#section-3.2.4

type helloState struct {
	history       HistoryVector
	expectedSeqNo proto.SequenceNumber
	ticker        *time.Ticker
}

func (s *helloState) Update(v *proto.Hello) {
	s.history.Update(v.Seqno, s.expectedSeqNo)

	s.expectedSeqNo = v.Seqno + 1

	if v.Interval > 0 {
		s.ticker.Reset(v.Interval * 3 / 2)
	}
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
	ihuTimeout                Deadline
	nextSendTimer             Deadline
	pendingValues             []proto.Value
}

func (i *Interface) NewNeighbour(addr proto.Address) (*Neighbour, error) {
	n := &Neighbour{
		Address: addr,

		ihuTimeout: NewDeadline(),
		intf:       i,

		logger: i.logger,
	}

	if interval := n.intf.speaker.config.UnicastHelloInterval; interval > 0 {
		n.helloUnicast.ticker = time.NewTicker(interval)
	} else {
		n.helloUnicast.ticker = time.NewTicker(math.MaxInt64)
		n.helloUnicast.ticker.Stop()
	}

	if interval := n.intf.speaker.config.MulticastHelloInterval; interval > 0 {
		n.helloMulticast.ticker = time.NewTicker(interval)
	} else {
		n.helloMulticast.ticker = time.NewTicker(math.MaxInt64)
		n.helloMulticast.ticker.Stop()
	}

	n.nextSendTimer = NewDeadline()

	go n.runTimers()

	return n, nil
}

func (n *Neighbour) runTimers() {
	for {
		select {
		case <-n.helloUnicast.ticker.C:
			if err := n.sendUnicastHello(); err != nil {
				n.logger.Error("Failed to send Hello", err)
			}

		case <-n.ihuTimeout.C:
			n.TxCost = 0xFFFF

		case <-n.nextSendTimer.C:
			if err := n.sendPacket(&proto.Packet{
				Body: n.pendingValues,
			}); err != nil {
				log.Printf("Failed to send packet: %s", err)
			} else {
				n.pendingValues = nil
			}
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
	if err := n.sendValuesWithJitter(ar.Interval, &proto.Acknowledgment{
		Opaque: ar.Opaque,
	}); err != nil {
		n.intf.logger.Error("Failed to send acknowledgement: %s", err)
	}
}

func (n *Neighbour) onAcknowledgment(a *proto.Acknowledgment) {
}

func (n *Neighbour) onPacket(pkt *proto.Packet) error {
	for _, value := range pkt.Body {
		n.logger.Debug("Received value",
			slog.String("type", fmt.Sprintf("%T", value)))

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

	return nil
}

func (n *Neighbour) sendValues(vs ...proto.Value) error {
	return n.sendValuesWithJitter(n.intf.speaker.config.MulticastHelloInterval/2, vs...)
}

// TODO: Use function
//
//nolint:unused
func (n *Neighbour) sendUrgentValues(vs ...proto.Value) error {
	return n.sendValuesWithJitter(n.intf.speaker.config.UrgentTimeout, vs...)
}

func (n *Neighbour) sendValuesWithJitter(maxDelay time.Duration, vs ...proto.Value) error {
	n.pendingValues = append(n.pendingValues, vs...)

	jitter := time.Nanosecond * time.Duration(rand.Float64()*float64(maxDelay))
	n.nextSendTimer.Reset(jitter)

	return nil
}

func (n *Neighbour) sendUnicastHello() error {
	n.outgoingUnicastHelloSeqNo++

	hello := &proto.Hello{
		Flags:    proto.FlagHelloUnicast,
		Seqno:    n.outgoingUnicastHelloSeqNo,
		Interval: n.intf.speaker.config.UnicastHelloInterval,
	}

	return n.sendValues(hello)
}

// TODO: Use function
//
//nolint:unused
func (n *Neighbour) sendIHU() error {
	ihu := &proto.IHU{
		RxCost:   n.RxCost,
		Address:  n.Address,
		Interval: n.intf.speaker.config.IHUInterval,
	}

	return n.sendValues(ihu)
}

func (n *Neighbour) sendPacket(pkt *proto.Packet) error {
	return n.intf.sendPacket(pkt, n.Address)
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

type NeighbourTable Map[proto.Address, *Neighbour]

func NewNeighbourTable() NeighbourTable {
	return NeighbourTable(NewMap[proto.Address, *Neighbour]())
}

func (t *NeighbourTable) Lookup(a proto.Address) (*Neighbour, bool) {
	return (*Map[proto.Address, *Neighbour])(t).Lookup(a)
}

func (t *NeighbourTable) Insert(n *Neighbour) {
	(*Map[proto.Address, *Neighbour])(t).Insert(n.Address, n)
}

func (t *NeighbourTable) Foreach(cb func(proto.Address, *Neighbour) error) error {
	return (*Map[proto.Address, *Neighbour])(t).Foreach(cb)
}
