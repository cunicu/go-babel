// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package proto

import (
	"math"
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
	TypePad1                  ValueType = iota // RFC 8966
	TypePadN                                   // RFC 8966
	TypeAcknowledgmentRequest                  // RFC 8966
	TypeAcknowledgment                         // RFC 8966
	TypeHello                                  // RFC 8966
	TypeIHU                                    // RFC 8966
	TypeRouterID                               // RFC 8966
	TypeNextHop                                // RFC 8966
	TypeUpdate                                 // RFC 8966
	TypeRouteRequest                           // RFC 8966
	TypeSeqnoRequest                           // RFC 8966
	TypeTSPC                                   // RFC 7298
	TypeHMAC                                   // RFC 7298
	_                                          // previously used in an early version of draft-ietf-babel-source-specific
	_                                          // previously used in an early version of draft-ietf-babel-source-specific
	_                                          // previously used in an early version of draft-ietf-babel-source-specific
	TypeMAC                                    // RFC 8967
	TypePC                                     // RFC 8967
	TypeChallengeRequest                       // RFC 8967
	TypeChallengeReply                         // RFC 8967

	// TypeInvalid is specified by any RFC and just used internal to
	// represent an invalid type
	TypeInvalid ValueType = math.MaxUint8
)

// Babel Sub-TLV Types
// https://www.iana.org/assignments/babel/babel.xhtml#sub-tlv-types
const (
	SubTypePad1      ValueType = iota // RFC 8966
	SubTypePadN                       // RFC 8966
	SubTypeDiversity                  // draft-chroboczek-babel-diversity-routing
	SubTypeTimestamp                  // draft-jonglez-babel-rtt-extension

	SubTypeSourcePrefix ValueType = 128 //	RFC 9079
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

func IsUrgent(v Value) bool {
	switch v.(type) {
	default:
		return false
	}
}

func ValuesType(v Value) ValueType {
	switch v.(type) {
	case *Pad1:
		return TypePad1
	case *PadN:
		return TypePadN
	case *AcknowledgmentRequest:
		return TypeAcknowledgmentRequest
	case *Acknowledgment:
		return TypeAcknowledgment
	case *Hello:
		return TypeHello
	case *IHU:
		return TypeIHU
	case *RouterIDValue:
		return TypeRouterID
	case *NextHop:
		return TypeNextHop
	case *Update:
		return TypeUpdate
	case *RouteRequest:
		return TypeRouteRequest
	case *SeqnoRequest:
		return TypeSeqnoRequest
	default:
		panic(ErrUnsupportedValue)
	}
}

func (t ValueType) String() string {
	switch t {
	case TypePad1:
		return "Pad1"
	case TypePadN:
		return "PadN"
	case TypeAcknowledgmentRequest:
		return "AcknowledgmentRequest"
	case TypeAcknowledgment:
		return "Acknowledgment"
	case TypeHello:
		return "Hello"
	case TypeIHU:
		return "IHU"
	case TypeRouterID:
		return "RouterID"
	case TypeNextHop:
		return "NextHop"
	case TypeUpdate:
		return "Update"
	case TypeRouteRequest:
		return "RouteRequest"
	case TypeSeqnoRequest:
		return "SeqnoRequest"
	case TypeTSPC:
		return "TSPC"
	case TypeHMAC:
		return "HMAC"
	case TypeMAC:
		return "MAC"
	case TypePC:
		return "PC"
	case TypeChallengeRequest:
		return "ChallengeRequest"
	case TypeChallengeReply:
		return "ChallengeReply"
	default:
		return "<Unknown>"
	}
}
