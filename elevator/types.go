//This package defines the different types of declarations that
//is used by the elevator module to communicate with the driver

package elevator

import (
	"../driver"
)

type FloorOrders map[int]bool
type Orders map[driver.BtnType]FloorOrders

type ReadDirection struct {
	Floor   int
	Dir     Direction
	Request request
	Resp    chan bool
}

type ReadOrder struct {
	Order driver.ClickEvent
	Resp  chan bool
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

//function handle motor direction
func (direction Direction) OppositeDirection() Direction {
	if direction == Up {
		return Down
	} else {
		return Up
	}
}

//function to handle button action for motor direction
func (direction Direction) toBtnType() driver.BtnType {
	if direction == Up {
		return driver.Up
	} else {
		return driver.Down
	}
}
