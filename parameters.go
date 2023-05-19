// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package babel

import (
	"net/netip"
	"time"
)

// Appendix B. Protocol Parameters
// https://datatracker.ietf.org/doc/html/rfc8966#section-appendix.b

type Parameters struct {
	IHUHoldTimeFactor      float32
	IHUInterval            time.Duration
	InitialRequestTimeout  time.Duration
	MulticastHelloInterval time.Duration
	RouteExpiryTime        time.Duration
	SourceGCTime           time.Duration
	UnicastHelloInterval   time.Duration
	UpdateInterval         time.Duration
	UrgentTimeout          time.Duration
	NominalLinkCost        uint16
}

const (
	DefaultIHUInterval            = 12 * time.Second // 3 * DefaultMulticastHelloInterval
	DefaultInitialRequestTimeout  = 2 * time.Second
	DefaultMulticastHelloInterval = 4 * time.Second
	DefaultRouteExpiryTime        = 56 // 3.5 * DefaultUpdateInterval
	DefaultSourceGCTime           = 3 * time.Minute
	DefaultUnicastHelloInterval   = 0                // infinitive, no Hellos are send
	DefaultUpdateInterval         = 16 * time.Second // 4 * DefaultMulticastHelloInterval
	DefaultUrgentTimeout          = 200 * time.Millisecond

	DefaultIHUHoldTimeFactor = 3.5 // times the advertised IHU interval
	DefaultWiredLinkCost     = 96
)

var DefaultParameters = Parameters{
	MulticastHelloInterval: DefaultMulticastHelloInterval,
	UnicastHelloInterval:   DefaultUnicastHelloInterval,
	UpdateInterval:         DefaultUpdateInterval,
	IHUInterval:            DefaultIHUInterval,
	RouteExpiryTime:        DefaultRouteExpiryTime,
	InitialRequestTimeout:  DefaultInitialRequestTimeout,
	UrgentTimeout:          DefaultUrgentTimeout,
	SourceGCTime:           DefaultSourceGCTime,
	NominalLinkCost:        DefaultWiredLinkCost, // TODO: estimated using ETX on wireless links; 2-out-of-3 with C=96 on wired links.
}

// 5. IANA Considerations
// https://datatracker.ietf.org/doc/html/rfc8966#name-iana-considerations
var (
	Port               = 6697
	MulticastGroupIPv6 = netip.MustParseAddr("ff02::1:6")
	MulticastGroupIPv4 = netip.MustParseAddr("224.0.0.111")
)
