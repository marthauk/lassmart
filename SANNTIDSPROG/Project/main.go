package main

import (
	. "./elevController"
	//"fmt"
	"time"
)

func main() {
	/* INITIALIZATION */
	FSM_setup_elevator()

	/* SETS INITIAL STATE VARIABLES */
	e_system := Initialize_elev_system()
	Orders_init(&e_system)

	/* CHANNELS FOR UPDATING THE ELEVATOR VARIABLES */
	Button_Press_Chan := make(chan Button, 10)
	Location_Chan := make(chan int, 1)
	Motor_Direction_Chan := make(chan int, 1)
	Destination_Chan := make(chan int, 1)
	State_Chan := make(chan int, 1)

	/* EVENT CHANNELS */
	Objective_Chan := make(chan Button, 1)
	Floor_Arrival_Chan := make(chan int, 1)
	Door_Open_Req_Chan := make(chan int, 1)
	Door_Close_Req_Chan := make(chan int, 1)

	/* MESSAGE CHANNELS */
	Rchv_Message_Chan := make(chan Message, 10)
	Broadcast_Message_Chan := make(chan Message, 1)
	New_Order_Handler_Flag_Chan := make(chan bool, 1)
	Orders_Deleted_Chan := make(chan Button, 3)

	/* MASTER CONTROLLED EVENTS */
	Ping_Slaves_Chan := make(chan string, 1)
	Time_Window_Timeout_Chan := make(chan int, 1)
	From_Master_NewUpdate_Chan := make(chan Message, 10)
	From_Master_ReqSys_Chan := make(chan int, 1)

	/* STARTS ESSENTIAL PROCESSES */
	go New_Order_handler(Button_Press_Chan, &e_system)
	go Set_Handler_Flag(New_Order_Handler_Flag_Chan)
	go FSM_light_controller(e_system) //controlls all the button lights for all of the floors
	go FSM_safekill()
	go FSM_sensor_pooler(Button_Press_Chan)
	go FSM_floor_tracker(&e_system, Location_Chan, Floor_Arrival_Chan)
	go FSM_objective_dealer(&e_system, State_Chan, Destination_Chan, Objective_Chan)
	go FSM_elevator_updater(&e_system, Motor_Direction_Chan, Location_Chan, Destination_Chan, State_Chan)

	/* STARTS THE NETWORK BETWEEN THE ELEVATORS AND THE MESSAGE-PASSING */
	go UDPListenForPing(PINGPORT, e_system, From_Master_ReqSys_Chan)               // used PORT earlier
	go UDPListenForUpdateSlave(SUPDATEPORT, &e_system, From_Master_NewUpdate_Chan) // used PORT earlier
	if Is_elev_master(e_system) {
		go UDPListenForUpdateMaster(MUPDATEPORT, Rchv_Message_Chan) // used PORT earlier
		Ping_Slaves_Chan <- "1"                                     //Initiates the master events
	}

	// Channels to see if slaves are alive

	time.Sleep(time.Millisecond * 200)

	Print_all_orders(e_system)

	for {
		//fmt.Println("Current floor: ", e_system.Elevators[e_system.SelfID].CurrentFloor)
		select {
		/* FSM EVENTS: */
		case newObjective := <-Objective_Chan:
			FSM_Start_Driving(newObjective, &e_system, State_Chan, Motor_Direction_Chan, Location_Chan)

		case newFloorArrival := <-Floor_Arrival_Chan:
			FSM_should_stop_or_not(newFloorArrival, &e_system, State_Chan, Motor_Direction_Chan, Door_Open_Req_Chan)

		case doorOpenReq := <-Door_Open_Req_Chan:
			go FSM_door_opener(doorOpenReq, Door_Close_Req_Chan, State_Chan)

		case doorCloseReq := <-Door_Close_Req_Chan:
			FSM_door_closer(doorCloseReq, &e_system, State_Chan, Orders_Deleted_Chan)

		/* NETWORK EVENTS: */
		/* MASTER ONLY EVENTS */
		case sendReq := <-Ping_Slaves_Chan:
			UDPSendReqToSlaves(PINGPORT, sendReq)           //Ping slaves for them to send their system info. Used PINGPORT earlier
			go Int_Timer_Chan(Time_Window_Timeout_Chan, 35) //Opens a time window

		case infoRec := <-Rchv_Message_Chan:
			Message_Compiler_Master(infoRec, &e_system) //Gathering system info meanwhile

		case <-Time_Window_Timeout_Chan: //Time window closes, starts processing info
			NetworkOrderHandler(&e_system)                //Start processing information gathered
			UDPSendSysInfoToSlaves(SUPDATEPORT, e_system) //Sends it out
			go String_Timer_Chan(Ping_Slaves_Chan, 100)   //Waits before a new time window opens

		/* SLAVE EVENTS */
		case <-From_Master_ReqSys_Chan:
			MessageSetter(Broadcast_Message_Chan, e_system)
			UDPSendToMaster(MUPDATEPORT, Broadcast_Message_Chan, New_Order_Handler_Flag_Chan)
			NewOrders_reset(&e_system)

		case newSysInfo := <-From_Master_NewUpdate_Chan:
			Sync_with_system(newSysInfo, &e_system, New_Order_Handler_Flag_Chan, Orders_Deleted_Chan)

			/*
				case aliveReq := <- Alive_Ping_Chan:
					UDPSendAliveMessage(blabla)
			*/

			/*
				case someoneDied := <- Death_Occured_Chan:
					function that removes the elev that died from the system
					and sets a new master if neccessary
			*/
		}
	}
}
