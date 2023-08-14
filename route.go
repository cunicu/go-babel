// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package babel

import (
	"cunicu.li/go-babel/proto"
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
