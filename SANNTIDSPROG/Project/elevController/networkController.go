//need to include message.go
package elevController

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	PINGPORT    int = 32114
	SUPDATEPORT int = 33414
	MUPDATEPORT int = 37814
	//PORT int = 30231

)

type Elevator_System struct {
	SelfID string
	SelfIP string
	//elevators map[int]*Elevator //elevator declared in FSM
	Elevators map[string]*Elevator
	MasterID  int
	MasterIP  string
	Timestamp int
}

type Message struct {
	DestinationFloor int //need to have ALL variables declared by Capital letter in the beginning, due to scope in go!! Check Terminal
	CurrentFloor     int
	ID               string
	Timestamp        int
	InternalOrders   [10]Button
	NewOrders        [10]Button //Please explain
	Master           bool       //Do we need to send this?
	//Master   int
	MasterIP string
}

func Initialize_elev_system() Elevator_System { //Removed elevator *Elevator from function call
	var e_system Elevator_System
	e_system.Elevators = make(map[string]*Elevator)
	addr, _ := net.InterfaceAddrs()
	tempVar := addr[1]
	e_system.Timestamp = 1
	ip := tempVar.String()
	//has a better function for finding selfIP in alive.go.. Check up
	e_system.SelfIP = ip[0:15]
	e_system.SelfID = strconv.Itoa(int(addr[1].String()[12]-'0')*100 + int(addr[1].String()[13]-'0')*10 + int(addr[1].String()[14]-'0')) //this will work for IP-addresses of format ###.###.###.###, but not with only for ###.###.###.##/
	e_system.Elevators[e_system.SelfID] = new(Elevator)
	//e_system.elevators[e_system.selfID].InternalOrders = elevator.InternalOrders
	//fmt.Println("\nelevator.InternalOrders: ", elevator.InternalOrders)
	//fmt.Println("\ne_system.elevators[e_system.selfID].InternalOrders: ", e_system.elevators[e_system.selfID].InternalOrders)

	//before initializing we need to actually listen if we receive something
	masterExists := UDPListenForMasterInit(SUPDATEPORT, &e_system)
	fmt.Println("Does it exists a master: " + strconv.FormatBool(masterExists))
	if masterExists == false {
		Set_master(&e_system)
	}

	//Checking Elev_system variables
	fmt.Println("Master is: " + strconv.Itoa(e_system.MasterID))
	fmt.Printf("Self ID is: %s \n", e_system.SelfID)
	fmt.Printf("Timestamp is: %d \n", e_system.Timestamp)
	fmt.Printf("MasterIP is:" + e_system.MasterIP)
	return e_system
}

func Is_elev_master(e_system Elevator_System) bool {
	isMaster := false
	if e_system.SelfID == strconv.Itoa(e_system.MasterID) {
		isMaster = true
	}
	return isMaster
}

func MessageSetter(Broadcast_Message_Chan chan Message, e_system Elevator_System) {
	var msg Message
	msg.DestinationFloor = e_system.Elevators[e_system.SelfID].DestinationFloor
	msg.CurrentFloor = e_system.Elevators[e_system.SelfID].CurrentFloor
	msg.ID = e_system.SelfID
	msg.MasterIP = e_system.MasterIP
	msg.Timestamp = e_system.Timestamp
	msg.InternalOrders = e_system.Elevators[e_system.SelfID].InternalOrders
	msg.NewOrders = e_system.Elevators[e_system.SelfID].NewOrders
	//fmt.Println(msg.ID)
	//msg.Master = e_system.MasterID
	selfIDint, _ := strconv.Atoi(e_system.SelfID)
	if selfIDint == (e_system.MasterID) {
		msg.Master = true
	} else {
		msg.Master = false
	}
	Broadcast_Message_Chan <- msg
}

func Message_Compiler_Master(msgFromSlave Message, e_system *Elevator_System) {
	var elevExistedInMap bool = false
	for i, _ := range e_system.Elevators {
		if i == msgFromSlave.ID {
			e_system.Elevators[i].NewOrders = msgFromSlave.NewOrders
			e_system.Elevators[i].InternalOrders = msgFromSlave.InternalOrders
			e_system.Elevators[i].CurrentFloor = msgFromSlave.CurrentFloor
			e_system.Elevators[i].DestinationFloor = msgFromSlave.DestinationFloor
			elevExistedInMap = true
			break
		}
	}
	if elevExistedInMap == false {
		fmt.Println("\n\n\nGABRIEL\n\n\n")
		e_system.Elevators[msgFromSlave.ID] = new(Elevator)
		e_system.Elevators[msgFromSlave.ID].InternalOrders = msgFromSlave.InternalOrders
		e_system.Elevators[msgFromSlave.ID].NewOrders = msgFromSlave.NewOrders
		e_system.Elevators[msgFromSlave.ID].CurrentFloor = msgFromSlave.CurrentFloor
		e_system.Elevators[msgFromSlave.ID].DestinationFloor = msgFromSlave.DestinationFloor
	}
}

