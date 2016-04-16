package elevController

import (
	//. "./elevDrivers"
	"fmt"
	//"time"
)

/*
	The matrix(10x2) for the orders of a elevator are on the form

			2

	[FLOOR][BUTTON_TYPE]
	[FLOOR][BUTTON_TYPE]
	[FLOOR][BUTTON_TYPE]
	.						}10
	.
	.
	[FLOOR][BUTTON_TYPE]

	Which is a priority list starting at the top. COMMAND button types have higher priorities than
	other button types and are automatically moved infront.

	There is four floors: 0, 1, 2, and 3.
	There is three button types: up, down and command.
*/

const (
	ROWS = 10
)

const (
	b_UP      = 0
	b_DOWN    = 1
	b_COMMAND = 2
)

type Button struct {
	Button_type int
	Floor       int
}

var New_Order_Handler_Flag bool = true

func Next_order(e_system Elevator_System) Button {
	return e_system.Elevators[e_system.SelfID].InternalOrders[0]
}

func Set_Handler_Flag(New_Order_Handler_Flag_Chan chan bool) {
	for {
		New_Order_Handler_Flag = <-New_Order_Handler_Flag_Chan
		//fmt.Println("New_Order_Handler_Fag: ", New_Order_Handler_Flag)
	}
}

func Add_new_order(button Button, e_system *Elevator_System) {
	myID := e_system.SelfID
	order_exists := Check_if_new_order_exists(button, *e_system, myID) //VEEEEEEEEEEEELDIG USIKKER PÅ OM DETTE FUNKER, MULIG ME MÅ TA INN EIN E_SYSTEM KOPI I TILLEGG
	if order_exists == 0 {
		for i := 0; i < ROWS; i++ {
			if e_system.Elevators[e_system.SelfID].NewOrders[i].Floor == -1 {
				e_system.Elevators[e_system.SelfID].NewOrders[i] = button
				//Elev_set_button_lamp(button.Button_type, button.Floor, 1)
				/*
					if orders[i].Button_type == b_COMMAND{
						move_order_infront(i)
					}
				*/
				return
			}
		}
	}
	Print_all_orders(*e_system)
}

func Check_if_new_order_exists(button Button, e_system Elevator_System, ID string) int {
	exists := 0
	for i := 0; i < ROWS; i++ {
		if e_system.Elevators[ID].NewOrders[i].Floor == button.Floor && e_system.Elevators[ID].NewOrders[i].Button_type == button.Button_type {
			exists = 1
		}
	}
	return exists
}

func Check_if_internal_order_exists(button Button, e_system Elevator_System, ID string) int {
	exists := 0
	for i := 0; i < ROWS; i++ {
		if e_system.Elevators[ID].InternalOrders[i].Floor == button.Floor && e_system.Elevators[ID].InternalOrders[i].Button_type == button.Button_type {
			exists = 1
		}
	}
	return exists
}

func Remove_order(current_floor int, e_system *Elevator_System, Orders_Deleted_Chan chan Button) {
	for i := 0; i < ROWS; i++ {
		if e_system.Elevators[e_system.SelfID].InternalOrders[i].Floor == current_floor {
			Orders_Deleted_Chan <- e_system.Elevators[e_system.SelfID].InternalOrders[i]
			e_system.Elevators[e_system.SelfID].InternalOrders[i].Floor = -1
			e_system.Elevators[e_system.SelfID].InternalOrders[i].Button_type = -1
			ESYS_left_shift_orders(i, e_system)
			//Orders_Deleted_Chan <- e_system.Elevators[e_system.SelfID].InternalOrders[i]
		}
	}
}

func ESYS_left_shift_orders(index int, e_system *Elevator_System) {
	for i := index; i < ROWS-1; i++ {
		e_system.Elevators[e_system.SelfID].InternalOrders[i].Floor = e_system.Elevators[e_system.SelfID].InternalOrders[i+1].Floor
		e_system.Elevators[e_system.SelfID].InternalOrders[i].Button_type = e_system.Elevators[e_system.SelfID].InternalOrders[i+1].Button_type
	}
	e_system.Elevators[e_system.SelfID].InternalOrders[ROWS-1].Floor = -1
	e_system.Elevators[e_system.SelfID].InternalOrders[ROWS-1].Button_type = -1
}

