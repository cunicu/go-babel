// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package babel

import "github.com/stv0g/go-babel/internal/table"

type InterfaceTable table.Table[int, *Interface]

func NewInterfaceTable() InterfaceTable {
	return InterfaceTable(table.New[int, *Interface]())
}

func (t *InterfaceTable) Lookup(idx int) (*Interface, bool) {
	return (*table.Table[int, *Interface])(t).Lookup(idx)
}

func (t *InterfaceTable) Insert(i *Interface) {
	(*table.Table[int, *Interface])(t).Insert(i.Index, i)
}

func (t *InterfaceTable) Foreach(cb func(int, *Interface) error) error {
	return (*table.Table[int, *Interface])(t).Foreach(cb)
}