/*
func RetardNetworkOrderHandler(e_system *Elevator_System) {
	for elev_ID, _ := range e_system.Elevators {
		for index := 0; index < 10; index++ {
			if e_system.Elevators[elev_ID].NewOrders[index].Floor == -1 && e_system.Elevators[elev_ID].NewOrders[index].Button_type == -1 {
				continue
				//break
			} else {
				if Check_if_internal_order_exists(e_system.Elevators[elev_ID].NewOrders[index], *e_system, elev_ID) == 1 {
					e_system.Elevators[elev_ID].NewOrders[index].Floor = -1
					e_system.Elevators[elev_ID].NewOrders[index].Button_type = -1
					//break
				} else {
					for i := 0; i < 10; i++ {
						if e_system.Elevators[elev_ID].InternalOrders[i].Floor == -1 {
							e_system.Elevators[elev_ID].InternalOrders[i].Floor = e_system.Elevators[elev_ID].NewOrders[index].Floor
							e_system.Elevators[elev_ID].InternalOrders[i].Button_type = e_system.Elevators[elev_ID].NewOrders[index].Button_type
							e_system.Elevators[elev_ID].NewOrders[index].Floor = -1
							e_system.Elevators[elev_ID].NewOrders[index].Button_type = -1
							break
						}
					}
				}
			}
		}
	}
}
*/
func NetworkOrderHandler(e_system *Elevator_System) {
	for elev_ID, _ := range e_system.Elevators {
		for i := 0; i < 10; i++ {
			if e_system.Elevators[elev_ID].NewOrders[i].Floor == -1 && e_system.Elevators[elev_ID].NewOrders[i].Button_type == -1 {
				break
			}
			if e_system.Elevators[elev_ID].NewOrders[i].Button_type == 2 {
				if Check_if_internal_order_exists(e_system.Elevators[elev_ID].NewOrders[i], *e_system, elev_ID) == 1 {
					e_system.Elevators[elev_ID].NewOrders[i].Floor = -1
					e_system.Elevators[elev_ID].NewOrders[i].Button_type = -1
					fmt.Println("\nThis order exists")
					//break
				} else {
					for j := 0; j < 10; j++ {
						if e_system.Elevators[elev_ID].InternalOrders[j].Floor == -1 {
							e_system.Elevators[elev_ID].InternalOrders[j].Floor = e_system.Elevators[elev_ID].NewOrders[i].Floor
							e_system.Elevators[elev_ID].InternalOrders[j].Button_type = e_system.Elevators[elev_ID].NewOrders[i].Button_type
							e_system.Elevators[elev_ID].NewOrders[i].Floor = -1
							e_system.Elevators[elev_ID].NewOrders[i].Button_type = -1
							break
						}
					}
				}
			} else {
				var min_cost float64 = math.Inf(1)
				var optimal_elev string
				var order_exists int = 0
				NewOrder := e_system.Elevators[elev_ID].NewOrders[i]
				for e_ID, _ := range e_system.Elevators {
					order_exists = Check_if_internal_order_exists(e_system.Elevators[elev_ID].NewOrders[i], *e_system, e_ID)
					if order_exists == 1 {
						e_system.Elevators[elev_ID].NewOrders[i].Floor = -1
						e_system.Elevators[elev_ID].NewOrders[i].Button_type = -1
						fmt.Println("\nThis order exists")
						break
					}
					cost := CostFunction(*e_system, NewOrder, e_ID)
					if cost <= min_cost {
						min_cost = cost
						optimal_elev = e_ID
					}
				}
				if order_exists == 1 {
					break
				} else {
					fmt.Println("\n \n \n This is the closest elevator: ", optimal_elev)
					for k := 0; k < 10; k++ {
						if e_system.Elevators[optimal_elev].InternalOrders[k].Floor == -1 {
							e_system.Elevators[optimal_elev].InternalOrders[k].Floor = e_system.Elevators[elev_ID].NewOrders[i].Floor
							e_system.Elevators[optimal_elev].InternalOrders[k].Button_type = e_system.Elevators[elev_ID].NewOrders[i].Button_type
							e_system.Elevators[elev_ID].NewOrders[i].Floor = -1
							e_system.Elevators[elev_ID].NewOrders[i].Button_type = -1
							break
						}
					}
				}
			}
		}
	}
}

