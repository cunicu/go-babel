// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package babel

import (
	"fmt"
	"log"
	"net"

	"github.com/stv0g/go-babel/proto"
)

// 3.2. Data Structures
// https://datatracker.ietf.org/doc/html/rfc8966#section-3.2

type SpeakerConfig struct {
	*Parameters

	RouterID        proto.RouterID
	InterfaceFilter func(string) bool
	RouteFilter     func(*Route) proto.Metric
	UnicastPeers    []net.UDPAddr
	Multicast       bool
	LoggerFactory   LoggerFactory
}

func (c *SpeakerConfig) SetDefaults() {
	if c.Parameters == nil {
		dp := DefaultParameters
		c.Parameters = &dp
	}

	if c.LoggerFactory == nil {
		c.LoggerFactory = &DefaultLoggerFactory{}
	}
}

type Speaker struct {
	// TODO: Use field
	//nolint:unused
	seqNo uint16

	Interfaces InterfaceTable
	Sources    SourceTable
	Routes     RouteTable

	config SpeakerConfig
	logger Logger
}

func NewSpeaker(cfg *SpeakerConfig) (*Speaker, error) {
	var err error

	s := &Speaker{
		config: *cfg,

		Interfaces: NewInterfaceTable(),
		Sources:    SourceTable{},
		Routes:     RouteTable{},
	}

	s.config.SetDefaults()

	s.logger = s.config.LoggerFactory.New("speaker")

	// Generate router ID
	if s.config.RouterID == 0 {
		if s.config.RouterID, err = proto.GenerateRouterID(); err != nil {
			return nil, fmt.Errorf("failed to generate router ID: %w", err)
		}

		log.Printf("Generated random router ID: %#x", s.config.RouterID)
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

		s.Interfaces.Insert(i)
	}

	return s, nil
}

func (s *Speaker) Close() error {
	return nil
}
