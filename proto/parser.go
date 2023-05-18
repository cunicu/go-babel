// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package proto

import (
	"encoding/binary"
	"errors"
	"math"
	"net"
	"net/netip"
	"time"
)

// Parser implements the protocol en/decoding of Babel packets
// It keeps internal state which is required for compressing
// prefixes and other details for Update TLVs.
//
// See also: 4.5. Parser State and Encoding of Updates
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.5
type Parser struct {
	CurrentDefaultPrefix map[AddressEncoding]Address
	CurrentNextHop       map[AddressFamily]Address
	CurrentRouterID      RouterID
}

// Reset resets the internal parser state
func (p *Parser) Reset() {
	p.CurrentDefaultPrefix = map[AddressEncoding]Address{}
	p.CurrentNextHop = map[AddressFamily]Address{}
	p.CurrentRouterID = RouterIDUnspecified
}

// Packet
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.2

// PacketLength calculates the length of the required buffer
// to encode the packet without actually encoding it.
// This should be used for pre-allocate an appropriately sized
// buffer.
func (p *Parser) PacketLength(pkt *Packet) uint16 {
	return PacketHeaderLength + p.ValuesLength(pkt.Body) + p.ValuesLength(pkt.Trailer)
}

// Packet attempts to decode a packet from the provided buffer.
// It returns a advanced buffer slice starting at the end
// of the parsed packet.
func (p *Parser) Packet(b []byte) ([]byte, *Packet, error) {
	var err error
	var magic, version uint8
	var bodyLength uint16

	pkt := &Packet{}

	if b, magic, err = p.uint8(b); err != nil {
		return nil, nil, err
	} else if magic != PacketHeaderMagic {
		return nil, nil, ErrInvalidMagic
	}

	if b, version, err = p.uint8(b); err != nil {
		return nil, nil, err
	} else if version != PacketHeaderVersion {
		return nil, nil, ErrUnsupportedVersion
	}

	b, bodyLength, err = p.uint16(b)
	if err != nil {
		return nil, nil, err
	}

	if len(b) < int(bodyLength) {
		return nil, nil, ErrTooShort
	}

	pkt.Body = nil
	pkt.Trailer = nil

	if b, err = p.forEachValue(b[:bodyLength], func(t ValueType, b []byte) ([]byte, error) {
		if b, v, err := p.valuePayload(t, b); err != nil {
			return nil, err
		} else {
			pkt.Body = append(pkt.Body, v)
			return b, nil
		}
	}); err != nil {
		return nil, nil, err
	}

	if b, err = p.forEachValue(b, func(t ValueType, b []byte) ([]byte, error) {
		if !t.IsTrailerType() {
			return nil, ErrInvalidValueForTrailer
		} else if b, v, err := p.valuePayload(t, b); err != nil {
			return nil, err
		} else {
			pkt.Trailer = append(pkt.Trailer, v)
			return b, nil
		}
	}); err != nil {
		return nil, nil, err
	}

	return b, pkt, nil
}

// AppendPacket encodes a packet by appending it to the provided
// buffer. Ideally the buffer should be pre-allocated with a
// capacity determined by PacketLength()
func (p *Parser) AppendPacket(b []byte, pkt *Packet) []byte {
	o := len(b)

	b = p.appendUint8(b, PacketHeaderMagic)
	b = p.appendUint8(b, PacketHeaderVersion)
	b = p.appendUint16(b, 0) // Placeholder: length
	b = p.AppendValues(b, pkt.Body)
	b = p.AppendValues(b, pkt.Trailer)

	// Fill in body length
	bodyLength := len(b) - o - PacketHeaderLength
	binary.BigEndian.PutUint16(b[o+2:], uint16(bodyLength))

	return b
}

// Generic integers
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.1.1

func (p *Parser) uint8(b []byte) ([]byte, uint8, error) {
	if len(b) < 1 {
		return nil, 0, ErrTooShort
	}
	return b[1:], b[0], nil
}

