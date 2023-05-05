// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package babel

import (
	"time"
)

type Deadline struct {
	C chan any

	Expired bool
	timer   *time.Timer
}

func NewDeadline() Deadline {
	return Deadline{
		C: make(chan any),
	}
}

func (t *Deadline) Close() error {
	close(t.C)
	return nil
}

func (t *Deadline) Reset(d time.Duration) {
	t.Expired = false

	t.Stop()
	t.timer = time.AfterFunc(d, func() {
		t.Expired = true
		t.C <- nil
	})
}

func (t *Deadline) Stop() {
	if t.timer != nil && !t.timer.Stop() {
		<-t.timer.C
	}
}
