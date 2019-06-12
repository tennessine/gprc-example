package main

import (
	"awesomeProject/calculator/calculatorpb"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
)

func main() {
	fmt.Println("Calculator Client")
	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer cc.Close()

	c := calculatorpb.NewCalculatorServiceClient(cc)
	//fmt.Printf("Created client: %f", c)

	//doUnary(c)
	//doServerStreaming(c)
	doErrorUnary(c)
}

func doErrorUnary(c calculatorpb.CalculatorServiceClient) {
	fmt.Printf("Starting to do a SquareRoot unary RPC..\n")
	// correct call
	doErrorCall(c, 10)
	// error call
	doErrorCall(c, -1)
}

func doErrorCall(c calculatorpb.CalculatorServiceClient, n int32) {
	resp, err := c.SquareRoot(context.Background(), &calculatorpb.SquareRootRequest{
		Number: n,
	})
	if err != nil {
		s, ok := status.FromError(err)
		if ok {
			// actual error from gRPC (user error)
			fmt.Printf("Error message from server: %v\n", s.Message())
			if s.Code() == codes.InvalidArgument {
				fmt.Println("We probably sent a negative number")
				return
			}
		} else {
			log.Fatalf("Big error calling SquareRoot: %v", err)
			return
		}
	}
	fmt.Printf("Result of square root of %v: %v\n", n, resp.GetNumberRoot())
}

func doUnary(c calculatorpb.CalculatorServiceClient) {
	fmt.Printf("Starting to do a Sum unary RPC..\n")
	req := &calculatorpb.SumRequest{
		FirstNumber: 1,
		SecondNumber: 2,
	}
	resp, err := c.Sum(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling Calculator RPC: %v", err)
	}
	log.Printf("Response from Calculator: %v\n", resp.SumResult)
}

func doServerStreaming(c calculatorpb.CalculatorServiceClient) {
	fmt.Printf("Starting to do a PrimeDecomposition Server Streaming RPC..\n")
	req := &calculatorpb.PrimeNumberDecompositionRequest{
		Number: 1239039284099999000,
	}
	stream, err := c.PrimeNumberDecomposition(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling PrimeDecomposition RPC: %v", err)
	}
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Something happened: %v", err)
		}
		fmt.Println(resp.GetPrimeFactor())
	}
}