func (p *Parser) uint16(b []byte) ([]byte, uint16, error) {
	if len(b) < 2 {
		return nil, 0, ErrTooShort
	}
	return b[2:], binary.BigEndian.Uint16(b), nil
}

func (p *Parser) uint32(b []byte) ([]byte, uint32, error) {
	if len(b) < 4 {
		return nil, 0, ErrTooShort
	}
	return b[4:], binary.BigEndian.Uint32(b), nil
}

func (p *Parser) uint64(b []byte) ([]byte, uint64, error) {
	if len(b) < 8 {
		return nil, 0, ErrTooShort
	}
	return b[8:], binary.BigEndian.Uint64(b), nil
}

func (p *Parser) appendUint8(b []byte, v uint8) []byte {
	return append(b, v)
}

func (p *Parser) appendUint16(b []byte, v uint16) []byte {
	return binary.BigEndian.AppendUint16(b, v)
}

func (p *Parser) appendUint32(b []byte, v uint32) []byte {
	return binary.BigEndian.AppendUint32(b, v)
}

func (p *Parser) appendUint64(b []byte, v uint64) []byte {
	return binary.BigEndian.AppendUint64(b, v)
}

// TLV Values
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.3

// ValueLength returns the number of octets of an TLV including the type / length fields.
func (p *Parser) ValueLength(v Value) (l int) {
	l = ValueHeaderLength

	switch v := v.(type) {
	case *Pad1:
		return 1
	case *PadN:
		l += v.N
	case *AcknowledgmentRequest:
		l += 6
	case *Acknowledgment:
		l += 2
	case *Hello:
		l += 6
		if v.Timestamp != nil {
			l += ValueHeaderLength + 4
		}
	case *IHU:
		l += 6 + p.addressLength(v.Address)
		if v.Timestamp != nil {
			l += ValueHeaderLength + 2*4
		}
	case *RouterIDValue:
		l += 2 + 8
	case *NextHop:
		l += 2 + p.addressLength(v.NextHop)
	case *Update:
		l += 10 + p.prefixLength(v.Prefix, true)
		if v.SourcePrefix != nil {
			l += ValueHeaderLength + 1 + p.prefixLength(*v.SourcePrefix, false)
		}
	case *RouteRequest:
		l += 2 + p.prefixLength(v.Prefix, false)
		if v.SourcePrefix != nil {
			l += ValueHeaderLength + 1 + p.prefixLength(*v.SourcePrefix, false)
		}
	case *SeqnoRequest:
		l += 14 + p.prefixLength(v.Prefix, false)
		if v.SourcePrefix != nil {
			l += ValueHeaderLength + 1 + p.prefixLength(*v.SourcePrefix, false)
		}
	default:
		panic(ErrUnsupportedValue)
	}

	return l
}

func (p *Parser) value(b []byte) ([]byte, Value, ValueType, error) {
	b, typ, length, err := p.valueHeader(b)
	if err != nil {
		return nil, nil, 0, err
	}

	if len(b) < length {
		return nil, nil, 0, ErrTooShort
	}

	b, v, err := p.valuePayload(typ, b[:length])
	return b, v, typ, err
}

func (p *Parser) valuePayload(t ValueType, b []byte) ([]byte, Value, error) {
	switch t {
	case TypePad1:
		return p.pad1(b)
	case TypePadN:
		return p.padN(b)
	case TypeAcknowledgmentRequest:
		return p.acknowledgmentRequest(b)
	case TypeAcknowledgment:
		return p.acknowledgment(b)
	case TypeHello:
		return p.hello(b)
	case TypeIHU:
		return p.ihu(b)
	case TypeRouterID:
		return p.routerIDValue(b)
	case TypeNextHop:
		return p.nextHop(b)
	case TypeUpdate:
		return p.update(b)
	case TypeRouteRequest:
		return p.routeRequest(b)
	case TypeSeqnoRequest:
		return p.seqnoRequest(b)
	default:
		return nil, nil, ErrUnsupportedValue
	}
}

