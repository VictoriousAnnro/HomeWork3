package Videolamportbranch

//Credit: https://github.com/rrrCode9/gRPC-Bidirectional-Streaming-ChatServer/blob/main/client.go
import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type messageUnit struct {
	ClientName  string
	MessageBody string
	Lamport     int32
}

type messageHandle struct {
	MQue []messageUnit
	mu   sync.Mutex
}

type chatserviceHandle struct {
	ClientMap map[int]clienthandle
	lo        sync.Mutex
}

type clienthandle struct {
	clientStream Services_ChatServiceServer
	cName        string
	id           int
}

var messageHandleObject = messageHandle{}
var chatserviceHandleObject = chatserviceHandle{ClientMap: make(map[int]clienthandle)}

type ChatServer struct {
}

func (is *ChatServer) ChatService(csi Services_ChatServiceServer) error {

	clientUniqueCode := rand.Intn(1e6)
	errch := make(chan error)

	go recieveFromStream(csi, clientUniqueCode, errch)
	go sendToStream(errch)

	return <-errch

}

func recieveFromStream(csi_ Services_ChatServiceServer, clientUniqueCode int, errch_ chan error) {

	for {
		mssg, err := csi_.Recv()

		if status.Code(err) == codes.Canceled {
			removeClient(clientUniqueCode)
			break
		}

		if err != nil {
			log.Printf("Error in reciving message from client :: %v", err)
			errch_ <- err
		} else {
			//tjek om join request
			if mssg.Body == "May I join?? uwu" {

				client := clienthandle{
					clientStream: csi_,
					cName:        mssg.Name,
					id:           clientUniqueCode,
				}

				chatserviceHandleObject.lo.Lock()
				chatserviceHandleObject.ClientMap[clientUniqueCode] = client
				chatserviceHandleObject.lo.Unlock()
				mssg.Body = "Has joined the channel!"
			}

			messageHandleObject.mu.Lock()

			messageHandleObject.MQue = append(messageHandleObject.MQue, messageUnit{
				ClientName:  mssg.Name,
				MessageBody: mssg.Body,
				Lamport:     mssg.Lamport,
			})

			messageHandleObject.mu.Unlock()
			log.Printf("%v", fmt.Sprint(messageHandleObject.MQue[len(messageHandleObject.MQue)-1], " Server received Lamport Value: ", mssg.Lamport))

		}

	}

}

func removeClient(clientUniqueCode int) {
	name := chatserviceHandleObject.ClientMap[clientUniqueCode].cName

	chatserviceHandleObject.lo.Lock()
	delete(chatserviceHandleObject.ClientMap, clientUniqueCode) //remove client from list
	log.Printf("removing client: %v", name)
	chatserviceHandleObject.lo.Unlock()

	messageHandleObject.mu.Lock()

	messageHandleObject.MQue = append(messageHandleObject.MQue, messageUnit{
		ClientName:  name,
		MessageBody: "Has left the chat",
	})

	messageHandleObject.mu.Unlock()
	log.Printf("%v", messageHandleObject.MQue[len(messageHandleObject.MQue)-1])

}

func sendToStream(errch_ chan error) {

	for {
		for {

			time.Sleep(500 * time.Millisecond)

			messageHandleObject.mu.Lock()

			if len(messageHandleObject.MQue) == 0 {
				messageHandleObject.mu.Unlock()
				break
			}

			senderName4Client := messageHandleObject.MQue[0].ClientName
			message4Client := messageHandleObject.MQue[0].MessageBody
			lamport4Client := messageHandleObject.MQue[0].Lamport

			messageHandleObject.mu.Unlock()

			for _, clientH := range chatserviceHandleObject.ClientMap {
				log.Printf("%v", fmt.Sprint("Server Sending the Message along with Lamport Value: '", lamport4Client+1, "' to client: ", clientH.cName))
				err := clientH.clientStream.Send(&FromServer{Name: senderName4Client, Body: message4Client, Lamport: lamport4Client + 1})

				if err != nil {
					errch_ <- err
				}
			}

			messageHandleObject.mu.Lock()

			if len(messageHandleObject.MQue) > 1 {
				messageHandleObject.MQue = messageHandleObject.MQue[1:]
			} else {
				messageHandleObject.MQue = []messageUnit{}
			}

			messageHandleObject.mu.Unlock()

		}

		time.Sleep(100 * time.Millisecond)
	}
}
