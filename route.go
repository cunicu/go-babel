// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package babel

import (
	"github.com/stv0g/go-babel/proto"
)

// 3.2.6. The Route Table
// https://datatracker.ietf.org/doc/html/rfc8966#section-3.2.6
type Route struct {
	Source    *Source
	Neighbour *Neighbour

	Metric         uint16
	SmoothedMetric uint16
	SeqNo          proto.SequenceNumber
	NextHop        proto.Address
	Selected       bool
}

func (r *Route) SetMetric(metric uint16) {
	// r.SmoothedMetric = (ALPHA * r.SmoothedMetric) + ((1 - ALPHA) * metric)
}

type routeKey struct {
	Prefix   proto.Prefix
	RouterID proto.RouterID
}

type RouteTable Map[routeKey, *Route]

func (t *RouteTable) Lookup(pfx proto.Prefix, rid proto.RouterID) (*Route, bool) {
	return (*Map[routeKey, *Route])(t).Lookup(routeKey{
		Prefix:   pfx,
		RouterID: rid,
	})
}

func (t *RouteTable) Insert(r *Route) {
	(*Map[routeKey, *Route])(t).Insert(routeKey{
		r.Source.Prefix,
		r.Source.RouterID,
	}, r)
}
