// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package history_test

import (
	"testing"

	"cunicu.li/go-babel/internal/history"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMain(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "History suite")
}

var _ = Describe("Vector", func() {
	var v history.HelloHistory

	BeforeEach(func() {
		v.Reset()
	})

	const Missed = 0xffff

	DescribeTable("simple",
		func(k, j int, result bool, seqnos ...int) {
			for _, seqno := range seqnos {
				if seqno == Missed {
					v.Missed()
				} else {
					v.Update(uint16(seqno))
				}
			}

			Expect(v.OutOf(k, j)).To(Equal(result))
		},
		Entry("empty history and 0-out-of-3 is okay", 0, 3, true),
		Entry("empty history and 1-out-of-3 must fail", 1, 3, false),
		Entry("1-out-of-3 with 1 entry", 1, 3, true, 1),
		Entry("2-out-of-3 but only 1 entry", 2, 3, false, 1),
		Entry("1-out-of-3 with reset is okay", 1, 3, true, 100),
		Entry("1-out-of-3 with missed in between is okay", 1, 3, true, 1, Missed, 2),
		Entry("2-out-of-3 with missed in between is okay", 2, 3, true, 1, Missed, 2),
		Entry("3-out-of-3 with missed in between must fail", 3, 3, false, 1, Missed, 2),
		Entry("2-out-of-3 with repetition", 2, 3, true, 1, 2, 2),
		Entry("2-out-of-3 with to many repititions", 2, 3, false, 1, 1, 1),
		Entry("2-out-of-3 with single repeated seqno must fail", 2, 2, false, 1, 1, 1),
		Entry("2-out-of-3 with reset", 2, 3, true, 1, 2, 3, 4, 100, 101),
		Entry("2-out-of-3 with reset but only 1 valid", 2, 3, false, 1, 2, 3, 4, 100),
		Entry("2-out-of-3 with skip", 2, 3, false, 1, 2, 3, 6),
		Entry("2-out-of-3 with less skip", 2, 3, true, 1, 2, 3, 5),
		Entry("2-out-of-3 with undo", 2, 3, true, 1, 2, 3, 4, 5, 6, 3),
		Entry("2-out-of-3 with more undo", 2, 3, false, 1, 2, 3, 1),
		Entry("2-out-of-3: missed and recovered", 2, 3, true, 100, 101, 102, Missed, Missed, Missed, 106, 107),
		Entry("case 15: missed with rewind", 2, 3, true, 100, 101, 102, Missed, Missed, Missed, 102, 103),
		Entry("case 16: missed and dead", 2, 3, false, 1, 2, 3, 4, 5, Missed, Missed),
	)

	It("Empty", func() {
		Expect(v.Empty()).To(BeTrue())
	})

	It("detects reset", func() {
		resetted := v.Update(100)
		Expect(resetted).To(BeTrue())

		resetted = v.Update(200)
		Expect(resetted).To(BeTrue())

		resetted = v.Update(201)
		Expect(resetted).To(BeFalse())
	})
})
