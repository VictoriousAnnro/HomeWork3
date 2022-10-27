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

	fmt.Println("--- Enter a Username to Join Chat ---")
	fmt.Printf("Your Name : ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Failed to read from console :: %v", err)
	}
	clientNameInput := strings.Trim(input, "\r\n")

	serverID := "localhost:5000" //strings.Trim(serverID, "\r\n")

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
	ch.joinChat(clientNameInput)
	go ch.sendMessage()
	go ch.receiveMessage()

	bl := make(chan bool)
	<-bl
}

type clienthandle struct {
	stream     Videobranch.Services_ChatServiceClient
	clientName string
}

func (ch *clienthandle) joinChat(clientNameInput string) {
	ch.clientName = clientNameInput
	clientMessageBox := &Videobranch.FromClient{
		Name: ch.clientName,
		Body: "May I join?? uwu", //"Has joined the channel!!"
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

		clientMessageBox := &Videobranch.FromClient{
			Name: ch.clientName,
			Body: clientMessage,
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
			log.Printf("Error in reciving message from server :: %v", err, ch.clientName)
		}
		fmt.Printf("%s : %s \n", mssg.Name, mssg.Body)

	}

}