func (p *Parser) AppendValue(b []byte, v Value) []byte {
	switch v := v.(type) {
	case *Pad1:
		return p.appendUint8(b, uint8(TypePad1))
	case *PadN:
		return p.appendValueHeader(b, TypePadN, func(b []byte) []byte {
			return p.appendPadN(b, v)
		})
	case *AcknowledgmentRequest:
		return p.appendValueHeader(b, TypeAcknowledgmentRequest, func(b []byte) []byte {
			return p.appendAcknowledgmentRequest(b, v)
		})
	case *Acknowledgment:
		return p.appendValueHeader(b, TypeAcknowledgment, func(b []byte) []byte {
			return p.appendAcknowledgment(b, v)
		})
	case *Hello:
		return p.appendValueHeader(b, TypeHello, func(b []byte) []byte {
			return p.appendHello(b, v)
		})
	case *IHU:
		return p.appendValueHeader(b, TypeIHU, func(b []byte) []byte {
			return p.appendIHU(b, v)
		})
	case *RouterIDValue:
		return p.appendValueHeader(b, TypeRouterID, func(b []byte) []byte {
			return p.appendRouterIDValue(b, v)
		})
	case *NextHop:
		return p.appendValueHeader(b, TypeNextHop, func(b []byte) []byte {
			return p.appendNextHop(b, v)
		})
	case *Update:
		return p.appendValueHeader(b, TypeUpdate, func(b []byte) []byte {
			return p.appendUpdate(b, v)
		})
	case *RouteRequest:
		return p.appendValueHeader(b, TypeRouteRequest, func(b []byte) []byte {
			return p.appendRouteRequest(b, v)
		})
	case *SeqnoRequest:
		return p.appendValueHeader(b, TypeSeqnoRequest, func(b []byte) []byte {
			return p.appendSeqnoRequest(b, v)
		})
	default:
		panic(ErrInvalidValueType)
	}
}

func (p *Parser) valueHeader(b []byte) ([]byte, ValueType, int, error) {
	b, t, err := p.uint8(b)
	if err != nil {
		return nil, 0, 0, err
	}

	typ := ValueType(t)
	if typ == TypePad1 {
		return b, typ, 0, nil
	}

	b, length, err := p.uint8(b)
	if err != nil {
		return nil, 0, 0, err
	}

	return b, typ, int(length), nil
}

func (p *Parser) appendValueHeader(b []byte, t ValueType, cb func([]byte) []byte) []byte {
	o := len(b)

	b = p.appendUint8(b, uint8(t))
	b = p.appendUint8(b, 0) // Placeholder: length

	b = cb(b)

	length := len(b) - o - ValueHeaderLength
	b[o+1] = uint8(length)

	return b
}

func (p *Parser) ValuesLength(vs []Value) uint16 {
	length := uint16(0)
	for _, v := range vs {
		length += uint16(p.ValueLength(v))
	}
	return length
}

func (p *Parser) forEachValue(b []byte, cb func(t ValueType, b []byte) ([]byte, error)) ([]byte, error) {
	var t ValueType
	var l int
	var err error

	for len(b) > 0 {
		b, t, l, err = p.valueHeader(b)
		if err != nil {
			return nil, err
		} else if len(b) < l {
			return nil, ErrTooShort
		} else if c, err := cb(t, b[:l]); err != nil {
			return nil, err
		} else if len(c) > 0 {
			return nil, ErrTooLong
		}

		b = b[l:]
	}

	return b, nil
}

func (p *Parser) forEachSubValue(b []byte, cb func(t ValueType, b []byte) ([]byte, error)) ([]byte, error) {
	return p.forEachValue(b, func(t ValueType, b []byte) ([]byte, error) {
		var err error
		if b, err = cb(t, b); err != nil {
			if errors.Is(err, ErrUnsupportedValue) && t.IsMandatory() {
				return nil, ErrUnsupportedButMandatoryValue
			} else {
				return nil, err
			}
		}

		return b, nil
	})
}

func (p *Parser) AppendValues(b []byte, vs []Value) []byte {
	for _, v := range vs {
		b = p.AppendValue(b, v)
	}

	return b
}

