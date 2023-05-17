// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package babel_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stv0g/go-babel"
	g "github.com/stv0g/gont/v2/pkg"
	"golang.org/x/exp/slog"
)

type mockNeighbourHandler struct {
	neighbours chan *babel.Neighbour
}

func (h *mockNeighbourHandler) NeighbourAdded(n *babel.Neighbour) {
	h.neighbours <- n
}

func (h *mockNeighbourHandler) NeighbourRemoved(n *babel.Neighbour) {}

var _ = Context("speaker", func() {
	var err error
	var n *g.Network

	var s1, s2 *babel.Speaker

	BeforeEach(func() {
		n, err = g.NewNetwork("")
		Expect(err).To(Succeed())
	})

	AfterEach(func() {
		err = n.Close()
		Expect(err).To(Succeed())
	})

	It("works", func() {
		sw, err := n.AddSwitch("sw1")
		Expect(err).To(Succeed())

		h1, err := n.AddHost("h1",
			g.NewInterface("eth0", sw))
		Expect(err).To(Succeed())

		h2, err := n.AddHost("h2",
			g.NewInterface("eth0", sw))
		Expect(err).To(Succeed())

		handler := &mockNeighbourHandler{
			neighbours: make(chan *babel.Neighbour),
		}

		speakerConfig := &babel.SpeakerConfig{
			Multicast: true,
			Handler:   handler,
		}

		err = h1.RunFunc(func() (err error) {
			speakerConfig.Logger = slog.Default().With(slog.String("speaker", "s1"))
			s1, err = babel.NewSpeaker(speakerConfig)
			return
		})
		Expect(err).To(Succeed())

		err = h2.RunFunc(func() (err error) {
			speakerConfig.Logger = slog.Default().With(slog.String("speaker", "s2"))
			s2, err = babel.NewSpeaker(speakerConfig)
			return
		})
		Expect(err).To(Succeed())

		<-handler.neighbours
		<-handler.neighbours

		err = s1.Close()
		Expect(err).To(Succeed())

		err = s2.Close()
		Expect(err).To(Succeed())
	})
})
