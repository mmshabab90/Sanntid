package main

import (
	"./driver/elevio"
	"./elevator"
	"log"
	"os"
	"os/signal"
)

func main()  {

	numFloors := 4

	elevio.Init("localhost:15657", numFloors)

	//Driver channels
	buttonEvent := make(chan elevio.ButtonEvent)
	sensorEvent := make(chan int)
	stopButtonEvent := make(chan bool)

	//Elevator event channels
	completedFloor := make(chan int)
	floorReached := make(chan int)
	newDirection := make(chan elevator.Direction)
	doorClosed := make(chan bool)
	startedMoving := make(chan bool)
	passingFloor := make(chan bool)

	//Elevator error channels
	elevatorError := make(chan bool)
	resumeAfterError := make(chan bool)
	externalError := make(chan bool)

	//Elevator order channels
	readDirection := make(chan elevator.ReadDirection)
	readOrder := make(chan elevator.ReadOrder)

	elevio.pollInit(buttonEvent, sensorEvent, stopButtonEvent)

	go elevator.Run (
		completedFloor,
		floorReached,
		newDirection,
		doorClosed,
		startedMoving,
		passingFloor,
		elevatorError,
		resumeAfterError,
		externalError,
		readDirection,
		readOrder)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	go func() {
		<- c
		elevio.SetMotorDirection(elevio.MD_Stop)
		log.Fatal("[FATAL]\tUser terminated program")
	}()
}