// Interval
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.1.2

func (p *Parser) interval(b []byte) ([]byte, time.Duration, error) {
	b, i, err := p.uint16(b)
	if err != nil {
		return nil, 0, err
	}
	return b, time.Duration(i*10) * time.Millisecond, nil
}

func (p *Parser) appendInterval(b []byte, i time.Duration) []byte {
	centisecs := uint16(100 * i.Seconds())
	return p.appendUint16(b, centisecs)
}

// Router ID
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.1.3

func (p *Parser) routerID(b []byte) ([]byte, RouterID, error) {
	if len(b) < 8 {
		return nil, RouterIDUnspecified, ErrTooShort
	}

	rid := *(*RouterID)(b)

	if rid == RouterIDUnspecified || rid == RouterIDAllOnes {
		return nil, RouterIDUnspecified, ErrInvalidRouterID
	}

	return b, rid, nil
}

func (p *Parser) appendRouterID(b []byte, rid RouterID) []byte {
	if rid == RouterIDUnspecified || rid == RouterIDAllOnes {
		panic(ErrInvalidRouterID)
	}

	return append(b, rid[:]...)
}

// Address
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.1.4

func (p *Parser) addressLength(addr Address) int {
	switch addressEncoding(&addr) {
	case AddressEncodingWildcard:
		return 0
	case AddressEncodingIPv4, AddressEncodingIPv4inIPv6:
		return net.IPv4len
	case AddressEncodingIPv6:
		return net.IPv6len
	case AddressEncodingIPv6LinkLocal:
		return 8
	default:
		panic(ErrInvalidAddress)
	}
}

func (p *Parser) address(b []byte, ae AddressEncoding, omitted uint8, plen int8) ([]byte, Address, error) {
	switch ae {
	case AddressEncodingWildcard:
		return b, netip.IPv6Unspecified(), nil

	case AddressEncodingIPv4, AddressEncodingIPv6:
		var alen, rplen uint8
		if ae == AddressEncodingIPv6 {
			alen = net.IPv6len
		} else {
			alen = net.IPv4len
		}

		if plen < 0 {
			rplen = alen * 8
		} else {
			rplen = uint8(plen)
		}

		blen := rplen/8 - omitted
		if rplen%8 != 0 { // Round upwards
			blen++
		}

		if len(b) < int(blen) {
			return nil, Address{}, ErrTooShort
		}

		abuf := make([]byte, alen)

		if omitted > 0 {
			dpfx, ok := p.CurrentDefaultPrefix[ae]
			if !ok {
				return nil, Address{}, ErrMissingDefaultPrefix
			}

			copy(abuf[0:], dpfx.AsSlice()[:omitted])
		}

		copy(abuf[omitted:], b[:blen])

		// If plen is not a multiple of 8, then any bits beyond plen
		// (i.e., the low-order (8 - plen % 8) bits of the last octet) are cleared
		if mod := rplen % 8; mod != 0 {
			mask := math.MaxUint8 << (8 - mod)
			abuf[rplen/8] &= uint8(mask)
		}

		if a, ok := netip.AddrFromSlice(abuf); !ok {
			return nil, Address{}, ErrInvalidAddress
		} else {
			return b[blen:], a, nil
		}

	case AddressEncodingIPv6LinkLocal:
		if len(b) < 8 {
			return nil, Address{}, ErrTooShort
		}

		abuf := make([]byte, 16)
		copy(abuf[0:], []byte{0xfe, 0x80})
		copy(abuf[8:], b[:8])

		if a, ok := netip.AddrFromSlice(abuf); !ok {
			return nil, Address{}, ErrInvalidAddress
		} else {
			return b[8:], a, nil
		}

	default:
		return nil, Address{}, ErrInvalidAddress
	}
}

