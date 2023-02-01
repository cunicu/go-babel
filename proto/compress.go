// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package proto

import (
	"net"
	"net/netip"
)

func AddressFrom(addr net.Addr) Address {
	switch a := addr.(type) {
	case *net.UDPAddr:
		return Address(a.AddrPort().Addr())
	}

	return Address{}
}

func addressEncoding(a *Address) AddressEncoding {
	switch {
	case a.IsUnspecified():
		return AddressEncodingWildcard
	case a.Is6() && a.IsLinkLocalUnicast():
		return AddressEncodingIPv6LinkLocal
	case a.Is4():
		return AddressEncodingIPv4
	case a.Is6():
		return AddressEncodingIPv6
	default:
		panic(ErrInvalidAddress)
	}
}

func addressFamilyFromAddressEncoding(ae AddressEncoding) AddressFamily {
	switch ae {
	case AddressEncodingIPv4:
		return AddressFamilyIPv4
	case AddressEncodingIPv6, AddressEncodingIPv6LinkLocal, AddressEncodingIPv4inIPv6:
		return AddressFamilyIPv6
	case AddressEncodingWildcard:
		fallthrough
	default:
		return AddressFamilyUnspecified
	}
}

func decompressIPv4(b []byte, o int, p netip.Addr) netip.Addr {
	c := [4]byte{}

	q := p.As4()
	for i := 0; i < o; i++ {
		c[i] = q[i]
	}

	for i := 0; i < 4; i++ {
		c[i+o] = b[i]
	}

	return netip.AddrFrom4(c)
}

func decompressIPv6(b []byte, o int, p netip.Addr) netip.Addr {
	c := [16]byte{}

	q := p.As16()
	for i := 0; i < o; i++ {
		c[i] = q[i]
	}

	for i := 0; i < 16; i++ {
		c[i+o] = b[i]
	}

	return netip.AddrFrom16(c)
}

func compressIPv4(b []byte, a netip.Addr, p netip.Addr) int {
	r := a.As4()
	s := p.As4()

	o := 0
	for r[o] == s[o] && o <= 4 {
		o++
	}

	return o
}

func compressIPv6(b []byte, a netip.Addr, p netip.Addr) int {
	r := a.As16()
	s := p.As16()

	o := 0
	for r[o] == s[o] && o <= 16 {
		o++
	}

	return o
}
