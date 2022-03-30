package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func printBackends() {
	lb.strategy.PrintTopology()
}

const CMD_StrategyChange = "strategy/change"
const CMD_StrategyEdit = "strategy/edit"
const CMD_BackendAdd = "backend/add"
const CMD_TopologyList = "topology/list"
const CMD_TopologyTest = "topology/test"
const CMD_Exit = "exit"

var commands []string = []string{
	CMD_StrategyChange,
	CMD_StrategyEdit,
	CMD_BackendAdd,
	CMD_TopologyList,
	CMD_TopologyTest,
	CMD_Exit,
}

func cli() {
	for {
		var command string
		fmt.Print(">>> ")
		fmt.Scanf("%s", &command)
		switch command {
		case CMD_Exit:
			lb.events <- Event{EventName: CMD_Exit}
			// TODO: this is not idea. End this gracefully.
			return
		case CMD_BackendAdd:
			var host string
			var port int

			fmt.Print("       Host: ")
			fmt.Scanf("%s", &host)

			fmt.Print("       Port: ")
			fmt.Scanf("%d", &port)

			lb.events <- Event{EventName: CMD_BackendAdd, Data: Backend{Host: host, Port: port}}
		case CMD_StrategyChange:
			var strategy string

			fmt.Print("       Name of the strategy: ")
			fmt.Scanf("%s", &strategy)

			lb.events <- Event{EventName: CMD_StrategyChange, Data: strategy}
		case CMD_StrategyEdit:
			if strategy, isOk := lb.strategy.(*StaticBalancingStrategy); isOk {
				var index int
				fmt.Print("       Index of Backend to be active: ")
				fmt.Scanf("%d", &index)
				strategy.Index = index
			} else {
				fmt.Println("this balancing strategy does not support edits.")
			}
		case CMD_TopologyTest:
			var reqId string

			fmt.Print("       Request ID: ")
			fmt.Scanf("%s", &reqId)

			backend := lb.strategy.GetNextBackend(IncomingReq{reqId: reqId})
			fmt.Printf("request: %s goes to backend: %s\n", reqId, backend)
		case CMD_TopologyList:
			lb.strategy.PrintTopology()
		default:
			fmt.Printf("available commands: %s\n", strings.Join(commands, ", "))
		}
	}
}

func setLogging() {
	f, err := os.OpenFile("lb.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
		panic(err)
	}
	log.SetOutput(f)

	// TODO: Closse the log file.
}

func main() {
	setLogging()
	InitLB()
	go lb.Run()
	cli()
}
