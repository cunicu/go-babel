// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package babel

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/stv0g/go-babel/proto"
	"golang.org/x/net/ipv6"
)

const (
	// TrafficClassNetworkControl represents a class selector code-point
	// as defined by RFC2474.
	// Routing protocols are recommended to use the __network control_ service class (CS6)
	// as recommended by RFC4594.
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

	Neighbours NeighbourTable

	helloMulticastSeqNo proto.SequenceNumber
	helloMulticastTimer *time.Ticker
	periodicUpdateTimer *time.Ticker

	speaker       *Speaker
	conn          *ipv6.PacketConn
	nextSendTimer Deadline
	pendingValues []proto.Value
	logger        Logger
}

func (s *Speaker) newInterface(index int) (*Interface, error) {
	intf, err := net.InterfaceByIndex(index)
	if err != nil {
		return nil, err
	}

	i := &Interface{
		Interface: intf,

		Neighbours: NewNeighbourTable(),

		nextSendTimer: NewDeadline(),
		speaker:       s,
		logger:        s.config.LoggerFactory.New(fmt.Sprintf("intf(%s)", intf.Name)),
	}

	if i.conn, err = i.createConn(); err != nil {
		return nil, fmt.Errorf("failed to create socket: %w", err)
	}

	if s.config.Multicast {
		multicastAddr := &net.UDPAddr{
			IP: MulticastGroupIPv6.AsSlice(),
		}

		if err := i.conn.SetMulticastInterface(i.Interface); err != nil {
			return nil, fmt.Errorf("failed to set multicast interface: %w", err)
		}

		if err := i.conn.JoinGroup(i.Interface, multicastAddr); err != nil {
			return nil, fmt.Errorf("failed to join multicast group: %w", err)
		}

		if err := i.conn.SetMulticastHopLimit(1); err != nil {
			return nil, fmt.Errorf("failed to set multicast hop limit: %w", err)
		}

		if err := i.conn.SetMulticastLoopback(false); err != nil {
			return nil, fmt.Errorf("failed to set multicast loopback: %w", err)
		}
	}

	go i.runTimers()
	go i.runReadLoop()

	return i, nil
}

func (i *Interface) Close() error {
	i.periodicUpdateTimer.Stop()

	if err := i.conn.Close(); err != nil {
		return fmt.Errorf("failed to close interface: %w", err)
	}

	return nil
}

func (i *Interface) runTimers() {
	i.helloMulticastTimer = time.NewTicker(i.speaker.config.MulticastHelloInterval)
	i.periodicUpdateTimer = time.NewTicker(i.speaker.config.UpdateInterval)

	for {
		select {
		case <-i.periodicUpdateTimer.C:
			if err := i.sendUpdate(); err != nil {
				log.Printf("Failed to send periodic update: %s", err)
			}

		case <-i.helloMulticastTimer.C:
			if err := i.sendMulticastHello(); err != nil {
				log.Printf("Failed to send multicast hello: %s", err)
			}

		case <-i.nextSendTimer.C:
			if err := i.sendPacket(&proto.Packet{
				Body: i.pendingValues,
			}, MulticastGroupIPv6); err != nil {
				log.Printf("Failed to send packet: %s", err)
			} else {
				i.pendingValues = nil
			}
		}
	}
}

func (i *Interface) runReadLoop() {
	buf := make([]byte, i.MTU)

	for {
		if err := i.read(buf); err != nil {
			i.logger.Errorf("Failed to read: %s", err)
		}
	}
}

func (i *Interface) read(buf []byte) error {
	n, _, sAddr, err := i.conn.ReadFrom(buf)
	if err != nil {
		return fmt.Errorf("failed to read: %w", err)
	}

	// TODO: Ignore silently if source address
	// - not link-local
	// - not IPv4 of non-local network
	// - source port is not well-known

	// Ignore packet silently in case of:
	// - magic mismatch
	// - version mismatch
	if !proto.IsBabelPacket(buf[:n]) {
		return fmt.Errorf("received invalid packet")
	}

	p := proto.Parser{}

	_, pkt, err := p.Packet(buf[:n])
	if err != nil {
		return fmt.Errorf("failed to decode packet: %w", err)
	}

	srcAddr := proto.AddressFrom(sAddr)

	return i.onPacket(pkt, srcAddr)
}

