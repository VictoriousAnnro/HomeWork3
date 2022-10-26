package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	gRPC "github.com/VictoriousAnnro/HomeWork3/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

/*To run server, go to gRPCServ folder, open 2 terminals and use these commands:
go run .\server\server.go
go run .\client\client.go
- In separate terminals! And order is important!*/

// Same principle as in client. Flags allows for user specific arguments/values
var clientsName = flag.String("name", "default", "Senders name")
var serverPort = flag.String("server", "5400", "Tcp server")

var server gRPC.PublishClient   //the server
var ServerConn *grpc.ClientConn //the server connection
var stream gRPC.Publish_JoinServerClient

func main() {
	//parse flag/arguments
	flag.Parse()

	fmt.Println("--- CLIENT APP ---")

	//log to file instead of console
	//setLog()

	//connect to server and close the connection when program closes
	fmt.Println("--- join Server ---")
	ConnectToServer()
	defer ServerConn.Close()

	//start the biding
	parseInput()
}

// connect to server
func ConnectToServer() {

	//dial options
	//the server is not using TLS, so we use insecure credentials
	//(should be fine for local testing but not in the real world)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	//use context for timeout on the connection
	timeContext, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel() //cancel the connection when we are done

	//dial the server to get a connection to it
	log.Printf("client %s: Attempts to dial on port %s\n", *clientsName, *serverPort)
	conn, err := grpc.DialContext(timeContext, fmt.Sprintf(":%s", *serverPort), opts...)
	if err != nil {
		log.Printf("Fail to Dial : %v", err)
		return
	}

	// makes a client from the server connection and saves the connection
	// and prints rather or not the connection was is READY
	server = gRPC.NewPublishClient(conn)
	ServerConn = conn
	log.Println("the connection is: ", conn.GetState().String())

	request := &gRPC.Request{
		ClientName:  *clientsName,
		ClientInput: "",
	}
	stream, _ = server.JoinServer(context.Background(), request) //- save stream and launch goroutine

}

func parseInput() {
	go func() {
		for {
			resp, err := stream.Recv()

			if err != nil {
				continue
			}

			fmt.Print(resp)
			//tjek om respons ik er nul?
		}
	}()

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("type your message. Press 'Enter' to publish")
	fmt.Println("--------------------")

	//Infinite loop to listen for clients input.
	for {
		go func() {
			for {
				resp, err := stream.Recv()
				fmt.Printf("stream: %v\n", stream)

				/*if err != nil {
					continue
				}
				fmt.Print("response:", resp)*/
				//tjek om respons ik er nul?
			}
		}()

		fmt.Print("-> ")

		//Read input into var input and any errors into err
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		input = strings.TrimSpace(input) //Trim input

		if !conReady(server) {
			log.Printf("Client %s: something was wrong with the connection to the server :(", *clientsName)
			continue
		}

		GetTheTime(input)
	}
}

func GetTheTime(input string) {
	//create request type
	request := &gRPC.Request{
		ClientName:  *clientsName,
		ClientInput: input,
	}

	//Make gRPC call to server with input, and recieve acknowlegdement back.
	ack, err := server.PublishMessage(context.Background(), request)
	if err != nil {
		log.Printf("Client %s: no response from the server, attempting to reconnect", *clientsName)
		log.Println(err)
	}
	//fmt.Print(ack)
	fmt.Print("Success, acknowledgement has been received: ", ack.AckMessage, "\n")
}

// Function which returns a true boolean if the connection to the server is ready, and false if it's not.
func conReady(s gRPC.PublishClient) bool {
	return ServerConn.GetState().String() == "READY"
}