func MSG_left_shift_orders(index int, msg *Message) {
	for i := index; i < ROWS-1; i++ {
		msg.InternalOrders[i].Floor = msg.InternalOrders[i+1].Floor
		msg.InternalOrders[i].Button_type = msg.InternalOrders[i+1].Button_type
	}
	msg.InternalOrders[ROWS-1].Floor = -1
	msg.InternalOrders[ROWS-1].Button_type = -1
}

/*
func right_shift_orders(index int) {
	for i := index; i > 0; i-- {
		e_system.elevators[e_system.selfID].InternalOrders[i].Floor = e_system.elevators[e_system.selfID].InternalOrders[i-1].Floor
		e_system.elevators[e_system.selfID].InternalOrders[i].Button_type = e_system.elevators[e_system.selfID].InternalOrders[i-1].Button_type
	}
}
*/
/*
func move_order_infront(index int){
	temp_floor := orders[index].Floor
	temp_button_type := orders[index].Button_type
	orders[index].Floor = -1
	orders[index].Button_type = -1
	right_shift_orders(index)
	orders[0].Floor = temp_floor
	orders[0].Button_type = temp_button_type
}
*/

func New_Order_handler(Button_Press_Chan chan Button, e_system *Elevator_System) {
	for {
		if New_Order_Handler_Flag == true {
			fmt.Println("\n\nGABRIEL\n\n")
			Add_new_order(<-Button_Press_Chan, e_system)
		}
	}
}

func Print_all_orders(e_system Elevator_System) {
	for i := 0; i < ROWS; i++ {
		fmt.Printf("%d,", e_system.Elevators[e_system.SelfID].InternalOrders[i])
	}
	fmt.Printf("\n\n")
}

func Orders_init(e_system *Elevator_System) {
	for i := 0; i < ROWS; i++ {
		e_system.Elevators[e_system.SelfID].InternalOrders[i].Floor = -1
		e_system.Elevators[e_system.SelfID].InternalOrders[i].Button_type = -1
		e_system.Elevators[e_system.SelfID].NewOrders[i].Floor = -1
		e_system.Elevators[e_system.SelfID].NewOrders[i].Button_type = -1
	}
}

func NewOrders_reset(e_system *Elevator_System) {
	for i := 0; i < ROWS; i++ {
		e_system.Elevators[e_system.SelfID].NewOrders[i].Floor = -1
		e_system.Elevators[e_system.SelfID].NewOrders[i].Button_type = -1
	}
}

func Sync_with_system(messageToSlave Message, e_system *Elevator_System, New_Order_Handler_Flag_Chan chan bool, Orders_Deleted_Chan chan Button) {
	var elevExistedInMap bool = false
	for i, _ := range e_system.Elevators {
		if i == messageToSlave.ID {
			if i == e_system.SelfID {
				for j := 0; j < len(Orders_Deleted_Chan); j++ {
					Orders_Deleted := <-Orders_Deleted_Chan
					for k := 0; k < 10; k++ {
						if messageToSlave.InternalOrders[k] == Orders_Deleted {
							//do not want to add order
							//fmt.Println("\n \n Gabriel!!!\n\n")
							messageToSlave.InternalOrders[k] = Button{Button_type: -1, Floor: -1}
							MSG_left_shift_orders(k, &messageToSlave)

						}
					}
				}
			}
			e_system.Elevators[i].InternalOrders = messageToSlave.InternalOrders
			e_system.Elevators[messageToSlave.ID].NewOrders = messageToSlave.NewOrders
			elevExistedInMap = true
			break
		}
	}
	if elevExistedInMap == false {
		e_system.Elevators[messageToSlave.ID] = new(Elevator)
		//fmt.Println("\n\n Noticed a new elevator \n\n")
		e_system.Elevators[messageToSlave.ID].InternalOrders = messageToSlave.InternalOrders
		e_system.Elevators[messageToSlave.ID].NewOrders = messageToSlave.NewOrders
	}
	New_Order_Handler_Flag_Chan <- true
	/*for i, _ := range e_system.Elevators {
		fmt.Println("\nELEVATOR: ", i, ", ORDERS: ", e_system.Elevators[i].InternalOrders)
	}*/
}