func (p *Parser) appendAddress(b []byte, addr Address, plen int8) ([]byte, AddressEncoding) {
	ae := addressEncoding(&addr)

	switch ae {
	case AddressEncodingIPv6LinkLocal:
		a := addr.As16()
		return append(b, a[8:]...), ae
	case AddressEncodingWildcard:
		return b, ae
	case AddressEncodingIPv4, AddressEncodingIPv6:
		var alen, rplen uint8
		if ae == AddressEncodingIPv6 {
			alen = net.IPv6len
		} else {
			alen = net.IPv4len
		}

		if plen < 0 {
			rplen = alen * 8
		} else {
			rplen = uint8(plen)
		}

		blen := rplen / 8
		if rplen%8 != 0 { // Round upwards
			blen++
		}

		a := addr.AsSlice()
		return append(b, a[:blen]...), ae
	default:
		panic(ErrInvalidAddress)
	}
}

// Prefix
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.1.5

func (p *Parser) prefixLength(pfx Prefix, compress bool) int {
	blen := pfx.Bits() / 8
	if pfx.Bits()%8 != 0 {
		blen++
	}

	// if compress {
	// TODO: Support prefix compression
	// }

	return blen
}

func (p *Parser) prefix(b []byte, ae AddressEncoding, plen, omitted uint8) ([]byte, Prefix, error) {
	b, addr, err := p.address(b, ae, omitted, int8(plen))
	if err != nil {
		return nil, Prefix{}, err
	}

	return b, netip.PrefixFrom(addr, int(plen)), nil
}

func (p *Parser) appendPrefix(b []byte, pfx Prefix, compress bool) ([]byte, AddressEncoding, uint8, uint8) {
	// TODO: Support prefix compression for update TLVs
	b, ae := p.appendAddress(b, pfx.Addr(), int8(pfx.Bits()))

	return b, ae, uint8(pfx.Bits()), 0
}

// Pad1
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.1

func (p *Parser) pad1(b []byte) ([]byte, *Pad1, error) {
	return b, &Pad1{}, nil
}

// TODO: Use function
//
//nolint:unused
func (p *Parser) appendPad1(b []byte, v *Pad1) []byte {
	return b
}

// PadN
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.2

func (p *Parser) padN(b []byte) ([]byte, *PadN, error) {
	v := &PadN{
		N: len(b),
	}

	return b[v.N:], v, nil
}

func (p *Parser) appendPadN(b []byte, v *PadN) []byte {
	for i := 0; i < v.N; i++ {
		b = append(b, 0)
	}

	return b
}

// AcknowledgmentRequest
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.3

func (p *Parser) acknowledgmentRequest(b []byte) ([]byte, *AcknowledgmentRequest, error) {
	var err error
	v := &AcknowledgmentRequest{}

	if b, _, err = p.uint16(b); err != nil { // Reserved
		return nil, nil, err
	}
	if b, v.Opaque, err = p.uint16(b); err != nil {
		return nil, nil, err
	}
	if b, v.Interval, err = p.interval(b); err != nil {
		return nil, nil, err
	}

	return b, v, nil
}

func (p *Parser) appendAcknowledgmentRequest(b []byte, v *AcknowledgmentRequest) []byte {
	b = p.appendUint16(b, 0) // Reserved
	b = p.appendUint16(b, v.Opaque)
	b = p.appendInterval(b, v.Interval)

	return b
}

// Acknowledgment
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.4

func (p *Parser) acknowledgment(b []byte) ([]byte, *Acknowledgment, error) {
	var err error
	v := &Acknowledgment{}

	if b, v.Opaque, err = p.uint16(b); err != nil {
		return nil, nil, err
	}

	return b, v, nil
}

func (p *Parser) appendAcknowledgment(b []byte, v *Acknowledgment) []byte {
	b = p.appendUint16(b, v.Opaque)

	return b
}

// Hello
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.5

