// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package babel

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	netx "cunicu.li/go-babel/internal/net"
	"cunicu.li/go-babel/internal/queue"
	"cunicu.li/go-babel/proto"
)

const (
	// TrafficClassNetworkControl represents a class selector code-point
	// as defined by RFC 2474.
	// Routing protocols are recommended to use the __network control_ service class (CS6)
	// as recommended by RFC 4594.
	// We shift it by 2 bits to account for the ECN bits of the traffic class octet.
	//
	// See:
	// - https://datatracker.ietf.org/doc/html/rfc2474#autoid-9
	// - https://datatracker.ietf.org/doc/html/rfc4594#section-3.1
	TrafficClassNetworkControl = 48 << 2 // DiffServ / DSCP name CS6
)

// 3.2.3. The Interface Table
// https://datatracker.ietf.org/doc/html/rfc8966#section-3.2.3

type Interface struct {
	*net.Interface

	multicast bool

	Neighbours NeighbourTable

	helloMulticastSeqNo proto.SequenceNumber
	helloMulticastTimer *time.Ticker
	periodicUpdateTimer *time.Ticker

	queue   *queue.Queue
	speaker *Speaker

	logger *slog.Logger
}

func (s *Speaker) newInterface(index int) (*Interface, error) {
	intf, err := net.InterfaceByIndex(index)
	if err != nil {
		return nil, err
	}

	i := &Interface{
		Interface:  intf,
		Neighbours: NewNeighbourTable(),

		speaker: s,

		multicast:           s.config.Multicast,
		helloMulticastTimer: time.NewTicker(s.config.MulticastHelloInterval),
		periodicUpdateTimer: time.NewTicker(s.config.UpdateInterval),

		logger: s.config.Logger.With(
			slog.String("intf", intf.Name)),
	}

	if i.multicast {
		multicastAddr := &net.UDPAddr{
			IP:   MulticastGroupIPv6.AsSlice(),
			Port: Port,
		}

		i.queue = queue.NewQueue(intf.MTU, &netx.PacketConnWriter{
			PacketConn: i.speaker.conn.PacketConn,
			Dest:       multicastAddr,
		})

		if err := i.speaker.conn.JoinGroup(i.Interface, multicastAddr); err != nil {
			return nil, fmt.Errorf("failed to join multicast group: %w", err)
		}
	}

	go i.runTimers()

	i.logger.Debug("Added new interface")

	return i, nil
}

func (i *Interface) Close() error {
	i.periodicUpdateTimer.Stop()

	if i.multicast {
		if err := i.queue.Close(); err != nil {
			return fmt.Errorf("failed to close queue: %w", err)
		}
	}

	return nil
}

func (i *Interface) runTimers() {
	for {
		select {
		case <-i.periodicUpdateTimer.C:
			if err := i.sendUpdate(); err != nil {
				i.logger.Error("Failed to send periodic update", err)
			}

		case <-i.helloMulticastTimer.C:
			if err := i.sendMulticastHello(); err != nil {
				i.logger.Error("Failed to send multicast hello", err)
			}
		}
	}
}

func (i *Interface) onPacket(pkt *proto.Packet, srcAddr, dstAddr proto.Address) error {
	isMulticast := dstAddr.IsLinkLocalMulticast()

	i.logger.Debug("Received packet",
		slog.Any("src_addr", srcAddr),
		slog.Any("dst_addr", dstAddr),
		slog.Bool("multicast", isMulticast),
		slog.Any("packet", pkt))

	n, ok := i.Neighbours.Lookup(srcAddr)
	if !ok {
		var err error
		if n, err = i.NewNeighbour(srcAddr); err != nil {
			return fmt.Errorf("failed to create neighbour: %w", err)
		}

		i.logger.Debug("Found new neighbour",
			slog.Any("addr", srcAddr))

		if h, ok := i.speaker.config.Handler.(NeighbourHandler); ok {
			h.NeighbourAdded(n)
		}

		i.Neighbours.Insert(n)
	}

	return n.onPacket(pkt, srcAddr, dstAddr)
}

func (i *Interface) sendMulticastHello() error {
	i.logger.Debug("Sending multicast hello")

	i.helloMulticastSeqNo++

	i.sendValue(&proto.Hello{
		Seqno:    i.helloMulticastSeqNo,
		Interval: i.speaker.config.MulticastHelloInterval,
	}, i.speaker.config.MulticastHelloInterval/2)

	return nil
}

func (i *Interface) sendUpdate() error {
	// TODO: Implement sending of updates
	// i.logger.Debug("Sending update")
	// i.sendValue(&proto.Update{}, i.speaker.config.MulticastHelloInterval/2)

	return nil
}

// TODO: Use function
func (i *Interface) sendMulticastRouteRequest() error { //nolint:unused
	i.logger.Debug("Sending multicast route request")

	i.sendValue(&proto.RouteRequest{}, i.speaker.config.MulticastHelloInterval/2)

	return nil
}

// TODO: Use function
func (i *Interface) sendMulticastSeqnoRequest() error { //nolint:unused
	i.logger.Debug("Sending multicast seqno request")

	i.sendValue(&proto.SeqnoRequest{}, i.speaker.config.MulticastHelloInterval/2)

	return nil
}

// TODO: Use function
//
//nolint:unused
func (i *Interface) findLinkLocalAddress() (net.IP, error) {
	addrs, err := i.Addrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		ipNetAddr, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}

		ipAddr := ipNetAddr.IP

		if ipAddr.To4() != nil {
			continue // skip IPv4
		}

		if !ipAddr.IsLinkLocalUnicast() {
			continue // skip non link-local
		}

		return ipAddr, nil
	}

	return nil, errors.New("failed to find IPv6 link-local address")
}

func (i *Interface) sendValue(v proto.Value, maxDelay time.Duration) {
	if i.multicast {
		i.queue.SendValue(v, maxDelay)
	} else {
		i.Neighbours.Foreach(func(n *Neighbour) error { //nolint:errcheck
			n.queue.SendValue(v, maxDelay)
			return nil
		})
	}
}
