// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package babel

import (
	"cunicu.li/go-babel/internal/table"
	"cunicu.li/go-babel/proto"
)

type routeKey struct {
	Prefix   proto.Prefix
	RouterID proto.RouterID
}

type RouteTable table.Table[routeKey, *Route]

func (t *RouteTable) Lookup(pfx proto.Prefix, rid proto.RouterID) (*Route, bool) {
	return (*table.Table[routeKey, *Route])(t).Lookup(routeKey{
		Prefix:   pfx,
		RouterID: rid,
	})
}

func (t *RouteTable) Insert(r *Route) {
	(*table.Table[routeKey, *Route])(t).Insert(routeKey{
		r.Source.Prefix,
		r.Source.RouterID,
	}, r)
}
