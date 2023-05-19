// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package proto

import (
	"crypto/rand"
	"fmt"
	"net/netip"
	"time"
)

type (
	// 4.1.2. Interval
	// https://datatracker.ietf.org/doc/html/rfc8966#section-4.1.2
	Interval = time.Duration

	// 4.1.3. Router-Id
	// https://datatracker.ietf.org/doc/html/rfc8966#section-4.1.3
	RouterID = [8]byte

	// 4.1.4. Address
	// https://datatracker.ietf.org/doc/html/rfc8966#section-4.1.4
	Address = netip.Addr

	// 4.1.5. Prefixes
	// https://datatracker.ietf.org/doc/html/rfc8966#section-4.1.5
	Prefix = netip.Prefix

	Metric         = uint16
	SequenceNumber = uint16

	// Timestamp for Babel RTT extension
	// https://datatracker.ietf.org/doc/html/draft-ietf-babel-rtt-extension-00#section-3
	Timestamp = uint32 // in microseconds

	AddressEncoding = uint8
	AddressFamily   = int

	ValueType uint8
)

const (
	Retraction Metric = 0xffff
)

var (
	RouterIDAllZeros = RouterID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	RouterIDAllOnes  = RouterID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	RouterIDUnspecified = RouterIDAllZeros
)

const (
	AddressFamilyUnspecified AddressFamily = iota
	AddressFamilyIPv4
	AddressFamilyIPv6
)

// Babel Address Encodings
// https://www.iana.org/assignments/babel/babel.xhtml#ae
const (
	AddressEncodingWildcard      AddressEncoding = 0 // RFC 8966: The value is 0 octets long.
	AddressEncodingIPv4          AddressEncoding = 1 // RFC 8966: Compression is allowed. 4 octets or less.
	AddressEncodingIPv6          AddressEncoding = 2 // RFC 8966: Compression is allowed. 16 octets or less.
	AddressEncodingIPv6LinkLocal AddressEncoding = 3 // RFC 8966: Compression is not allowed. The value is 8 octets long, a prefix of fe80::/64 is implied.
	AddressEncodingIPv4inIPv6    AddressEncoding = 4 // RFC 9229: IPv4 Routes with an IPv6 Next Hop
)

// GenerateRouterID generates a random router ID
func GenerateRouterID() (RouterID, error) {
	b := make([]byte, 8)

	n, err := rand.Read(b)
	if err != nil {
		return RouterIDUnspecified, err
	} else if n != 8 {
		return RouterIDUnspecified, fmt.Errorf("failed to generated router id")
	}

	return *(*RouterID)(b), nil
}

// IsMandatory checks whether the sub-TLV type is mandatory
func (t ValueType) IsMandatory() bool {
	return t&0x80 != 0
}

func (t ValueType) IsSubType() bool {
	switch t {
	case SubTypePad1, SubTypePadN, SubTypeDiversity, SubTypeTimestamp, SubTypeSourcePrefix:
		return true
	default:
		return false
	}
}

func (t ValueType) IsTrailerType() bool {
	switch t {
	case TypePad1, TypePadN, TypeMAC:
		return true
	default:
		return false
	}
}

func RouterIDFromAddr(addr netip.Addr) RouterID {
	i := RouterIDAllZeros
	s := addr.AsSlice()

	switch {
	case addr.Is4():
		copy(i[4:], s)
	case addr.Is4In6():
		copy(i[4:], s[12:])
	case addr.Is6():
		copy(i[:], s[8:])
	}

	return i
}

func SeqnoDistance(a, b SequenceNumber) int16 {
	return (int16)(b - a)
}

func SeqnoAbsDistance(a, b SequenceNumber) int16 {
	if d := SeqnoDistance(a, b); d > 0 {
		return d
	} else {
		return -d
	}
}

func SeqnoLess(a, b SequenceNumber) bool {
	return SeqnoDistance(a, b) > 0
}
