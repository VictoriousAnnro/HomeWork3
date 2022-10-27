package Videobranch

//Credit: https://github.com/rrrCode9/gRPC-Bidirectional-Streaming-ChatServer/blob/main/client.go
import (
	"log"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type messageUnit struct {
	ClientName  string
	MessageBody string
}

type messageHandle struct {
	MQue []messageUnit
	mu   sync.Mutex
}

type chatserviceHandle struct {
	ClientMap map[string]Services_ChatServiceServer
	lo        sync.Mutex
}

var messageHandleObject = messageHandle{}
var chatserviceHandleObject = chatserviceHandle{ClientMap: make(map[string]Services_ChatServiceServer)}

type ChatServer struct {
	cName string
}

func (is *ChatServer) ChatService(csi Services_ChatServiceServer) error {

	//clientUniqueCode := rand.Intn(1e6)
	errch := make(chan error)

	go recieveFromStream(csi, is, errch)
	go sendToStream(errch) //make global somehow?

	return <-errch

}

func recieveFromStream(csi_ Services_ChatServiceServer, is *ChatServer, errch_ chan error) {

	for {
		mssg, err := csi_.Recv()
		if err != nil {
			log.Printf("Error in reciving message from client :: %v", err)
			if status.Code(err) == codes.Canceled { //does the code say the context has been cancelled, aka client disconnected?
				removeClient(is)
			}
			errch_ <- err
		} else {
			//tjek om join request
			if mssg.Body == "May I join?? uwu" {
				is.cName = mssg.Name

				chatserviceHandleObject.lo.Lock()
				chatserviceHandleObject.ClientMap[mssg.Name] = csi_
				chatserviceHandleObject.lo.Unlock()
				mssg.Body = "Has joined the channel!"
			}

			//make this into sep. method, to avoid duplicate code in removeClient()? - later
			messageHandleObject.mu.Lock()

			messageHandleObject.MQue = append(messageHandleObject.MQue, messageUnit{
				ClientName:  mssg.Name,
				MessageBody: mssg.Body,
			})

			messageHandleObject.mu.Unlock()
			log.Printf("%v", messageHandleObject.MQue[len(messageHandleObject.MQue)-1])

		}

	}

}

func removeClient(is *ChatServer) {

	chatserviceHandleObject.lo.Lock()
	delete(chatserviceHandleObject.ClientMap, is.cName) //remove client from list
	chatserviceHandleObject.lo.Unlock()

	messageHandleObject.mu.Lock()

	messageHandleObject.MQue = append(messageHandleObject.MQue, messageUnit{
		ClientName:  is.cName,
		MessageBody: "Has left the chat",
	})
	//prints this message 2 times for some reason??

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

			messageHandleObject.mu.Unlock()

			for _, stream := range chatserviceHandleObject.ClientMap {

				err := stream.Send(&FromServer{Name: senderName4Client, Body: message4Client})

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
