// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package babel

import (
	"net/netip"

	"cunicu.li/go-babel/internal/table"
	"cunicu.li/go-babel/proto"
)

type NeighbourTable table.Table[proto.Address, *Neighbour]

func NewNeighbourTable() NeighbourTable {
	return NeighbourTable(table.New[proto.Address, *Neighbour]())
}

func (t *NeighbourTable) Lookup(a proto.Address) (*Neighbour, bool) {
	return (*table.Table[proto.Address, *Neighbour])(t).Lookup(a)
}

func (t *NeighbourTable) Insert(n *Neighbour) {
	(*table.Table[proto.Address, *Neighbour])(t).Insert(n.Address, n)
}

func (t *NeighbourTable) Foreach(cb func(*Neighbour) error) error {
	return (*table.Table[proto.Address, *Neighbour])(t).ForEach(func(k netip.Addr, v *Neighbour) error {
		return cb(v)
	})
}
