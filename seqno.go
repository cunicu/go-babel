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

type pendingSeqNoRequestKey struct {
	Prefix   proto.Prefix
	RouterID proto.RouterID
}

type PendingSeqNoRequestTable Map[pendingSeqNoRequestKey, *PendingSeqNoRequest]

func (t *PendingSeqNoRequestTable) Lookup(pfx proto.Prefix, rid proto.RouterID) (*PendingSeqNoRequest, bool) {
	return (*Map[pendingSeqNoRequestKey, *PendingSeqNoRequest])(t).Lookup(pendingSeqNoRequestKey{
		Prefix:   pfx,
		RouterID: rid,
	})
}

func (t *PendingSeqNoRequestTable) Insert(req *PendingSeqNoRequest) {
	(*Map[pendingSeqNoRequestKey, *PendingSeqNoRequest])(t).Insert(pendingSeqNoRequestKey{
		Prefix:   req.Prefix,
		RouterID: req.RouterID,
	}, req)
}
