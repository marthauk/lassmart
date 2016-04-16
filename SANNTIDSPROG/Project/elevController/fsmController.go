package elevController

import (
	. "./elevDrivers"
	//"fmt"
	"log"
	"os"
	"os/signal"
	"time"
)

const (
	IDLE       = 0
	DRIVING    = 1
	DOOR_TIMER = 2
)

type Elevator struct {
	State            int
	CurrentFloor     int
	DestinationFloor int
	Direction        int
	InternalOrders   [10]Button
	NewOrders        [10]Button
}

func FSM_setup_elevator() {
	Elev_init()
	Elev_set_motor_direction(DIRN_DOWN)
	for {
		if Elev_get_floor_sensor_signal() != -1 {
			Elev_set_motor_direction(DIRN_STOP)
			break
		}
	}
	Elev_set_floor_indicator(Elev_get_floor_sensor_signal())
}

/*func FSM_create_elevator() Elevator {
	e := Elevator{State: IDLE, CurrentFloor: Elev_get_floor_sensor_signal(),
	 	DestinationFloor: Elev_get_floor_sensor_signal(), Direction: DIRN_STOP, InternalOrders: orders}
	return e
}*/

func FSM_Start_Driving(NewObjective Button, e_system *Elevator_System, State_Chan chan int, Motor_Direction_Chan chan int, Location_Chan chan int) {
	if e_system.Elevators[e_system.SelfID].CurrentFloor > NewObjective.Floor {
		Elev_set_motor_direction(-1)
		Motor_Direction_Chan <- -1
		//fmt.Printf("\nThe elevator started driving in %d direction", -1)
		State_Chan <- DRIVING
		//fmt.Println("\nState message sent")
	}
	if e_system.Elevators[e_system.SelfID].CurrentFloor < NewObjective.Floor {
		Elev_set_motor_direction(1)
		Motor_Direction_Chan <- 1
		//fmt.Printf("\nThe elevator started driving in %d direction", 1)
		State_Chan <- DRIVING
		//fmt.Println("\nState message sent")
	}
	if e_system.Elevators[e_system.SelfID].CurrentFloor == NewObjective.Floor {
		State_Chan <- DRIVING
	}
}

func FSM_objective_dealer(e_system *Elevator_System, State_Chan chan int, Destination_Chan chan int, Objective_Chan chan Button) {
	for {
		time.Sleep(time.Millisecond * 50)
		nextOrder := Next_order(*e_system)
		if e_system.Elevators[e_system.SelfID].State == IDLE && nextOrder.Floor != -1 {
			Objective_Chan <- nextOrder
			Destination_Chan <- nextOrder.Floor
			//fmt.Println("\n\nA new objective was sent to the elevator")
		}
	}
}

func FSM_elevator_updater(e_system *Elevator_System, Motor_Direction_Chan chan int, Location_Chan chan int, Destination_Chan chan int, State_Chan chan int) {
	for {
		select {
		case NewDirection := <-Motor_Direction_Chan:
			e_system.Elevators[e_system.SelfID].Direction = NewDirection
			//fmt.Println("\nNew direction: ", NewDirection)
		case NewFloor := <-Location_Chan:
			e_system.Elevators[e_system.SelfID].CurrentFloor = NewFloor
			//fmt.Println("\nNew location: ", NewFloor)
		case NewDestination := <-Destination_Chan:
			e_system.Elevators[e_system.SelfID].DestinationFloor = NewDestination
			//fmt.Println("\nNew destination: ", NewDestination)
		case NewState := <-State_Chan:
			e_system.Elevators[e_system.SelfID].State = NewState
			//fmt.Println("\nNew state: ", NewState)
		}
	}
}

func FSM_floor_tracker(e_system *Elevator_System, Location_Chan chan int, Floor_Arrival_Chan chan int) {
	for {
		time.Sleep(time.Millisecond * 200)
		if Elev_get_floor_sensor_signal() != -1 {
			NewFloor := Elev_get_floor_sensor_signal()
			Elev_set_floor_indicator(NewFloor)
			Location_Chan <- NewFloor
			Floor_Arrival_Chan <- NewFloor
		}
	}
}

func FSM_sensor_pooler(Button_Press_Chan chan Button) {
	for {
		for button := B_UP; button <= B_COMMAND; button++ {
			for floor := 0; floor < N_FLOORS; floor++ {
				if button == B_UP && floor == N_FLOORS-1 {
					continue
				}
				if button == B_DOWN && floor == 0 {
					continue
				}
				button_signal := Elev_get_button_signal(button, floor)
				if button_signal == 1 {
					Button_Press_Chan <- Button{Button_type: button, Floor: floor}
				}
			}
		}
		time.Sleep(time.Millisecond * 40)
	}
}

func FSM_light_controller(e_system Elevator_System) {
	for { //forever
		for elev_ID, _ := range e_system.Elevators { //for every elevator
			for button_type := B_UP; button_type <= B_COMMAND; button_type++ { //for every buttontype .EDIT:Martin; Du hadde skreve button_type:=B_UP;button<=B_COMMAND;button++, endrer alt til button_type istadenfor button
				for floor := 0; floor < N_FLOORS; floor++ { //for every floor
					time.Sleep(time.Millisecond * 50)
					order_exists := false
					for index := 0; index < ROWS; index++ { //for the entire queue
						if button_type == B_UP && floor == N_FLOORS-1 {
							continue
						} //this button doesnt exist
						if button_type == B_DOWN && floor == 0 {
							continue
						} //this button doesnt exist
						if e_system.Elevators[elev_ID].InternalOrders[index].Floor == floor && e_system.Elevators[elev_ID].InternalOrders[index].Button_type == button_type {
							if elev_ID != e_system.SelfID && button_type == B_COMMAND {
								continue
							} //buttons inside of other elevators should not have its light turned on
							order_exists = true
						}
					}
					if order_exists == true {
						Elev_set_button_lamp(button_type, floor, 1)
					} else {
						Elev_set_button_lamp(button_type, floor, 0)
					}
				}
			}
		}
	}
}

func FSM_should_stop_or_not(newFloorArrival int, e_system *Elevator_System, State_Chan chan int, Motor_Direction_Chan chan int, Door_Open_Req_Chan chan int) {
	if newFloorArrival == e_system.Elevators[e_system.SelfID].DestinationFloor && e_system.Elevators[e_system.SelfID].State == DRIVING {
		Elev_set_motor_direction(0)
		Motor_Direction_Chan <- 0
		Door_Open_Req_Chan <- 1
		State_Chan <- DOOR_TIMER
	}
}

func FSM_door_opener(doorOpenReq int, Door_Close_Req_Chan chan int, State_Chan chan int) {
	Elev_set_door_open_lamp(1)
	State_Chan <- 2
	time.Sleep(time.Second * 3)
	Door_Close_Req_Chan <- 1
}

func FSM_door_closer(doorCloseReq int, e_system *Elevator_System, State_Chan chan int, Orders_Deleted_Chan chan Button) {
	Elev_set_door_open_lamp(0)
	Remove_order(e_system.Elevators[e_system.SelfID].CurrentFloor, e_system, Orders_Deleted_Chan)
	State_Chan <- 0
}

func FSM_safekill() {
	var c = make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	<-c
	Elev_set_motor_direction(0)
	log.Printf("\nUser terminated program")
	os.Exit(1)
}
