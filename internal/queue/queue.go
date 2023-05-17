// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package queue

import (
	"encoding/binary"
	"io"
	"math/rand"
	"time"

	"github.com/stv0g/go-babel/internal/deadline"
	"github.com/stv0g/go-babel/proto"
	"golang.org/x/exp/slog"
)

// Queue sends out TLV values over a net.PacketConn
type Queue struct {
	mtu    int
	buffer []byte
	parser proto.Parser
	writer io.Writer

	timer deadline.Deadline

	values  chan proto.Value
	stop    chan any
	stopped chan any
}

func NewQueue(mtu int, writer io.Writer) *Queue {
	q := &Queue{
		mtu:     mtu,
		buffer:  make([]byte, 0, mtu),
		writer:  writer,
		stop:    make(chan any),
		stopped: make(chan any),
		values:  make(chan proto.Value),
		timer:   deadline.NewDeadline(),
	}

	q.resetBuffer()

	go q.run()

	return q
}

func (q *Queue) Close() error {
	close(q.stop)
	<-q.stopped

	return nil
}

func (b *Queue) SendValues(vs []proto.Value, maxDelay time.Duration) {
	for _, v := range vs {
		b.values <- v
	}

	b.FlushIn(maxDelay)
}

func (b *Queue) SendValue(v proto.Value, maxDelay time.Duration) {
	b.values <- v

	b.FlushIn(maxDelay)
}

func (q *Queue) FlushIn(maxDelay time.Duration) {
	jitter := maxDelay*3/4 + time.Duration(rand.Float64()*float64(maxDelay/2))

	q.timer.Reset(jitter)
}

func (q *Queue) run() {
	for {
		select {
		case v := <-q.values:
			vlen := int(q.parser.ValueLength(v))

			if cap(q.buffer)-len(q.buffer)-vlen < 0 {
				if err := q.flushBuffer(); err != nil {
					panic(err) // TODO: Handle logging
				}
			}

			q.buffer = q.parser.AppendValue(q.buffer, v)

		case <-q.timer.C:
			if err := q.flushBuffer(); err != nil {
				panic(err) // TODO: Handle logging
			}
		}
	}
}

func (q *Queue) flushBuffer() error {
	// Fill in body length
	bodyLength := len(q.buffer) - proto.PacketHeaderLength
	binary.BigEndian.PutUint16(q.buffer[2:], uint16(bodyLength))

	// TODO: Handle partial writes
	if _, err := q.writer.Write(q.buffer); err != nil {
		return err
	}

	slog.Debug("Flushing queue")

	q.resetBuffer()

	return nil
}

func (q *Queue) resetBuffer() {
	q.buffer = q.parser.AppendPacket(q.buffer[:0], &proto.Packet{})
}
