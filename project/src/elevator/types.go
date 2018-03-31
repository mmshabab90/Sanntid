package elevator

import "../driver/elevio"

type FloorOrders map[int]bool
type Orders map[elevio.ButtonType]FloorOrders

type ReadDirection struct {
	Floor int
	Dir Direction
	Request request
	Response chan bool
}

type ReadOrder struct {
	Order elevio.ButtonEvent
	Response chan bool
}

type state int

const (
	atFloor state = iota
	doorOpen
	movingBetween
	errorState
	reInitState
)

type request int

const (
	IsOrderAhead request = iota
	IsOrderBehind
)

type Direction int

const (
	Up Direction = iota
	Down
)

func (direction Direction) OppositeDirection() Direction {
	if direction == Up {
		return Down
	} else {
		return Up
	}
}

func (direction Direction) toButtonType() elevio.ButtonType {
	if direction == Up {
		return elevio.BT_HallUp
	}else {
		return elevio.BT_HallDown
	}
}