package main

import (
	"context"
	"crypto/sha256"
	"log"
	"math/rand"
	"net"
	"time"

	pb "github.com/markusgp/mandatory2/grpc"
	"google.golang.org/grpc"
)

const port = ":8081"

type server struct {
	pb.UnimplementedDiceGameServer
}

func (*server) Initiate(ctx context.Context, req *pb.Commitment) (*pb.Value, error) {
	rand.Seed(time.Now().UnixNano())
	return &pb.Value{V: int32(rand.Intn(5) + 1)}, nil
}

func main() {
	rand.Seed(time.Now().UnixNano())

	opts := grpc.WithInsecure()
	con, err := grpc.Dial("localhost:8080", opts)
	if err != nil {
		log.Fatalf("Error connecting: %v \n", err)
	}

	defer con.Close()
	c := pb.NewDiceGameClient(con)

	send(c)

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterDiceGameServer(s, &server{})
	log.Printf("Alice listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func send(c pb.DiceGameClient) {
	roll := int32(rand.Intn(5) + 1)
	salt := int32(rand.Intn(999999))
	commitment := sha256.Sum256([]byte(string(salt + roll)))
	res, err := c.Initiate(context.Background(), &pb.Commitment{Commitment: commitment[:]})
	if err != nil {
		log.Fatalf("Error connecting: %v \n", err)
	}

	res2, err := c.Confirmation(context.Background(), &pb.Secrets{R: salt, V: roll})

	if !res2.Accepted || err != nil {
		log.Fatalln("Bob did not accept the result 😠")
	}

	finalRoll := ((res.V + roll) % 6) + 1

	log.Printf("Final roll: %d\n", finalRoll)
}
