package Videobranch

//Credit: https://github.com/rrrCode9/gRPC-Bidirectional-Streaming-ChatServer/blob/main/client.go
import (
	"log"
	"math/rand"
	"sync"
	"time"
)

type messageUnit struct {
	ClientName        string
	MessageBody       string
	MessageUniqueCode int
	ClientUniqueCode  int
}

type messageHandle struct {
	MQue []messageUnit
	mu   sync.Mutex
}

var messageHandleObject = messageHandle{}
var chatServiceHandler = []Services_ChatServiceServer{}

type ChatServer struct {
}

func (is *ChatServer) ChatService(csi Services_ChatServiceServer) error {

	clientUniqueCode := rand.Intn(1e6)
	errch := make(chan error)

	go recieveFromStream(csi, clientUniqueCode, errch)
	go sendToStream(csi, clientUniqueCode, errch)
	chatServiceHandler = append(chatServiceHandler, csi)
	log.Printf("%v", "A new client has joined the chat")

	return <-errch

}

func recieveFromStream(csi_ Services_ChatServiceServer, clientUniqueCode_ int, errch_ chan error) {

	for {
		mssg, err := csi_.Recv()
		if err != nil {
			log.Printf("Error in reciving message from client :: %v", err)
			errch_ <- err
		} else {

			messageHandleObject.mu.Lock()

			messageHandleObject.MQue = append(messageHandleObject.MQue, messageUnit{
				ClientName:        mssg.Name,
				MessageBody:       mssg.Body,
				MessageUniqueCode: rand.Intn(1e8),
				ClientUniqueCode:  clientUniqueCode_,
			})

			messageHandleObject.mu.Unlock()
			log.Printf("%v", messageHandleObject.MQue[len(messageHandleObject.MQue)-1].ClientName+" Has sent a message")
			log.Printf("%v", messageHandleObject.MQue[len(messageHandleObject.MQue)-1])
		}

	}

}

func sendToStream(csi_ Services_ChatServiceServer, clientUniqueCode_ int, errch_ chan error) {

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

			messageHandleObject.mu.Unlock()

			//err := csi_.Send(&FromServer{Name: senderName4Client, Body: message4Client})
			for i := 0; i < len(chatServiceHandler); i++ {
				err := chatServiceHandler[i].Send(&FromServer{Name: senderName4Client, Body: message4Client})
				log.Printf("%v", "Server distributed a message from: "+senderName4Client)
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
