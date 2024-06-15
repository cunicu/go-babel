// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package queue

import (
	"container/list"
	"io"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"cunicu.li/go-babel/internal/deadline"
	"cunicu.li/go-babel/proto"
)

const (
	// TODO: Is this a reasonable choice?
	pacingTimeout = 10 * time.Millisecond
)

// Queue sends out TLV values over a net.PacketConn
type Queue struct {
	mtu    int
	writer io.Writer

	timer deadline.Deadline

	values *list.List // protected by mu
	mu     sync.Mutex

	stop    chan any
	stopped chan any
}

func NewQueue(mtu int, writer io.Writer) *Queue {
	q := &Queue{
		mtu:     mtu,
		writer:  writer,
		stop:    make(chan any),
		stopped: make(chan any),
		values:  list.New(),
		timer:   deadline.NewDeadline(),
	}

	go q.run()

	return q
}

func (q *Queue) Close() error {
	close(q.stop)
	<-q.stopped

	return nil
}

func (b *Queue) SendValues(vs []proto.Value, maxDelay time.Duration) {
	b.push(vs...)
	b.SendIn(maxDelay)
}

func (b *Queue) SendValue(v proto.Value, maxDelay time.Duration) {
	b.push(v)
	b.SendIn(maxDelay)
}

func (q *Queue) SendIn(maxDelay time.Duration) {
	jitter := maxDelay*3/4 + time.Duration(rand.Float64()*float64(maxDelay/2))

	q.timer.Reset(jitter)
}

func (q *Queue) run() {
	for range q.timer.C {
		if err := q.send(); err != nil {
			slog.Error("Failed to send packet", err)
		}
	}
}

func (q *Queue) send() error {
	var empty bool
	var v proto.Value

	p := proto.NewParser()

	b := make([]byte, 0, q.mtu)
	b = p.StartPacket(b)

	for {
		// Take next value from queue if it can still fit
		// into the MTU-sized buffer.
		if v, empty = q.popIf(func(v any) bool {
			return cap(b)-len(b)-p.ValueLength(v) >= 0
		}); v == nil {
			break
		}

		b = p.AppendValue(b, v)
	}

	p.FinalizePacket(b)

	// TODO: Handle partial writes?
	if _, err := q.writer.Write(b); err != nil {
		return err
	}

	// Send next packet within a short time if
	// there are still values in the queue
	if !empty {
		q.SendIn(pacingTimeout)
	}

	return nil
}

func (q *Queue) push(vs ...proto.Value) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, v := range vs {
		q.values.PushBack(v)
	}
}

func (q *Queue) popIf(test func(v any) bool) (proto.Value, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	v := q.values.Front()
	if v == nil {
		return nil, true
	}

	if !test(v.Value) {
		return nil, false
	}

	return q.values.Remove(v), q.values.Len() == 0
}
