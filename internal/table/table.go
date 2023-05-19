// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package table

import "sync"

type Table[K comparable, V any] struct {
	kvs map[K]V
	mu  sync.RWMutex
}

func New[K comparable, V any]() Table[K, V] {
	return Table[K, V]{
		kvs: map[K]V{},
	}
}

func (t *Table[K, V]) Lookup(k K) (V, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	v, ok := t.kvs[k]
	return v, ok
}

func (t *Table[K, V]) Insert(k K, v V) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.kvs[k] = v
}

func (t *Table[K, V]) Foreach(cb func(k K, v V) error) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	for k, v := range t.kvs {
		if err := cb(k, v); err != nil {
			return err
		}
	}

	return nil
}

func (t *Table[K, V]) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.kvs = map[K]V{}
}

func (t *Table[K, V]) Len() int {
	t.mu.Lock()
	defer t.mu.Unlock()

	return len(t.kvs)
}

func (t *Table[K, V]) Empty() bool {
	return t.Len() == 0
}

func (t *Table[K, V]) Update(m map[K]V) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for k, v := range m {
		t.kvs[k] = v
	}
}
