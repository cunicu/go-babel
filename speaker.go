// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package babel

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/netip"

	"github.com/stv0g/go-babel/proto"
	"golang.org/x/exp/slog"
	"golang.org/x/net/ipv6"
)

// 3.2. Data Structures
// https://datatracker.ietf.org/doc/html/rfc8966#section-3.2

type SpeakerConfig struct {
	*Parameters

	Handler         any
	RouterID        proto.RouterID
	InterfaceFilter func(string) bool
	RouteFilter     func(*Route) proto.Metric
	UnicastPeers    []net.UDPAddr
	Multicast       bool
	Logger          *slog.Logger
}

func (c *SpeakerConfig) SetDefaults() error {
	if c.RouterID == proto.RouterIDUnspecified {
		var err error
		if c.RouterID, err = proto.GenerateRouterID(); err != nil {
			return fmt.Errorf("failed to generate router ID: %w", err)
		}

		// TODO: Use slog
		log.Printf("Generated random router ID: %#x", c.RouterID)
	}

	if c.Parameters == nil {
		dp := DefaultParameters
		c.Parameters = &dp
	}

	if c.Logger == nil {
		c.Logger = slog.Default()
	}

	return nil
}

type Speaker struct {
	// TODO: Use field
	//nolint:unused
	seqNo uint16

	Interfaces InterfaceTable
	Sources    SourceTable
	Routes     RouteTable

	conn *ipv6.PacketConn

	config SpeakerConfig
	logger *slog.Logger
}

func NewSpeaker(cfg *SpeakerConfig) (*Speaker, error) {
	var err error

	s := &Speaker{
		config: *cfg,

		Interfaces: NewInterfaceTable(),
		Sources:    SourceTable{},
		Routes:     RouteTable{},
	}

	if err := s.config.SetDefaults(); err != nil {
		return nil, err
	}

	s.logger = s.config.Logger

	if s.conn, err = s.createConn(); err != nil {
		return nil, fmt.Errorf("failed to create conn: %w", err)
	}

	// Find local interfaces
	intfs, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get interfaces: %w", err)
	}

	for _, intf := range intfs {
		if intf.Flags&net.FlagLoopback != 0 {
			continue
		}

		if cfg.InterfaceFilter != nil && !cfg.InterfaceFilter(intf.Name) {
			continue
		}

		i, err := s.newInterface(intf.Index)
		if err != nil {
			return nil, fmt.Errorf("failed to create interface: %w", err)
		}

		if h, ok := s.config.Handler.(InterfaceHandler); ok {
			h.InterfaceAdded(i)
		}

		s.Interfaces.Insert(i)
	}

	go s.runReadLoop()

	return s, nil
}

func (s *Speaker) Close() error {
	if err := s.conn.Close(); err != nil {
		return fmt.Errorf("failed to close interface: %w", err)
	}

	return nil
}

func (s *Speaker) runReadLoop() {
	s.logger.Debug("Start receiving packets")

	// TODO: Check for largest MTU of attached interfaces
	buf := make([]byte, 1500)

	for {
		n, cm, sAddr, err := s.conn.ReadFrom(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}

			s.logger.Error("Failed to read from socket", err)
			continue
		}

		srcAddr := proto.AddressFrom(sAddr)
		dstAddr, ok := netip.AddrFromSlice(cm.Dst)
		if !ok {
			s.logger.Error("Invalid destination address", nil)
			continue
		}

		// TODO: Ignore silently if source address is IPv4 of non-local network

		// Ignore packet from non-link-local source address
		if !srcAddr.IsLinkLocalUnicast() {
			s.logger.Debug("Ignoring packet from non-link-local source", slog.Any("saddr", srcAddr))
			continue
		}

		// Ignore packet from non well-known Babel port number
		if udpAddr, ok := sAddr.(*net.UDPAddr); !ok {
			s.logger.Debug("Ignoring non UDP source address", slog.Any("saddr", srcAddr))
			continue
		} else if udpAddr.Port != Port {
			s.logger.Debug("Ignoring packet from non-babel source port", slog.Any("saddr", udpAddr))
			continue
		}

		// Ignore packet silently in case of:
		// - magic mismatch
		// - version mismatch
		if !proto.IsBabelPacket(buf[:n]) {
			s.logger.Debug("Ignoring non-babel packet")
			continue
		}

		p := proto.Parser{}

		_, pkt, err := p.Packet(buf[:n])
		if err != nil {
			s.logger.Error("Failed to decode packet: %w", err)
			continue
		}

		if err := s.onPacket(pkt, cm.IfIndex, srcAddr, dstAddr); err != nil {
			s.logger.Error("Failed to handle packet", err)
			continue
		}
	}
}

// createConn creates a single UDP socket for the speaker
// See: Section 4. Protocol Encoding
// https://datatracker.ietf.org/doc/html/rfc8966#section-4
func (s *Speaker) createConn() (*ipv6.PacketConn, error) {
	udpConn, err := net.ListenUDP("udp6", &net.UDPAddr{
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

	if err := pktConn.SetMulticastHopLimit(1); err != nil {
		return nil, fmt.Errorf("failed to set multicast hop limit: %w", err)
	}

	if err := pktConn.SetMulticastLoopback(false); err != nil {
		return nil, fmt.Errorf("failed to set multicast loopback: %w", err)
	}

	if err := pktConn.SetTrafficClass(TrafficClassNetworkControl); err != nil {
		return nil, fmt.Errorf("failed to set traffic class: %w", err)
	}

	return pktConn, nil
}

func (s *Speaker) onPacket(pkt *proto.Packet, ifIndex int, srcAddr, dstAddr proto.Address) error {
	i, ok := s.Interfaces.Lookup(ifIndex)
	if !ok {
		s.logger.Debug("Ignoring packet from unknown interface", slog.Int("ifindex", ifIndex))
		return nil
	}

	return i.onPacket(pkt, srcAddr, dstAddr)
}
