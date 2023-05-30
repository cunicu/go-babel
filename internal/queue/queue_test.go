// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package queue_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestQueue(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Queue suite")
}

var _ = Describe("Queue", func() {
})
