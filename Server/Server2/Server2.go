package main

import (
	proto "Replication/grpc"
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var serverLogicalTime int64 = 0
var highestBid int64 = 0
var ls int64 = 0
var auctionGoing bool
var TimeLeftOfAuction int64 = 0
var leadingClientId int64 = 0

type Auction struct {
	proto.UnimplementedAuctionServer
	bids []int64
}

func NewServer() *Auction { return &Auction{} }

func main() {
	// Set up gRPC server
	grpcServer := grpc.NewServer()

	svc := NewServer()
	proto.RegisterAuctionServer(grpcServer, svc)
	listener, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		log.Fatalf("Lorte program det virker ikke", err)
	}

	go func() {
		grpcServer.Serve(listener)
	}()
	// Set up connection to Server 1
	conn, err := grpc.NewClient(":50050", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("kan ikke oprette mig som client")
	}

	auctionClient := proto.NewAuctionClient(conn)
	go func() {
		for {
			// check if primary server is alive every 10 seconds by sending heartbeat and waiting for response
			fmt.Println("Sending a HeartBeat to server 1")
			resp, err := auctionClient.HeartBeat(context.Background(), &proto.Empty{})
			if err != nil {
				// activate plan B if no response from primary
				fmt.Println("CHAOS ALARM")
				fmt.Println(err, resp)
				PLANB(grpcServer)
				break
			}
			time.Sleep(10 * time.Second)
		}
	}()

	for {
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		txt := strings.TrimSpace(line)
		if txt == "shutdown" {
			grpcServer.Stop()
			fmt.Println("Server", "Crashed")
			os.Exit(0)
		}
		if txt == "start" {
			auctionGoing = true
			fmt.Println("The auction has started")
			TimeLeftOfAuction := 100
			go func() {
				for {
					if TimeLeftOfAuction%10 == 0 {
						fmt.Println("There is", TimeLeftOfAuction, " seconds left of the auction")
						fmt.Println("The highest bid is", highestBid)
					}

					if TimeLeftOfAuction == 0 {
						fmt.Println("The auction has finished")
						fmt.Println("The highest bid is", highestBid)
						fmt.Println("The client winner is: ", leadingClientId)
						auctionGoing = false
						break
					}
					TimeLeftOfAuction = TimeLeftOfAuction - 1
					time.Sleep(time.Second)
				}
			}()
		}
	}
}

// recieve bid update logical time and highest bid
func (a *Auction) Bid(ctx context.Context, BidIn *proto.BidIn) (*proto.BidAck, error) {
	if ls < BidIn.Ls { // && id < BidIn.Nid))
		ls = BidIn.Ls + 1
	}
	if BidIn.Bid > highestBid {
		highestBid = BidIn.Bid
		leadingClientId = BidIn.ClientId

		return &proto.BidAck{
			Ack: "Success",
			Ls:  serverLogicalTime,
		}, nil
	}

	return &proto.BidAck{
		Ack: "Your bid is lower than the highest bid",
		Ls:  serverLogicalTime,
	}, nil

}

// Redundancy plan when primary server fails
// start server as primary if auction is still ongoing(has time left)
func PLANB(grpcServer *grpc.Server) {
	listener, err := net.Listen("tcp", "localhost:50050")
	if err != nil {
		log.Fatalf("Lorte program det virker ikke", err)
	}
	go grpcServer.Serve(listener)
	if auctionGoing {
		for {
			if TimeLeftOfAuction%10 == 0 {
				fmt.Println("There is", TimeLeftOfAuction, " seconds left of the auction")
				fmt.Println("The highest bid is", highestBid)
			}

			if TimeLeftOfAuction == 0 {
				fmt.Println("The auction has finished")
				fmt.Println("The highest bid is", highestBid)
				fmt.Println("The winner is: ", leadingClientId)
				auctionGoing = false
				break
			}
			TimeLeftOfAuction = TimeLeftOfAuction - 1
			time.Sleep(time.Second)
		}
	}
}

// receive backup from primary
func (a *Auction) BackUpToReplicas(ctx context.Context, data *proto.DataBackup) (*proto.Empty, error) {
	highestBid = data.HighestBid
	ls = data.Ls
	auctionGoing = data.AuctionGoing
	TimeLeftOfAuction = data.TimeLeftOfAuction
	return &proto.Empty{}, nil
}

func (a *Auction) Result(ctx context.Context, x *proto.Empty) (*proto.ResultOut, error) {
	return &proto.ResultOut{
		HighestBid: highestBid,
	}, nil
}
