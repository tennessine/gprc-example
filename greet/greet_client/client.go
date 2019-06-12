package main

import (
	"awesomeProject/greet/greetpb"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"time"
)

func main() {
	fmt.Println("Hello I'm a client")

	certFile := "ssl/ca.crt"
	creds, err := credentials.NewClientTLSFromFile(certFile, "")
	if err != nil {
		log.Fatalf("Error while loading CA trust certificate: %v", err)
		return
	}

	cc, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer cc.Close()

	c := greetpb.NewGreetServiceClient(cc)
	//fmt.Printf("Created client: %f", c)

	doUnary(c)
	//doServerStreaming(c)
	//doBiDiStreaming(c)
	//doUnaryWithDeadline(c, 5*time.Second) // should complete
	//doUnaryWithDeadline(c, time.Second) // should timeout
}

func doUnary(c greetpb.GreetServiceClient) {
	fmt.Printf("Starting to do a unary RPC..\n")
	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Kongfei",
			LastName:  "Ge",
		},
	}
	resp, err := c.Greet(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling Greet RPC: %v", err)
	}
	log.Printf("Response from Greet: %v\n", resp.Result)
}

func doServerStreaming(c greetpb.GreetServiceClient) {
	fmt.Printf("Starting to do a ServerStreaming RPC..\n")
	req := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Kongfei",
			LastName:  "Ge",
		},
	}
	rs, err := c.GreetManyTimes(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling GreetManyTimes RPC: %v", err)
	}
	for {
		msg, err := rs.Recv()
		if err == io.EOF {
			// we've reached the end of the stream
			break
		}
		if err != nil {
			log.Fatalf("Error while reading stream: %v", err)
		}
		log.Printf("Response from GreetManyTimes: %v\n", msg.GetResult())
	}
}

func doBiDiStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting to do a BiDi Streaming RPC...")
	stream, err := c.GreetEveryone(context.Background())
	if err != nil {
		log.Fatalf("Error while creating stream: %v", err)
		return
	}
	requests := []*greetpb.GreetEveryoneRequest{
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "one",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "two",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "three",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "four",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "five",
			},
		},
	}
	waitc := make(chan struct{})
	go func() {
		for _, req := range requests {
			fmt.Printf("Sending message: %v\n", req)
			stream.Send(req)
			time.Sleep(time.Second)
		}
		stream.CloseSend()
	}()
	go func() {
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error while receiving: %v", err)
				break
			}
			fmt.Printf("Received: %v\n", msg.GetResult())
		}
		close(waitc)
	}()
	<-waitc
}

func doUnaryWithDeadline(c greetpb.GreetServiceClient, timeout time.Duration) {
	fmt.Printf("Starting to do a UnaryWithDeadline RPC..\n")
	req := &greetpb.GreetWithDeadlineRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Kongfei",
			LastName:  "Ge",
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := c.GreetWithDeadline(ctx, req)
	if err != nil {
		s, ok := status.FromError(err)
		if ok {
			if s.Code() == codes.DeadlineExceeded {
				fmt.Println("Deadline was exceeded")
			} else {
				fmt.Printf("Unexpected error: %v\n", s)
			}
		} else {
			log.Fatalf("error while calling Greet RPC: %v", err)
		}
		return
	}
	log.Printf("Response from Greet: %v\n", resp.Result)
}
