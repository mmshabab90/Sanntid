package elevator

import (
	"fmt"
	"log"
	"time"

	"../driver/elevio"
)

const (
	deadLinePeriod = 3 * time.Second
	doorPeriod = 3 * time.Second
)

var	currentDirection Direction = Up

func GetCurrentDirection() Direction {
	return currentDirection
}

func Run (
	completedFloor chan<- int,
	floorReached <-chan int,
	newDirection chan<- Direction,
	doorClosed chan<- bool,
	startedMoving chan<- bool,
	passingFloor chan<- bool,
	elevatorError chan<- bool,
	resumeAfterError <-chan bool,
	externalError <-chan bool,
	readDirection chan<- ReadDirection,
	readOrders chan<- ReadOrder) {
		readResult := make(chan bool)

		deadLineTimer := time.NewTimer(deadLinePeriod)
		deadLineTimer.Stop()
		doorTimer := time.NewTimer(doorPeriod)
		doorTimer.Stop()

		state := atFloor
		lastPassedFloor := elevio.getFloor()

	if lastPassedFloor == -1 {
		log.Fatal("[FATAL]\tElevator initializing between floors")
	}

	isPassingFloor := false

	for {
		select {
		case <- externalError:
			fmt.Println("Elevator: received external error")
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevatorError <- true
			state = errorState
		default:
		}

		switch state {
		case atFloor:
			readOrders <- ReadOrder{elevio.ButtonEvent{lastPassedFloor, elevio.BT_Cab}, readResult}
			internalOrderAtThisFloor := <- readResult
			readOrders <- ReadOrder{elevio.ButtonEvent{lastPassedFloor, currentDirection.toButtonType()}, readResult}
			orderForwardAtThisFloor := <- readResult

			if internalOrderAtThisFloor || orderForwardAtThisFloor {
				isPassingFloor = false
				elevio.SetMotorDirection(elevio.MD_Stop)
				elevio.SetDoorOpenLamp(true)
				completedFloor <- lastPassedFloor
				deadLineTimer.Stop()
				doorTimer.Reset(doorPeriod)
				state = doorOpen
				break
			}

			readDirection <- ReadDirection{lastPassedFloor, currentDirection, IsOrderAhead, readResult}
			orderAhead := <- readResult

			if orderAhead {
				startedMoving <- true
				if isPassingFloor {
					passingFloor <- true
				}
				switch currentDirection {
				case Up:
					elevio.SetMotorDirection(elevio.MD_Up)
				case Down:
					elevio.SetMotorDirection(elevio.MD_Down)
				}
				deadLineTimer.Reset(deadLinePeriod)
				state = movingBetween
				isPassingFloor = true
				break
			}

			readDirection <- ReadDirection{lastPassedFloor, currentDirection, IsOrderBehind, readResult}
			orderBehind := <- readResult
			readOrders <- ReadOrder{elevio.ButtonEvent{lastPassedFloor, currentDirection.OppositeDirection().toButtonType()}, readResult}
			orderBackwardAtThisFloor := <- readResult

			if orderBehind || orderBackwardAtThisFloor {
				currentDirection = currentDirection.OppositeDirection()
				newDirection <- currentDirection
			}

		case doorOpen:
			<- doorTimer.C
			elevio.SetDoorOpenLamp(false)
			state = atFloor
			doorClosed <- true

		case movingBetween:
			select {
			case floor := <- floorReached:
				if ((currentDirection == Up) && (floor != lastPassedFloor + 1)) || ((currentDirection == Down) && (floor != lastPassedFloor - 1)) {
					fmt.Println("ELevator: missed floor signal, entering error state")
					elevio.SetMotorDirection(elevio.MD_Stop)
					elevatorError <- true
					state = errorState
					break
				}
				lastPassedFloor = floor
				elevio.SetFloorIndicator(floor)
				state = atFloor
				deadLineTimer.Stop()

			case <- deadLineTimer.C:
				fmt.Println("Elevator: timeout while moving")
				elevio.SetMotorDirection(elevio.MD_Stop)
				elevatorError <- true
				state = errorState
			}

		case reInitState:
			select {
			case floor := <- floorReached:
				elevio.SetMotorDirection(elevio.MD_Stop)
				lastPassedFloor = floor
				elevio.SetFloorIndicator(floor)
				deadLineTimer.Stop()
				state = errorState

			case <- deadLineTimer.C:
				fmt.Println("Elevator: timeout while reinitializing")
				elevio.SetMotorDirection(elevio.MD_Stop)
				elevatorError <- true
				state = errorState
			}

		case errorState:
			deadLineTimer.Stop()
			<- resumeAfterError
			if elevio.getFloor() == -1 {
				startedMoving <- true

				switch currentDirection {
				case Up:
					elevio.SetMotorDirection(elevio.MD_Up)

				case Down:
					elevio.SetMotorDirection(elevio.MD_Down)
				}

				deadLineTimer.Reset(deadLinePeriod)
				state = reInitState
			} else {
				lastPassedFloor = elevio.getFloor()
				state = atFloor
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
}