func CostFunction(e_system Elevator_System, NewOrder Button, elev_ID string) float64 {
	// One distance costs 1 and equals the distance between two elevators.
	// Stopping at a floor also costs 1.
	var cost float64 = 0.00
	var order_diff float64
	if e_system.Elevators[elev_ID].InternalOrders[0].Floor != -1 {
		for i := 0; i < 9; i++ {
			if i == 0 {
				order_diff = float64(e_system.Elevators[elev_ID].CurrentFloor - e_system.Elevators[elev_ID].InternalOrders[i].Floor)
				cost += math.Abs(order_diff) + 1.0
			} else if e_system.Elevators[elev_ID].InternalOrders[i+1].Floor != -1 {
				order_diff = float64(e_system.Elevators[elev_ID].InternalOrders[i].Floor - NewOrder.Floor)
				cost += math.Abs(order_diff)
			} else {
				order_diff = float64(e_system.Elevators[elev_ID].InternalOrders[i].Floor - e_system.Elevators[elev_ID].InternalOrders[i+1].Floor)
				cost += math.Abs(order_diff) + 1.0
			}
		}
	} else {
		order_diff = float64(e_system.Elevators[elev_ID].CurrentFloor - NewOrder.Floor)
		cost = math.Abs(order_diff)
	}
	return cost
}

/*
func Remove_elev(ID string, e *Elevator_System) {
	delete(e.elevators, ID)
	fmt.Println("Elevator ", ID, " removed from network")
	Set_master(e)
}
*/

func Add_elev(ID string, e_system *Elevator_System) {
	e_system.Elevators[ID] = new(Elevator)
}

func Set_master(e_system *Elevator_System) {
	// Checking which elevator has the highest IP to determine who is the master
	max := 0
	for i, _ := range e_system.Elevators {
		j, _ := strconv.Atoi(i)
		if max < j {
			max = j

		}
	}
	e_system.MasterID = max
	e_system.Timestamp = 1
	var tempIP string = e_system.SelfIP[0:12]
	e_system.MasterIP = tempIP + strconv.Itoa(e_system.MasterID)
	fmt.Println("New master is", e_system.MasterID)
}

func Int_Timer_Chan(Timer_Chan chan int, n int) {
	timer := time.NewTimer(time.Millisecond * time.Duration(n))
	<-timer.C
	//fmt.Println("Timer timeout")
	Timer_Chan <- 1
}

func String_Timer_Chan(Timer_Chan chan string, n int) {
	timer := time.NewTimer(time.Millisecond * time.Duration(n))
	<-timer.C
	//fmt.Println("Timer timeout")
	Timer_Chan <- "1"
}

/* A Simple function to verify error */
func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
	}
}

func UDPListenForMasterInit(listenPort int, e_system *Elevator_System) bool {
	ServerAddr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(listenPort))
	CheckError(err)
	var masterExists bool

	/* Now listen at selected port */
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	defer ServerConn.Close()
	buf := make([]byte, 1024)
	ServerConn.SetDeadline(time.Now().Add(time.Millisecond * 1000))
	_, addr, err1 := ServerConn.ReadFromUDP(buf)
	if err1 != nil {
		masterExists = false
	} else {
		e_system.MasterIP = strings.Split(addr.String(), ":")[0]
		masterID, _ := strconv.Atoi(e_system.MasterIP[12:]) //little bit insecure about the ParseInt function
		//fmt.Printf("\nDummyint contains: %d,",masterID)
		e_system.MasterID = masterID
		masterExists = true
	} //removed for-loop containing breaks;(if not responding)
	return masterExists

}

func UDPListenForPing(listenPort int, e_system Elevator_System, From_Master_ReqSys_Chan chan int) {

	//ServerAddr, err := net.ResolveUDPAddr("udp", ":40000")
	ServerAddr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(listenPort))
	CheckError(err)

	/* Now listen at selected port */
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	defer ServerConn.Close()

	buffer := make([]byte, 1024)
	trimmed_buffer := make([]byte, 1)
	for { // what if the elevators crashes and the slave becomes the master
		n, addr, err := ServerConn.ReadFromUDP(buffer)
		//fmt.Printf("\naddr: %d", addr.String())
		//fmt.Printf("\ne_system.masterIP: %d", e_system.masterIP)
		if strings.Split(addr.String(), ":")[0] == e_system.MasterIP {
			//fmt.Println("hei")
			trimmed_buffer = buffer[0:n]
			i := string(trimmed_buffer)
			//fmt.Println("i:", i, "  i == \"1\":", i == "1")
			if i == "1" {
				//fmt.Println("\nPing received ", i, " from ", addr)
				CheckError(err)

				//err = json.Unmarshal(trimmed_buffer, &received_message)
				From_Master_ReqSys_Chan <- 1
				time.Sleep(time.Millisecond * 2)
			}
		}
	}
}

