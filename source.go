// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package babel

import (
	"net/netip"

	"github.com/stv0g/go-babel/proto"
)

// 3.2.5. The Source Table
// https://datatracker.ietf.org/doc/html/rfc8966#section-3.2.5

type Source struct {
	Prefix   netip.Prefix
	RouterID proto.RouterID

	Metric int
	SeqNo  uint16
}

type sourceKey struct {
	Prefix   netip.Prefix
	RouterID proto.RouterID
}

type SourceTable Map[sourceKey, *Source]

func (t *SourceTable) Lookup(pfx netip.Prefix, rid proto.RouterID) (*Source, bool) {
	return (*Map[sourceKey, *Source])(t).Lookup(sourceKey{
		Prefix:   pfx,
		RouterID: rid,
	})
}

func (t *SourceTable) Insert(s *Source) {
	(*Map[sourceKey, *Source])(t).Insert(sourceKey{
		s.Prefix,
		s.RouterID,
	}, s)
}
