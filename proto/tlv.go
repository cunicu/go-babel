// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package proto

import (
	"encoding/hex"
	"math"
	"net/netip"
	"time"

	"golang.org/x/exp/slog"
)

const (
	ValueHeaderLength = 2
)

// Value represents a Type-Length-Value (TLV)
// See also: 4.3. TLV Format
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.3
type Value any

// SubValue represents a sub-TLV
// See also: 4.4. Sub-TLV Format
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.4
type SubValue any

// 5. IANA Considerations
// https://datatracker.ietf.org/doc/html/rfc8966#section-5

// Babel TLV Types
// https://www.iana.org/assignments/babel/babel.xhtml#tlv-types
const (
	TypePad1                  ValueType = iota // RFC8966
	TypePadN                                   // RFC8966
	TypeAcknowledgmentRequest                  // RFC8966
	TypeAcknowledgment                         // RFC8966
	TypeHello                                  // RFC8966
	TypeIHU                                    // RFC8966
	TypeRouterID                               // RFC8966
	TypeNextHop                                // RFC8966
	TypeUpdate                                 // RFC8966
	TypeRouteRequest                           // RFC8966
	TypeSeqnoRequest                           // RFC8966
	TypeTSPC                                   // RFC7298
	TypeHMAC                                   // RFC7298
	_                                          // previously used in an early version of draft-ietf-babel-source-specific
	_                                          // previously used in an early version of draft-ietf-babel-source-specific
	_                                          // previously used in an early version of draft-ietf-babel-source-specific
	TypeMAC                                    // RFC8967
	TypePC                                     // RFC8967
	TypeChallengeRequest                       // RFC8967
	TypeChallengeReply                         // RFC8967

	// TypeInvalid is specified by any RFC and just used internal to
	// represent an invalid type
	TypeInvalid ValueType = math.MaxUint8
)

// Babel Sub-TLV Types
// https://www.iana.org/assignments/babel/babel.xhtml#sub-tlv-types
const (
	SubTypePad1      ValueType = iota // RFC8966
	SubTypePadN                       // RFC8966
	SubTypeDiversity                  // draft-chroboczek-babel-diversity-routing
	SubTypeTimestamp                  // draft-jonglez-babel-rtt-extension

	SubTypeSourcePrefix ValueType = 128 //	RFC9079
)

// Flags for Hello TLV
// https://datatracker.ietf.org/doc/html/rfc8966#name-hello
// https://www.iana.org/assignments/babel/babel.xhtml#hello
const (
	FlagHelloUnicast uint16 = 0x8000
)

// Flags for Update TLV
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.9
// https://www.iana.org/assignments/babel/babel.xhtml#flags
const (
	FlagUpdatePrefix   uint8 = 0x80
	FlagUpdateRouterID uint8 = 0x40
)

// 4.6.1. Pad1
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.1
type Pad1 struct{}

// 4.6.2. PadN
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.2
type PadN struct {
	N int
}

func (p *PadN) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("n", p.N))
}

// 4.6.3. Acknowledgment Request
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.3
type AcknowledgmentRequest struct {
	Opaque   uint16        // An arbitrary value that will be echoed in the receiver's Acknowledgment TLV.
	Interval time.Duration // A time interval after which the sender will assume that this packet has been lost. This MUST NOT be 0. The receiver MUST send an Acknowledgment TLV before this time has elapsed (with a margin allowing for propagation time).
}

func (a *AcknowledgmentRequest) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("opaque", a.Opaque))
}

// 4.6.4. Acknowledgment
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.4
type Acknowledgment struct {
	Opaque uint16 // Set to the Opaque value of the Acknowledgment Request that prompted this Acknowledgment.
}

func (a *Acknowledgment) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("opaque", a.Opaque))
}

// 4.6.5. Hello
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.5
type Hello struct {
	Flags    uint16         // The individual bits of this field specify special handling of this TLV (see FlagHello*).
	Seqno    SequenceNumber // If the Unicast flag is set, this is the value of the sending node's outgoing Unicast Hello seqno for this neighbour. Otherwise, it is the sending node's outgoing Multicast Hello seqno for this interface.
	Interval time.Duration  // If nonzero, this is an upper bound, on the time after which the sending node will send a new scheduled Hello TLV with the same setting of the Unicast flag. If this is 0, then this Hello represents an unscheduled Hello and doesn't carry any new information about times at which Hellos are sent.

	// Sub-TLVs
	Timestamp *TimestampHello
}

func (h *Hello) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.Any("flags", h.Flags),
		slog.Any("seqno", h.Seqno),
		slog.Any("intv", h.Interval),
	}

	if h.Timestamp != nil {
		attrs = append(attrs, slog.Any("timestamp", *h.Timestamp))
	}

	return slog.GroupValue(attrs...)
}

