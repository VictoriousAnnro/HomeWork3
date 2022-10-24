package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	// this has to be the same as the go.mod module,
	// followed by the path to the folder the proto file is in.
	gRPC "github.com/VictoriousAnnro/HomeWork3/gRPCServ/proto"

	"google.golang.org/grpc"
)

type Server struct {
	gRPC.UnimplementedPublishServer        // You need this line if you have a server
	name                            string // Not required but useful if you want to name your server
	port                            string // Not required but useful if your server needs to know what port it's listening to

	//currentTime time.Time
	pubMessage  string
	mutex       sync.Mutex // used to lock the server to avoid race conditions.
	messageList map[string]gRPC.Publish_JoinServerServer
}

// flags are used to get arguments from the terminal. Flags take a value, a default value and a description of the flag.
// to use a flag then just add it as an argument when running the program.
var serverName = flag.String("name", "default", "Senders name") // set with "-name <name>" in terminal
// go run .\client\ -name NavnHer
var port = flag.String("port", "5400", "Server port") // set with "-port <port>" in terminal

func main() {

	// setLog() //uncomment this line to log to a log.txt file instead of the console

	// This parses the flags and sets the correct/given corresponding values.
	flag.Parse()
	fmt.Println(".:server is starting:.")

	// starts a goroutine executing the launchServer method.
	go launchServer()

	// This makes sure that the main method is "kept alive"/keeps running
	for {
		time.Sleep(time.Second * 5)
	}
}

func launchServer() {
	log.Printf("Server %s: Attempts to create listener on port %s\n", *serverName, *port)

	// Create listener tcp on given port or default port 5400
	list, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", *port))
	if err != nil {
		log.Printf("Server %s: Failed to listen on port %s: %v", *serverName, *port, err) //If it fails to listen on the port, run launchServer method again with the next value/port in ports array
		return
	}

	// makes gRPC server using the options
	// you can add options here if you want or remove the options part entirely
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	// makes a new server instance using the name and port from the flags.
	server := &Server{
		name:        *serverName,
		port:        *port,
		messageList: make(map[string]gRPC.Publish_JoinServerServer),
	}

	gRPC.RegisterPublishServer(grpcServer, server) //Registers the server to the gRPC server.

	log.Printf("Server %s: Listening on port %s\n", *serverName, *port)

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to serve %v", err)
	}
	// code here is unreachable because grpcServer.Serve occupies the current thread.
}

func (s *Server) JoinServer(request *gRPC.Request, pms gRPC.Publish_JoinServerServer) error {
	//hvis join, tilføj pms til messageList
	s.messageList[request.ClientName] = pms

	fmt.Print("client has joined: ", request.ClientName)

	return nil //lav lige ordentligt error handling
	//gRPC returnerer automatisk stream, dont worry abt writing it here
}

func (s *Server) PublishMessage(ctx context.Context, request *gRPC.Request) (*gRPC.Ack, error) {
	//s.currentTime = time.Now()
	//fmt.Print(request.ClientInput)
	//s.pubMessage = request.ClientInput
	returnMess := &gRPC.Mess{
		Message: "This would be input", //request.ClientInput,
	}

	//pms.Send(s.pubMessage) - iterær over pms og send til hver
	for _, val := range s.messageList {
		val.Send(returnMess)
		//fmt.Print(key, val)
	}

	//ack :=  // make an instance of your return type
	return &gRPC.Ack{AckMessage: "Message published!"}, nil
}
