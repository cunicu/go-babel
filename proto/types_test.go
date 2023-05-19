// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package proto_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stv0g/go-babel/proto"
)

var _ = Context("Types", func() {
	DescribeTable("SeqnoDistance",
		func(a, b, d int) {
			Expect(proto.SeqnoDistance(uint16(a), uint16(b))).To(Equal(int16(d)))
			Expect(proto.SeqnoDistance(uint16(b), uint16(a))).To(Equal(int16(-d)))
		},
		Entry("case 1", 0x0001, 0x0001, 0),
		Entry("case 2", 0x0001, 0x0002, 1),
		Entry("case 3", 0xffff, 0x0000, 1),
		Entry("case 4a", 0x0000, 0x8000, 32768),
		Entry("case 4b", 0x8000, 0x0000, 32768),
		Entry("case 5", 0x8000, 0x8001, 1),
		Entry("case 6", 0x0000, 0x7fff, 32767),
		Entry("case 7", 0x0000, 0x0001, 1),
		Entry("case 8", 0xfffe, 0x0000, 2),
	)

	DescribeTable("SeqnoLess",
		func(a, b int, l bool) {
			Expect(proto.SeqnoLess(uint16(a), uint16(b))).To(Equal(l))
		},
		Entry("case 1", 0x0001, 0x0001, false),
		Entry("case 2", 0x0001, 0x0002, true),
		Entry("case 3", 0x0002, 0x0001, false),
		Entry("case 4", 0xffff, 0x0000, true),
		Entry("case 5", 0x0000, 0xffff, false),
		Entry("case 6", 0x0000, 0x8000, false),
		Entry("case 7", 0x0000, 0x7fff, true),
		Entry("case 8", 0x0000, 0x8001, false),
	)

	// All seqnos which are equal or 1<<15 apart,
	// are neither larger or smaller in relation to each other
	DescribeTable("equal",
		func(a, b int) {
			Expect(proto.SeqnoLess(uint16(a), uint16(b))).To(BeFalse())
			Expect(proto.SeqnoLess(uint16(b), uint16(a))).To(BeFalse())
		},
		Entry("case 1", 0x0000, 0x8000),
		Entry("case 2", 0x0100, 0x8100),
		Entry("case 3", 0x0000, 0x0000),
		Entry("case 4", 0x0100, 0x0100),
	)
})
