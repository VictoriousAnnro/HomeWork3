package main

import (
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
)

func main() {

	Port := os.Getenv("PORT")
	if Port == "" {
		Port = "5000"
	}

	listen, err := net.Listen("tcp", ":"+Port)
	if err != nil {
		log.Fatalf("Could not listen on @ %v :: %v", Port, err)
	}
	log.Println("Listening @ : " + Port)

	grpcserver := grpc.NewServer()

	cs := chatserver.ChatServer{}
	chatserver.RegisterServicesServer(grpcserver, &cs)

	err = grpcserver.Serve(listen)
	if err != nil {
		log.Fatalf("Failed to start gRPC server :: %v", err)
	}

}
