// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package babel

type NeighbourHandler interface {
	NeighbourAdded(*Neighbour)
	NeighbourRemoved(*Neighbour)
}

type InterfaceHandler interface {
	InterfaceAdded(*Interface)
	InterfaceRemoved(*Interface)
}
