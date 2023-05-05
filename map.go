// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package babel

import "sync"

type Map[K comparable, V any] struct {
	kvs map[K]V
	mu  sync.RWMutex
}

func NewMap[K comparable, V any]() Map[K, V] {
	return Map[K, V]{
		kvs: map[K]V{},
	}
}

func (m *Map[K, V]) Lookup(k K) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	v, ok := m.kvs[k]
	return v, ok
}

func (m *Map[K, V]) Insert(k K, v V) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.kvs[k] = v
}

func (m *Map[K, V]) Foreach(cb func(k K, v V) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for k, v := range m.kvs {
		if err := cb(k, v); err != nil {
			return err
		}
	}

	return nil
}