func (i *Interface) onPacket(pkt *proto.Packet, srcAddr proto.Address) error {
	n, ok := i.Neighbours.Lookup(srcAddr)
	if !ok {
		var err error
		if n, err = i.NewNeighbour(srcAddr); err != nil {
			return fmt.Errorf("failed to create neighbour: %w", err)
		}

		i.Neighbours.Insert(n)
	}

	i.logger.Debugf("Received multicast packet on interface %s:\n%s", i.Name, spew.Sdump(pkt))

	return n.onPacket(pkt)
}

func (i *Interface) sendPacket(pkt *proto.Packet, dstAddr proto.Address) error {
	// TODO: Implement pacing
	// TODO: Implement chunking

	p := proto.Parser{}

	fmt.Printf("Packet send: %s\n", dstAddr)
	spew.Dump(pkt)

	pktLen := p.PacketLength(pkt)
	buf := make([]byte, 0, pktLen)
	buf = p.AppendPacket(buf, pkt)

	addr := &net.UDPAddr{
		IP:   dstAddr.AsSlice(),
		Port: Port,
	}

	cm := &ipv6.ControlMessage{
		// IfIndex: ifIndex,
	}

	_, err := i.conn.WriteTo(buf, cm, addr)
	return err
}

func (i *Interface) sendMulticastHello() error {
	i.helloMulticastSeqNo++

	if i.speaker.config.Multicast {
		return i.sendValues(&proto.Hello{
			Seqno:    i.helloMulticastSeqNo,
			Interval: i.speaker.config.MulticastHelloInterval,
		})
	} else {
		if err := i.Neighbours.Foreach(func(a proto.Address, n *Neighbour) error {
			return i.sendValues(&proto.Hello{
				Seqno:    i.helloMulticastSeqNo,
				Interval: i.speaker.config.MulticastHelloInterval,
			})
		}); err != nil {
			return err
		}
	}

	return nil
}

func (i *Interface) sendUpdate() error {
	return i.sendValues(&proto.Update{})
}

func (i *Interface) sendValues(vs ...proto.Value) error {
	return i.sendValuesWithJitter(i.speaker.config.MulticastHelloInterval/2, vs...)
}

func (i *Interface) sendValuesWithJitter(maxDelay time.Duration, vs ...proto.Value) error {
	i.pendingValues = append(i.pendingValues, vs...)

	jitter := time.Nanosecond * time.Duration(rand.Float64()*float64(maxDelay))
	i.nextSendTimer.Reset(jitter)

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

// createConn creates a single UDP socket for the speaker
// See: Section 4. Protocol Encoding
// https://datatracker.ietf.org/doc/html/rfc8966#section-4
func (i *Interface) createConn() (*ipv6.PacketConn, error) {
	udpConn, err := net.ListenUDP("udp6", &net.UDPAddr{
		IP:   MulticastGroupIPv6.AsSlice(),
		Port: Port,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create socket: %w", err)
	}

	pktConn := ipv6.NewPacketConn(udpConn)

	if err := pktConn.SetControlMessage(ipv6.FlagDst|ipv6.FlagInterface, true); err != nil {
		return nil, fmt.Errorf("failed to set destination flag: %w", err)
	}

	if err := pktConn.SetHopLimit(1); err != nil {
		return nil, fmt.Errorf("failed to set hop limit: %w", err)
	}

	if err := pktConn.SetTrafficClass(TrafficClassNetworkControl); err != nil {
		return nil, fmt.Errorf("failed to set traffic class: %w", err)
	}

	return pktConn, nil
}

type InterfaceTable Map[int, *Interface]

func NewInterfaceTable() InterfaceTable {
	return InterfaceTable(NewMap[int, *Interface]())
}

func (t *InterfaceTable) Lookup(idx int) (*Interface, bool) {
	return (*Map[int, *Interface])(t).Lookup(idx)
}

func (t *InterfaceTable) Insert(i *Interface) {
	(*Map[int, *Interface])(t).Insert(i.Index, i)
}

func (t *InterfaceTable) Foreach(cb func(int, *Interface) error) error {
	return (*Map[int, *Interface])(t).Foreach(cb)
}
