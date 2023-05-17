// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package babel

import (
	"github.com/stv0g/go-babel/proto"
)

// 3.2.7. The Table of Pending Seqno Requests
// https://datatracker.ietf.org/doc/html/rfc8966#section-3.2.7

type PendingSeqNoRequest struct {
	Prefix   proto.Prefix
	RouterID proto.RouterID

	Neighbour *Neighbour
	Resent    int
}
