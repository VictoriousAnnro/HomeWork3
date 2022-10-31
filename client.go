package main

//Credit: https://github.com/rrrCode9/gRPC-Bidirectional-Streaming-ChatServer/blob/main/client.go
import (
	"bufio"
	"context"
	"fmt"
	Videobranch "grpcChatServer/chatserver"
	"log"
	"os"
	"strings"

	"google.golang.org/grpc"
)

func main() {

	fmt.Println("Enter Server IP:PORT ::: ")
	reader := bufio.NewReader(os.Stdin)
	serverID, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Failed to read from console :: %v", err)
	}
	serverID = strings.Trim(serverID, "\r\n")

	log.Println("Connecting : " + serverID)

	conn, err := grpc.Dial(serverID, grpc.WithInsecure())

	if err != nil {
		log.Fatalf("Failed to connect to gRPC server :: %v", err)
	}
	defer conn.Close()

	client := Videobranch.NewServicesClient(conn)

	stream, err := client.ChatService(context.Background())
	if err != nil {
		log.Fatalf("Failed to call ChatService :: %v", err)

	}

	ch := clienthandle{stream: stream}
	ch.clientConfig()
	go ch.sendMessage()
	go ch.receiveMessage()

	bl := make(chan bool)
	<-bl

}

type clienthandle struct {
	stream     Videobranch.Services_ChatServiceClient
	clientName string
	lamport    int32
}

func (ch *clienthandle) clientConfig() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Your Name : ")
	name, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf(" Failed to read from console :: %v", err)
	}
	ch.clientName = strings.Trim(name, "\r\n")
	ch.lamport = 1
	clientMessageBox := &Videobranch.FromClient{
		Name:    ch.clientName,
		Body:    "Have joined the channel!!!!",
		Lamport: ch.lamport,
	}
	ch.stream.Send(clientMessageBox)

}

func (ch *clienthandle) sendMessage() {
	for {
		reader := bufio.NewReader(os.Stdin)

		clientMessage, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf(" Failed to read from console :: %v", err)
		}
		clientMessage = strings.Trim(clientMessage, "\r\n")
		ch.lamport++
		clientMessageBox := &Videobranch.FromClient{
			Name:    ch.clientName,
			Body:    clientMessage,
			Lamport: ch.lamport,
		}

		err = ch.stream.Send(clientMessageBox)

		if err != nil {
			log.Printf("Error while sending message to server :: %v", err)
		}

	}

}

func (ch *clienthandle) receiveMessage() {

	for {
		mssg, err := ch.stream.Recv()
		if err != nil {
			log.Printf("Error in reciving message from server :: %v", err)
		}
		fmt.Printf("%s : %s \n", mssg.Name, mssg.Body)
		if mssg.Name == ch.clientName {
			//do nothing as this is just the message it just sent
		} else if mssg.Lamport > ch.lamport {
			ch.lamport = mssg.Lamport
			ch.lamport++
		} else {
			ch.lamport++
		}
		fmt.Print("Current Local Lamport Timestamp: ")
		fmt.Printf("%v", ch.lamport)
		fmt.Println()
		fmt.Println()

	}

}
