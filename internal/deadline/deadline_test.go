// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package deadline_test

import (
	"testing"
	"time"

	"cunicu.li/go-babel/internal/deadline"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMain(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Deadline suite")
}

var _ = Describe("Deadline", func() {
	var d deadline.Deadline

	BeforeEach(func() {
		d = deadline.NewDeadline()
	})

	AfterEach(func() {
		err := d.Close()
		Expect(err).To(Succeed())
	})

	It("channel is not nil", func() {
		Expect(d.C).NotTo(BeNil())
	})

	It("can be stopped when not armed", func() {
		d.Stop()
		Expect(d.Expired()).To(BeFalse())
	})

	It("can be stopped when armed", func() {
		d.Reset(10 * time.Millisecond)
		d.Stop()
		Consistently(d.C).ShouldNot(Receive())
		Expect(d.Expired()).To(BeFalse())
	})

	It("does not expire when not set", func() {
		Consistently(d.C).ShouldNot(Receive())
		Expect(d.Expired()).To(BeFalse())
	})

	It("should expire when set", func() {
		d.Reset(10 * time.Millisecond)
		Eventually(d.C).Should(Receive())
		Expect(d.Expired()).To(BeTrue())
	})

	It("can be re-armed", func() {
		d.Reset(10 * time.Millisecond)
		Eventually(d.C).Should(Receive())
		Expect(d.Expired()).To(BeTrue())

		d.Reset(10 * time.Millisecond)
		Expect(d.Expired()).To(BeFalse())

		Eventually(d.C).Should(Receive())
		Expect(d.Expired()).To(BeTrue())
	})

	It("can be reset while armed", func() {
		started := time.Now()
		d.Reset(10 * time.Millisecond)
		d.Reset(100 * time.Millisecond)
		Eventually(d.C).Should(Receive())
		Expect(time.Since(started)).To(BeNumerically(">", 100*time.Millisecond))
	})

	It("can be reset twice while armed", func() {
		started := time.Now()
		d.Reset(10 * time.Millisecond)
		d.Reset(100 * time.Millisecond)
		d.Reset(10 * time.Millisecond)
		Eventually(d.C).Should(Receive())
		Expect(time.Since(started)).To(And(
			BeNumerically(">", 10*time.Millisecond),
			BeNumerically("<", 50*time.Millisecond),
		))
	})
})
