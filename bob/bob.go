package main

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"log"
	"math/rand"
	"net"
	"time"

	pb "https://github.itu.dk/mgrp/Security_Handin2/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const port = ":8080"

type server struct {
	pb.UnimplementedDiceGameServer
}

var commitment [32]byte
var roll int32

func (*server) Initiate(ctx context.Context, req *pb.Commitment) (*pb.Value, error) {
	rand.Seed(time.Now().UnixNano())
	copy(commitment[:], req.Commitment)
	roll = int32(rand.Intn(5) + 1)
	return &pb.Value{V: roll}, nil
}

func (*server) Confirmation(ctx context.Context, req *pb.Secrets) (*pb.Ack, error) {
	confirmation := sha256.Sum256([]byte(string(req.R + req.V)))
	if confirmation != commitment {
		return &pb.Ack{Accepted: false}, nil
	}

	finalRoll := ((roll + req.V) % 6) + 1
	log.Printf("Final roll: %d\n", finalRoll)

	return &pb.Ack{Accepted: true}, nil
}

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	// Load server's certificate and private key
	serverCert, err := tls.LoadX509KeyPair("../cert/bob.cert.pem", "../cert/bob.key.pem")
	if err != nil {
		return nil, err
	}

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.NoClientCert,
	}

	return credentials.NewTLS(config), nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	tlsCredentials, err := loadTLSCredentials()
	if err != nil {
		log.Fatal("cannot load TLS credentials: ", err)
	}

	s := grpc.NewServer(
		grpc.Creds(tlsCredentials),
	)
	pb.RegisterDiceGameServer(s, &server{})
	log.Printf("Bob listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