func UDPListenForUpdateMaster(listenPort int, infoRec chan Message) {

	/* For testing: sett addresse lik ip#255:30000*/
	//ServerAddr, err := net.ResolveUDPAddr("udp", ":40000")
	ServerAddr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(listenPort))
	CheckError(err)

	/* Now listen at selected port */
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	defer ServerConn.Close()

	var received_message Message

	//var storageChannel := make(chan Message)
	buffer := make([]byte, 1024)
	trimmed_buffer := make([]byte, 1)
	for {
		n, _, err := ServerConn.ReadFromUDP(buffer)
		trimmed_buffer = buffer[0:n]
		//fmt.Println("Received ", string(buffer[0:n]), " from ", addr)
		CheckError(err)
		//fmt.Println("\nReceived ", received_message)
		err = json.Unmarshal(trimmed_buffer, &received_message)
		CheckError(err)
		//fmt.Println("\nReceived ", received_message)
		infoRec <- received_message
		time.Sleep(time.Millisecond * 1)
	}

}

func UDPListenForUpdateSlave(listenPort int, e_system *Elevator_System, From_Master_NewUpdate_Chan chan Message) {

	//For testing: sett addresse lik ip#255:30000
	//ServerAddr, err := net.ResolveUDPAddr("udp", ":40000")
	ServerAddr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(listenPort))
	CheckError(err)

	// Now listen at selected port
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	defer ServerConn.Close()

	var msg Message

	buffer := make([]byte, 1024)
	trimmed_buffer := make([]byte, 1)

	for { // what if the elevators crashes and the slave becomes the master
		n, addr, err := ServerConn.ReadFromUDP(buffer)
		if strings.Split(addr.String(), ":")[0] == e_system.MasterIP {
			trimmed_buffer = buffer[0:n]
			//fmt.Println("Received ", string(buffer[0:n]), " from ", addr)
			CheckError(err)
			err = json.Unmarshal(trimmed_buffer, &msg)
			CheckError(err)
			//fmt.Println("\n Received this message from master: ", msg)
			From_Master_NewUpdate_Chan <- msg
		}
	}
}

func UDPSendReqToSlaves(transmitPort int, ping string) {
	BroadcastAddr, err := net.ResolveUDPAddr("udp", "255.255.255.255:"+strconv.Itoa(transmitPort))
	CheckError(err)

	/* Create a connection to the server */
	Conn, err := net.DialUDP("udp", nil, BroadcastAddr)
	CheckError(err)

	//fmt.Println("This is the master")
	defer Conn.Close()

	Conn.Write([]byte(ping))
}

func UDPSendSysInfoToSlaves(transmitPort int, reworkedSystem Elevator_System) {
	BroadcastAddr, err := net.ResolveUDPAddr("udp", "255.255.255.255:"+strconv.Itoa(transmitPort))
	CheckError(err)

	/* Create a connection to the server */
	Conn, err := net.DialUDP("udp", nil, BroadcastAddr)
	CheckError(err)

	defer Conn.Close()

	var msg Message
	for i, _ := range reworkedSystem.Elevators {
		msg.DestinationFloor = reworkedSystem.Elevators[i].DestinationFloor
		msg.CurrentFloor = reworkedSystem.Elevators[i].CurrentFloor
		msg.ID = i
		msg.MasterIP = reworkedSystem.MasterIP
		msg.Timestamp = reworkedSystem.Timestamp
		msg.InternalOrders = reworkedSystem.Elevators[i].InternalOrders
		msg.NewOrders = reworkedSystem.Elevators[i].NewOrders
		//need to set bool Master in msg
		selfIDint, _ := strconv.Atoi(reworkedSystem.SelfID)
		if selfIDint == (reworkedSystem.MasterID) {
			msg.Master = true
		} else {
			msg.Master = false
		}
		//fmt.Println("Sent out this message : ", msg)

		buf, err := json.Marshal(msg)
		Conn.Write(buf)
		CheckError(err)
	}
}

func UDPSendToMaster(transmitPort int, broadcastMessage_Chan chan Message, New_Order_Handler_Flag_Chan chan bool) {
	msg := <-broadcastMessage_Chan
	//fmt.Println("\n MSG: ",msg,"\n")
	//fmt.Println("This is sent from slave :", msg)
	MasterAddr, err := net.ResolveUDPAddr("udp", msg.MasterIP+":"+strconv.Itoa(transmitPort))
	CheckError(err)

	/* Create a connection to the server */
	Conn, err := net.DialUDP("udp", nil, MasterAddr)
	CheckError(err)

	//fmt.Println("This is a slave")
	defer Conn.Close()
	//msg.timestamp = broadcastMessage.timestamp

	/* Loads the buffer with the message in json-format */
	buf, err := json.Marshal(msg)
	CheckError(err)

	New_Order_Handler_Flag_Chan <- false
	Conn.Write(buf)
}
