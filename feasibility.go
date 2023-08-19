package babel

import "github.com/stv0g/go-babel/proto"

type FeasibilityDistance struct {
	SeqNo  proto.SequenceNumber
	Metric uint
}

// Less checks if the the feasibility is better than the provided one
//
// https://datatracker.ietf.org/doc/html/rfc8966#section-3.5.1
func (d FeasibilityDistance) IsBetter(o FeasibilityDistance) bool {
	return proto.SeqnoLess(o.SeqNo, d.SeqNo) || (d.SeqNo == o.SeqNo && d.Metric < o.Metric)
}
