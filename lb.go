package main

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/google/uuid"
)

type Backend struct {
	Host        string
	Port        int
	IsHealthy   bool
	NumRequests int
}

func (b *Backend) String() string {
	return fmt.Sprintf("%s:%d", b.Host, b.Port)
}

type Event struct {
	EventName string
	Data      interface{}
}

type LB struct {
	backends []*Backend
	events   chan Event
	strategy BalancingStrategy
}

type IncomingReq struct {
	srcConn net.Conn
	reqId   string
}

var lb *LB

func InitLB() {
	backends := []*Backend{
		&Backend{Host: "localhost", Port: 8081, IsHealthy: true},
		&Backend{Host: "localhost", Port: 8082, IsHealthy: true},
		&Backend{Host: "localhost", Port: 8083, IsHealthy: true},
		&Backend{Host: "localhost", Port: 8084, IsHealthy: true},
	}
	lb = &LB{
		events:   make(chan Event),
		backends: backends,
		strategy: NewHashedBalancingStrategy(backends),
	}
}

func (lb *LB) Run() {
	listener, err := net.Listen("tcp", ":9090")
	if err != nil {
		panic(err)
	}

	defer listener.Close()
	log.Println("LB listening on port 9090 ...")

	go func() {
		for {
			select {
			case event := <-lb.events:
				switch event.EventName {
				case CMD_Exit:
					log.Println("gracefully terminating ...")
					return
				case CMD_BackendAdd:
					backend, isOk := event.Data.(Backend)
					if !isOk {
						panic(err)
					}
					lb.backends = append(lb.backends, &backend)
					lb.strategy.RegisterBackend(&backend)
				case CMD_StrategyChange:
					strategyName, isOk := event.Data.(string)
					if !isOk {
						panic(err)
					}
					switch strategyName {
					case "round-robin":
						lb.strategy = NewRRBalancingStrategy(lb.backends)
					case "static":
						lb.strategy = NewStaticBalancingStrategy(lb.backends)
					case "hash":
						lb.strategy = NewHashedBalancingStrategy(lb.backends)
					default:
						lb.strategy = NewHashedBalancingStrategy(lb.backends)
					}
				}
			}
		}
	}()

	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Printf("unable to accept connection: %s", err.Error())
			panic(err)
		}

		// log.Printf("local: %v remote: %v", connection.LocalAddr().String(), connection.RemoteAddr().String())

		// Once the connection is accepted proxying it to backend
		go lb.proxy(IncomingReq{
			srcConn: connection,
			reqId:   uuid.NewString(),
		})
	}
}

func (lb *LB) proxy(req IncomingReq) {
	// Get backend sserver depending on some algorithm
	backend := lb.strategy.GetNextBackend(req)
	log.Printf("in-req: %s out-req: %s", req.reqId, backend.String())

	// Setup backend connection
	backendConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", backend.Host, backend.Port))
	if err != nil {
		log.Printf("error connecting to backend: %s", err.Error())

		// send back error to src
		req.srcConn.Write([]byte("backend not available"))
		req.srcConn.Close()
		panic(err)
	}

	backend.NumRequests++
	go io.Copy(backendConn, req.srcConn)
	go io.Copy(req.srcConn, backendConn)
}
