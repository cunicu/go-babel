// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package babel

import (
	"net/netip"

	"cunicu.li/go-babel/proto"
)

// 3.2.5. The Source Table
// https://datatracker.ietf.org/doc/html/rfc8966#section-3.2.5

type Source struct {
	Prefix   netip.Prefix
	RouterID proto.RouterID

	Metric int
	SeqNo  uint16
}