func (p *Parser) hello(b []byte) ([]byte, *Hello, error) {
	var err error
	v := &Hello{}

	if b, v.Flags, err = p.uint16(b); err != nil {
		return nil, nil, err
	}
	if b, v.Seqno, err = p.uint16(b); err != nil {
		return nil, nil, err
	}
	if b, v.Interval, err = p.interval(b); err != nil {
		return nil, nil, err
	}

	// Decode sub-TLVs
	b, err = p.forEachSubValue(b, func(t ValueType, b []byte) ([]byte, error) {
		switch t {
		case SubTypeTimestamp:
			v.Timestamp = &TimestampHello{}
			if b, v.Timestamp.Transmit, err = p.uint32(b); err != nil {
				return nil, err
			}
			return b, nil

		default:
			return nil, ErrUnsupportedValue
		}
	})

	return b, v, err
}

func (p *Parser) appendHello(b []byte, v *Hello) []byte {
	b = p.appendUint16(b, v.Flags)
	b = p.appendUint16(b, v.Seqno)
	b = p.appendInterval(b, v.Interval)

	// Encode sub-TLVs
	if v.Timestamp != nil {
		b = p.appendValueHeader(b, SubTypeTimestamp, func(b []byte) []byte {
			return p.appendUint32(b, v.Timestamp.Transmit)
		})
	}

	return b
}

// IHU
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.6

func (p *Parser) ihu(b []byte) ([]byte, *IHU, error) {
	var err error
	var ae uint8
	v := &IHU{}

	if b, ae, err = p.uint8(b); err != nil {
		return nil, nil, err
	}
	if b, _, err = p.uint8(b); err != nil { // Reserved
		return nil, nil, err
	}
	if b, v.RxCost, err = p.uint16(b); err != nil {
		return nil, nil, err
	}
	if b, v.Interval, err = p.interval(b); err != nil {
		return nil, nil, err
	}
	if b, v.Address, err = p.address(b, ae, 0, -1); err != nil {
		return nil, nil, err
	}

	b, err = p.forEachSubValue(b, func(t ValueType, b []byte) ([]byte, error) {
		switch t {
		case SubTypeTimestamp:
			v.Timestamp = &TimestampIHU{}
			if b, v.Timestamp.Origin, err = p.uint32(b); err != nil {
				return nil, err
			}
			if b, v.Timestamp.Receive, err = p.uint32(b); err != nil {
				return nil, err
			}
			return b, nil

		default:
			return nil, ErrUnsupportedValue
		}
	})

	return b, v, nil
}

func (p *Parser) appendIHU(b []byte, v *IHU) []byte {
	o := len(b)

	b = p.appendUint8(b, 0) // Placeholder: ae
	b = p.appendUint8(b, 0) // Reserved
	b = p.appendUint16(b, v.RxCost)
	b = p.appendInterval(b, v.Interval)
	b, ae := p.appendAddress(b, v.Address, -1)

	b[o+0] = ae

	// Encode sub-TLVs
	if v.Timestamp != nil {
		b = p.appendValueHeader(b, SubTypeTimestamp, func(b []byte) []byte {
			b = p.appendUint32(b, v.Timestamp.Origin)
			b = p.appendUint32(b, v.Timestamp.Receive)

			return b
		})
	}

	return b
}

// RouterID
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.7

func (p *Parser) routerIDValue(b []byte) ([]byte, *RouterIDValue, error) {
	var err error
	v := &RouterIDValue{}

	if b, _, err = p.uint16(b); err != nil { // Reserved
		return nil, nil, err
	}
	if b, v.RouterID, err = p.routerID(b); err != nil {
		return nil, nil, err
	}

	return b, v, nil
}

func (p *Parser) appendRouterIDValue(b []byte, v *RouterIDValue) []byte {
	b = p.appendUint16(b, 0) // Reserved
	b = p.appendRouterID(b, v.RouterID)

	return b
}

// NextHop
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.8

func (p *Parser) nextHop(b []byte) ([]byte, *NextHop, error) {
	var err error
	var ae uint8
	v := &NextHop{}

	if b, ae, err = p.uint8(b); err != nil {
		return nil, nil, err
	}
	if b, _, err = p.uint8(b); err != nil { // Reserved
		return nil, nil, err
	}
	if b, v.NextHop, err = p.address(b, ae, 0, -1); err != nil {
		return nil, nil, err
	}

	return b, v, nil
}

