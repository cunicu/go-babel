// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package proto

import (
	"net/netip"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("Parser", func() {
	var p *Parser
	BeforeEach(func() {
		p = NewParser()
	})

	Describe("Types", func() {
		It("uint8", func() {
			var b []byte
			var i int
			var err error
			var n uint8

			vs := []uint8{0xaa, 0xbb, 0xcc}
			for i, v := range vs {
				b = p.appendUint8(b, v)
				Expect(b).To(HaveLen((i + 1) * 1))
			}

			for i = 0; len(b) > 0; i++ {
				b, n, err = p.uint8(b)
				Expect(err).To(Succeed())
				Expect(n).To(Equal(vs[i]))
			}

			Expect(b).To(BeEmpty())
		})

		It("uint16", func() {
			var b []byte
			var i int
			var err error
			var n uint16

			vs := []uint16{0xaaaa, 0xbbbb, 0xcccc}
			for i, v := range vs {
				b = p.appendUint16(b, v)
				Expect(b).To(HaveLen((i + 1) * 2))
			}

			for i = 0; len(b) > 0; i++ {
				b, n, err = p.uint16(b)
				Expect(err).To(Succeed())
				Expect(n).To(Equal(vs[i]))
			}

			Expect(b).To(BeEmpty())
		})

		It("uint64", func() {
			var b []byte
			var i int
			var err error
			var n uint64

			vs := []uint64{0xaaaaaaaa, 0xbbbbbbbb, 0xcccccccc}
			for i, v := range vs {
				b = p.appendUint64(b, v)
				Expect(b).To(HaveLen((i + 1) * 8))
			}

			for i = 0; len(b) > 0; i++ {
				b, n, err = p.uint64(b)
				Expect(err).To(Succeed())
				Expect(n).To(Equal(vs[i]))
			}

			Expect(b).To(BeEmpty())
		})

		It("Interval", func() {
			v1 := 12 * time.Second

			b := p.appendInterval(nil, v1)
			Expect(b).To(HaveLen(2))

			b, v2, err := p.interval(b)
			Expect(err).To(Succeed())
			Expect(v2).To(Equal(v1))
			Expect(b).To(BeEmpty())
		})

		It("RouterID", func() {
			v1 := RouterID{0x01, 0x23, 0x34, 0x45, 0x67, 0x89, 0x0a, 0xbc}

			b := p.appendRouterID(nil, v1)
			Expect(b).To(HaveLen(8))

			b, v2, err := p.routerID(b)
			Expect(err).To(Succeed())
			Expect(v2).To(Equal(v1))
			Expect(b).To(BeEmpty())
		})

		DescribeTable("RouterID invalids",
			func(rid RouterID) {
				_, _, err := p.routerID(rid[:])
				Expect(err).To(MatchError(ErrInvalidRouterID))
			},
			Entry("AllOnes", RouterIDAllOnes),
			Entry("AllZeros", RouterIDAllZeros),
		)

		DescribeTable("RouterIDFromAddr",
			func(addrStr string, rid RouterID) {
				addr := netip.MustParseAddr(addrStr)
				Expect(RouterIDFromAddr(addr)).To(Equal(rid))
			},
			Entry("IPv4", "10.168.44.55", RouterID{0, 0, 0, 0, 10, 168, 44, 55}),
			Entry("IPv6", "2a09:bac0:35::826:93f9", RouterID{0, 0, 0, 0, 0x08, 0x26, 0x93, 0xf9}),
			Entry("IPv6 link-local", "fe80::210:5aff:feaa:20a2", RouterID{0x02, 0x10, 0x5a, 0xff, 0xfe, 0xaa, 0x20, 0xa2}),
			Entry("IPv4in6 mapped", "::ffff:1.2.3.4", RouterID{0, 0, 0, 0, 1, 2, 3, 4}),
		)

		DescribeTable("Address",
			func(addr string, len int, expAE uint8) {
				v1 := netip.MustParseAddr(addr)

				b, ae := p.appendAddress(nil, v1, -1)
				Expect(b).To(HaveLen(len))
				Expect(ae).To(Equal(expAE))

				b, v2, err := p.address(b, ae, 0, -1)
				Expect(err).To(Succeed())
				Expect(v2).To(Equal(v1), "Addresses are not equal: %v != %v", v1, v2)
				Expect(b).To(BeEmpty())

				if ae == AddressEncodingIPv4inIPv6 {
					Expect(v2.Is4In6()).To(BeTrue())
					Expect(v2.Is6()).To(BeTrue())
				}
			},
			Entry("AddressEncodingIPv4", "1.1.1.1", 4, AddressEncodingIPv4),
			Entry("AddressEncodingIPv6", "fd3d:bd4f:9738::1036:d55b:fb01:b6d1", 16, AddressEncodingIPv6),
			Entry("AddressEncodingWildcard", "::", 0, AddressEncodingWildcard),
			Entry("AddressEncodingIPv6LinkLocal", "fe80::1234:5678:90AB:CDEF", 8, AddressEncodingIPv6LinkLocal),
			Entry("AddressEncodingIPv4inIPv6", "::ffff:1.2.3.4", 4, AddressEncodingIPv4inIPv6),
		)

		DescribeTable("Prefix",
			func(addr string, len int, expAE, expPlen uint8) {
				v1 := netip.MustParsePrefix(addr)

				b, ae, plen, _ := p.appendPrefix(nil, v1, false)
				Expect(b).To(HaveLen(len))
				Expect(ae).To(Equal(expAE))
				Expect(plen).To(Equal(expPlen))

				b, v2, err := p.prefix(b, ae, plen, 0)
				Expect(err).To(Succeed())
				Expect(v2).To(Equal(v1))
				Expect(b).To(BeEmpty())

				if ae == AddressEncodingIPv4inIPv6 {
					Expect(v2.Addr().Is4In6()).To(BeTrue())
					Expect(v2.Addr().Is6()).To(BeTrue())
				}
			},
			Entry("AddressEncodingIPv4", "1.1.0.0/16", 2, AddressEncodingIPv4, uint8(16)),
			Entry("AddressEncodingIPv6", "fd3d:bd4f:9738::/48", 6, AddressEncodingIPv6, uint8(48)),
			Entry("AddressEncodingWildcard", "::/0", 0, AddressEncodingWildcard, uint8(0)),
			Entry("AddressEncodingIPv6LinkLocal", "fe80::1234:5678:90AB:CDEF/128", 8, AddressEncodingIPv6LinkLocal, uint8(128)),
			Entry("AddressEncodingIPv4inIPv6", "::ffff:10.0.0.0/16", 2, AddressEncodingIPv4inIPv6, uint8(16)),
		)

		It("Prefixes compression", Pending, func() {
			// TODO
		})
	})

	pfx := netip.MustParsePrefix("10.0.0.0/24")

	Describe("Values", func() {
		Describe("Flags", func() {
			DescribeTable("FlagUpdateRouterID",
				func(pfxStr string) {
					pfx := netip.MustParsePrefix(pfxStr)
					rid := RouterIDFromAddr(pfx.Addr())

					b := p.AppendValue(nil, &Update{
						Flags:  FlagUpdateRouterID,
						Prefix: pfx,
					})

					// Check that an update without the flag does not overwrite the current router ID
					b = p.AppendValue(b, &Update{
						Prefix: netip.MustParsePrefix("10.169.0.0/16"),
					})

					b = p.AppendValue(b, &Update{
						Prefix: netip.MustParsePrefix("2a10:bac0:35::826:93f9/128"),
					})

					b = p.AppendValue(b, &Update{
						Prefix: netip.MustParsePrefix("fe80::310:5aff:feaa:20a2/64"),
					})

					b = p.AppendValue(b, &Update{
						Prefix: netip.MustParsePrefix("::ffff:1.2.3.5/128"),
					})

					Expect(p.CurrentRouterID).To(Equal(rid))

					p.Reset()

					_, _, err := p.Values(b, false)
					Expect(err).To(Succeed())

					Expect(p.CurrentRouterID).To(Equal(rid))
				},
				Entry("IPv4", "10.168.0.0/16"),
				Entry("IPv6", "2a09:bac0:35::826:93f9/128"),
				Entry("IPv6 link-local", "fe80::210:5aff:feaa:20a2/64"),
				Entry("IPv4in6 mapped", "::ffff:1.2.3.4/128"),
			)

			It("FlagUpdatePrefix", func() {
				pfx4 := netip.MustParsePrefix("10.168.0.0/16")
				pfx6 := netip.MustParsePrefix("fd5e:181e:5bbd::/48")
				pfx6LL := netip.MustParsePrefix("fe80::1234:5678:90ab:cdef/128")
				pfx4in6 := netip.MustParsePrefix("::ffff:1.2.3.4/128")

				expectedDefaultPrefix := map[AddressEncoding]Address{
					AddressEncodingIPv4:          pfx4.Addr(),
					AddressEncodingIPv6:          pfx6.Addr(),
					AddressEncodingIPv6LinkLocal: pfx6LL.Addr(),
					AddressEncodingIPv4inIPv6:    pfx4in6.Addr(),
				}

				// IPv4
				b := p.AppendValue(nil, &Update{
					Flags:  FlagUpdatePrefix,
					Prefix: pfx4,
				})

				// IPv6
				b = p.AppendValue(b, &Update{
					Flags:  FlagUpdatePrefix,
					Prefix: pfx6,
				})

				// IPv6 link-local
				b = p.AppendValue(b, &Update{
					Flags:  FlagUpdatePrefix,
					Prefix: pfx6LL,
				})

				// IPv4in6 mapped
				b = p.AppendValue(b, &Update{
					Flags:  FlagUpdatePrefix,
					Prefix: pfx4in6,
				})

				// Check that an update without the flag does not overwrite the current router ID
				b = p.AppendValue(b, &Update{
					Prefix: netip.MustParsePrefix("10.169.0.0/16"),
				})

				b = p.AppendValue(b, &Update{
					Prefix: netip.MustParsePrefix("fe80::1337/128"),
				})

				Expect(p.CurrentDefaultPrefix).To(Equal(expectedDefaultPrefix))

				p.Reset()

				_, _, err := p.Values(b, false)
				Expect(err).To(Succeed())

				Expect(p.CurrentDefaultPrefix).To(Equal(expectedDefaultPrefix))
			})
		})

		DescribeTable("Values",
			func(typ1 ValueType, v1 Value) {
				b := p.AppendValue(nil, v1)
				Expect(b).To(HaveLen(p.ValueLength(v1)))

				p.Reset()

				_, v2, typ2, err := p.value(b)
				Expect(err).To(Succeed())
				Expect(typ2).To(Equal(typ1))
				Expect(v2).To(Equal(v1))
			},
			Entry("Pad1", TypePad1, &Pad1{}),
			Entry("PadN", TypePadN, &PadN{
				N: 111,
			}),
			Entry("AcknowledgmentRequest", TypeAcknowledgmentRequest, &AcknowledgmentRequest{
				Opaque:   0x1234,
				Interval: 4 * time.Second,
			}),
			Entry("Acknowledgment", TypeAcknowledgment, &Acknowledgment{
				Opaque: 0x1234,
			}),
			Entry("Hello", TypeHello, &Hello{
				Flags:    FlagHelloUnicast,
				Seqno:    1233,
				Interval: 33 * time.Second,
			}),
			Entry("Hello with Timestamp", TypeHello, &Hello{
				Flags:    FlagHelloUnicast,
				Seqno:    1233,
				Interval: 33 * time.Second,
				Timestamp: &TimestampHello{
					Transmit: 532235,
				},
			}),
			Entry("IHU", TypeIHU, &IHU{
				RxCost:   0xABCD,
				Interval: 2 * time.Second,
				Address:  netip.MustParseAddr("1.2.3.4"),
			}),
			Entry("IHU with Timestamp", TypeIHU, &IHU{
				RxCost:   0xABCD,
				Interval: 2 * time.Second,
				Address:  netip.MustParseAddr("1.2.3.4"),
				Timestamp: &TimestampIHU{
					Origin:  42394723,
					Receive: 23283423,
				},
			}),
			Entry("RouterID", TypeRouterID, &RouterIDValue{
				RouterID: RouterID{0x01, 0x23, 0x34, 0x45, 0x67, 0x89, 0x0a, 0xbc},
			}),
			Entry("NextHop", TypeNextHop, &NextHop{
				NextHop: netip.MustParseAddr("1.2.3.4"),
			}),
			Entry("Update", TypeUpdate, &Update{
				Flags:    FlagUpdatePrefix,
				Interval: 2 * time.Second,
				Seqno:    1233,
				Metric:   100,
				Prefix:   netip.MustParsePrefix("192.168.0.0/16"),
			}),
			Entry("Update with SourcePrefix", TypeUpdate, &Update{
				Flags:        FlagUpdatePrefix,
				Interval:     2 * time.Second,
				Seqno:        1233,
				Metric:       100,
				Prefix:       netip.MustParsePrefix("192.168.0.0/16"),
				SourcePrefix: &pfx,
			}),
			Entry("RouteRequest", TypeRouteRequest, &RouteRequest{
				Prefix: netip.MustParsePrefix("192.168.0.0/16"),
			}),
			Entry("RouteRequest with SourcePrefix", TypeRouteRequest, &RouteRequest{
				Prefix:       netip.MustParsePrefix("192.168.0.0/16"),
				SourcePrefix: &pfx,
			}),
			Entry("SeqnoRequest", TypeSeqnoRequest, &SeqnoRequest{
				Seqno:    1233,
				HopCount: 99,
				RouterID: RouterID{0x01, 0x23, 0x34, 0x45, 0x67, 0x89, 0x0a, 0xbc},
				Prefix:   netip.MustParsePrefix("192.168.0.0/16"),
			}),
			Entry("SeqnoRequest with SourcePrefix", TypeSeqnoRequest, &SeqnoRequest{
				Seqno:        1233,
				HopCount:     99,
				RouterID:     RouterID{0x01, 0x23, 0x34, 0x45, 0x67, 0x89, 0x0a, 0xbc},
				Prefix:       netip.MustParsePrefix("192.168.0.0/16"),
				SourcePrefix: &pfx,
			}),
		)
	})

	Describe("Sub-TLVs", func() {
		It("Ignore unsupported sub-TLVs", Pending, func() {
			// TODO
		})

		It("Fail on mandatory sub-TLVs", Pending, func() {
			// TODO
		})
	})

	It("Ignore unsupported TLVs", Pending, func() {
		// TODO
	})
})
