// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0
package babel

import (
	"net/netip"

	"cunicu.li/go-babel/internal/table"
	"cunicu.li/go-babel/proto"
)

type sourceKey struct {
	Prefix   netip.Prefix
	RouterID proto.RouterID
}

type SourceTable table.Table[sourceKey, *Source]

func (t *SourceTable) Lookup(pfx netip.Prefix, rid proto.RouterID) (*Source, bool) {
	return (*table.Table[sourceKey, *Source])(t).Lookup(sourceKey{
		Prefix:   pfx,
		RouterID: rid,
	})
}

func (t *SourceTable) Insert(s *Source) {
	(*table.Table[sourceKey, *Source])(t).Insert(sourceKey{
		s.Prefix,
		s.RouterID,
	}, s)
}