func (p *Parser) appendNextHop(b []byte, v *NextHop) []byte {
	o := len(b)
	b = p.appendUint8(b, 0) // Placeholder: ae
	b = p.appendUint8(b, 0) // Reserved
	b, ae := p.appendAddress(b, v.NextHop, -1)

	b[o+0] = ae

	return b
}

// Update
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.9

func (p *Parser) update(b []byte) ([]byte, *Update, error) {
	var err error
	var ae, plen, omitted uint8
	v := &Update{}

	if b, ae, err = p.uint8(b); err != nil {
		return nil, nil, err
	}
	if b, v.Flags, err = p.uint8(b); err != nil {
		return nil, nil, err
	}
	if b, plen, err = p.uint8(b); err != nil {
		return nil, nil, err
	}
	if b, omitted, err = p.uint8(b); err != nil {
		return nil, nil, err
	}
	if b, v.Interval, err = p.interval(b); err != nil {
		return nil, nil, err
	}
	if b, v.Seqno, err = p.uint16(b); err != nil {
		return nil, nil, err
	}
	if b, v.Metric, err = p.uint16(b); err != nil {
		return nil, nil, err
	}
	if b, v.Prefix, err = p.prefix(b, ae, plen, omitted); err != nil {
		return nil, nil, err
	}

	// Decode sub-TLVs
	if b, err = p.forEachSubValue(b, func(t ValueType, b []byte) ([]byte, error) {
		switch t {
		case SubTypeSourcePrefix:
			var pfx Prefix
			if b, pfx, err = p.sourcePrefix(b, ae); err != nil {
				return nil, err
			} else {
				v.SourcePrefix = &pfx
			}
			return b, nil

		default:
			return nil, ErrUnsupportedValue
		}
	}); err != nil {
		return nil, nil, err
	}

	if v.Flags&FlagUpdateRouterID != 0 {
		switch ae {
		case AddressEncodingIPv4:
			addr := v.Prefix.Addr().As4()
			rid := []byte{0, 0, 0, 0}
			rid = append(rid, addr[:4]...)

			p.CurrentRouterID = *(*RouterID)(rid)
		case AddressEncodingIPv6, AddressEncodingIPv6LinkLocal:
			addr := v.Prefix.Addr().As16()

			p.CurrentRouterID = *(*RouterID)(addr[8:16])
		}
	}

	if v.Flags&FlagUpdatePrefix != 0 {
		p.CurrentDefaultPrefix[ae] = v.Prefix.Addr()
	}

	// Fill in fields from parser state
	af := addressFamilyFromAddressEncoding(ae)

	v.RouterID = p.CurrentRouterID
	v.NextHop = p.CurrentNextHop[af]

	return b, v, nil
}

func (p *Parser) appendUpdate(b []byte, v *Update) []byte {
	o := len(b)

	b = p.appendUint8(b, 0) // Placeholder: ae
	b = p.appendUint8(b, v.Flags)
	b = p.appendUint8(b, 0) // Placeholder: plen
	b = p.appendUint8(b, 0) // Placeholder: omitted
	b = p.appendInterval(b, v.Interval)
	b = p.appendUint16(b, v.Seqno)
	b = p.appendUint16(b, v.Metric)
	b, ae, plen, omitted := p.appendPrefix(b, v.Prefix, true)

	b[o+0] = ae
	b[o+2] = plen
	b[o+3] = omitted

	if v.SourcePrefix != nil {
		b = p.appendSourcePrefix(b, *v.SourcePrefix)
	}

	return b
}

// RouteRequest
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.10