// 4.6.6. IHU
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.6
type IHU struct {
	RxCost   uint16        // The rxcost according to the sending node of the interface whose address is specified in the Address field. The value FFFF hexadecimal (infinity) indicates that this interface is unreachable.
	Interval time.Duration // An upper bound, on the time after which the sending node will send a new IHU; this MUST NOT be 0. The receiving node will use this value in order to compute a hold time for this symmetric association.
	Address  netip.Addr    // The address of the destination node, in the format specified by the AE field. Address compression is not allowed.

	// Sub-TLVs
	Timestamp *TimestampIHU
}

func (i *IHU) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.Any("rxcost", i.RxCost),
		slog.Duration("intv", i.Interval),
		slog.Any("addr", i.Address),
	}

	if i.Timestamp != nil {
		attrs = append(attrs,
			slog.Any("ts", *i.Timestamp))
	}

	return slog.GroupValue(attrs...)
}

// 4.6.7. Router-Id
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.7
type RouterIDValue struct {
	RouterID RouterID // The router-id for routes advertised in subsequent Update TLVs. This MUST NOT consist of all zeroes or all ones.
}

func (r *RouterIDValue) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("rid", hex.EncodeToString(r.RouterID[:])))
}

// 4.6.8. Next Hop
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.8
type NextHop struct {
	NextHop netip.Addr // The next-hop address advertised by subsequent Update TLVs for this address family.
}

func (n *NextHop) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("nh", n.NextHop))
}

// 4.6.9. Update
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.9
type Update struct {
	Flags    uint8          // The individual bits of this field specify special handling of this TLV (see FlagUpdate*).
	Interval Interval       // An upper bound, expressed in centiseconds, on the time after which the sending node will send a new update for this prefix. This MUST NOT be 0. The receiving node will use this value to compute a hold time for the route table entry. The value FFFF hexadecimal (infinity) expresses that this announcement will not be repeated unless a request is received (Section 3.8.2.3).
	Seqno    SequenceNumber // The originator's sequence number for this update.
	Metric   Metric         // The sender's metric for this route. The value FFFF hexadecimal (infinity) means that this is a route retraction.
	Prefix   netip.Prefix   // The prefix being advertised. This field's size is (Plen/8 - Omitted) rounded upwards.

	// The following fields are not actually encoded in an update TLV.
	// Instead are derived from the parser state / preceding TLVs.
	RouterID RouterID // Taken from a previous RouterID TLV
	NextHop  Address  // Taken from a previous NextHop TLV

	// Sub-TLVs
	SourcePrefix *Prefix
}

func (u *Update) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.Any("flags", u.Flags),
		slog.Duration("intv", u.Interval),
		slog.Any("seqno", u.Seqno),
		slog.Any("metric", u.Metric),
		slog.Any("pfx", u.Prefix),
		slog.Any("rid", u.RouterID),
		slog.Any("nh", u.NextHop),
	}

	if u.SourcePrefix != nil {
		attrs = append(attrs,
			slog.Any("src_prefix", *u.SourcePrefix))
	}

	return slog.GroupValue(attrs...)
}

// 4.6.10. Route Request
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.10
type RouteRequest struct {
	Prefix netip.Prefix // The prefix being requested. This field's size is Plen/8 rounded upwards.

	// Sub-TLVs
	SourcePrefix *Prefix
}

func (r *RouteRequest) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.Any("pfx", r.Prefix),
	}

	if r.SourcePrefix != nil {
		attrs = append(attrs,
			slog.Any("src_pfx", *r.SourcePrefix))
	}

	return slog.GroupValue(attrs...)
}

// 4.6.11. Seqno Request
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.11
type SeqnoRequest struct {
	Seqno    SequenceNumber // The sequence number that is being requested.
	HopCount uint8          // The maximum number of times that this TLV may be forwarded, plus 1. This MUST NOT be 0.
	RouterID RouterID       // The Router-Id that is being requested. This MUST NOT consist of all zeroes or all ones.
	Prefix   netip.Prefix   // The prefix being requested. This field's size is Plen/8 rounded upwards.

	// Sub-TLVs
	SourcePrefix *Prefix
}

func (s *SeqnoRequest) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.Any("seqno", s.Seqno),
		slog.Any("hopcnt", s.HopCount),
		slog.Any("rid", s.RouterID),
		slog.Any("pfx", s.Prefix),
	}

	if s.SourcePrefix != nil {
		attrs = append(attrs,
			slog.Any("src_pfx", *s.SourcePrefix))
	}

	return slog.GroupValue(attrs...)
}

// 3.1.  Timestamp sub-TLV in Hello TLVs
// https://datatracker.ietf.org/doc/html/draft-ietf-babel-rtt-extension-00#section-3.1
type TimestampHello struct {
	Transmit Timestamp
}

func (t *TimestampHello) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("tx", t.Transmit))
}

// 3.2.  Timestamp sub-TLV in IHU TLVs
// https://datatracker.ietf.org/doc/html/draft-ietf-babel-rtt-extension-00#section-3.2
type TimestampIHU struct {
	Origin  Timestamp
	Receive Timestamp
}

func (t *TimestampIHU) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("orign", t.Origin),
		slog.Any("rx", t.Receive))
}

func IsUrgent(v Value) bool {
	switch v.(type) {
	default:
		return false
	}
}