func (p *Parser) routeRequest(b []byte) ([]byte, *RouteRequest, error) {
	var err error
	var ae, plen uint8
	v := &RouteRequest{}

	if b, ae, err = p.uint8(b); err != nil {
		return nil, nil, err
	}
	if b, plen, err = p.uint8(b); err != nil {
		return nil, nil, err
	}
	if b, v.Prefix, err = p.prefix(b, ae, plen, 0); err != nil {
		return nil, nil, err
	}

	// Decode sub-TLVs
	if b, err = p.forEachSubValue(b, func(t ValueType, b []byte) ([]byte, error) {
		switch t {
		case SubTypeSourcePrefix:
			var pfx Prefix
			if b, pfx, err = p.sourcePrefix(b, ae); err != nil {
				return nil, err
			} else {
				v.SourcePrefix = &pfx
			}
			return b, nil

		default:
			return nil, ErrUnsupportedValue
		}
	}); err != nil {
		return nil, nil, err
	}

	return b, v, nil
}

func (p *Parser) appendRouteRequest(b []byte, v *RouteRequest) []byte {
	o := len(b)

	b = p.appendUint8(b, 0) // Placeholder: ae
	b = p.appendUint8(b, 0) // Placeholder: plen
	b, ae, plen, _ := p.appendPrefix(b, v.Prefix, false)

	b[o+0] = ae
	b[o+1] = plen

	if v.SourcePrefix != nil {
		b = p.appendSourcePrefix(b, *v.SourcePrefix)
	}

	return b
}

// SeqnoRequest
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.11

func (p *Parser) seqnoRequest(b []byte) ([]byte, *SeqnoRequest, error) {
	var err error
	var ae, plen uint8
	v := &SeqnoRequest{}

	if b, ae, err = p.uint8(b); err != nil {
		return nil, nil, err
	}
	if b, plen, err = p.uint8(b); err != nil {
		return nil, nil, err
	}
	if b, v.Seqno, err = p.uint16(b); err != nil {
		return nil, nil, err
	}
	if b, v.HopCount, err = p.uint8(b); err != nil {
		return nil, nil, err
	}
	if b, _, err = p.uint8(b); err != nil { // Reserved
		return nil, nil, err
	}
	if b, v.RouterID, err = p.routerID(b); err != nil {
		return nil, nil, err
	}
	if b, v.Prefix, err = p.prefix(b, ae, plen, 0); err != nil {
		return nil, nil, err
	}

	// Decode sub-TLVs
	if b, err = p.forEachSubValue(b, func(t ValueType, b []byte) ([]byte, error) {
		switch t {
		case SubTypeSourcePrefix:
			var pfx Prefix
			if b, pfx, err = p.sourcePrefix(b, ae); err != nil {
				return nil, err
			} else {
				v.SourcePrefix = &pfx
			}
			return b, nil

		default:
			return nil, ErrUnsupportedValue
		}
	}); err != nil {
		return nil, nil, err
	}

	return b, v, nil
}

func (p *Parser) appendSeqnoRequest(b []byte, v *SeqnoRequest) []byte {
	o := len(b)

	b = p.appendUint8(b, 0) // Placeholder: ae
	b = p.appendUint8(b, 0) // Placeholder: plen
	b = p.appendUint16(b, v.Seqno)
	b = p.appendUint8(b, v.HopCount)
	b = p.appendUint8(b, 0) // Reserved
	b = p.appendRouterID(b, v.RouterID)
	b, ae, plen, _ := p.appendPrefix(b, v.Prefix, false)

	b[o+0] = ae
	b[o+1] = plen

	if v.SourcePrefix != nil {
		b = p.appendSourcePrefix(b, *v.SourcePrefix)
	}

	return b
}

// SourcePrefix sub-TLV

func (p *Parser) sourcePrefix(b []byte, ae AddressEncoding) ([]byte, Prefix, error) {
	b, plen, err := p.uint8(b)
	if err != nil {
		return nil, Prefix{}, err
	}

	return p.prefix(b, ae, plen, 0)
}

func (p *Parser) appendSourcePrefix(b []byte, pfx Prefix) []byte {
	return p.appendValueHeader(b, SubTypeSourcePrefix, func(b []byte) []byte {
		o := len(b)

		b = p.appendUint8(b, 0) // Placeholder: plen
		b, _, plen, _ := p.appendPrefix(b, pfx, false)

		b[o] = plen

		return b
	})
}
