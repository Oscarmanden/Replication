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
var leadingClientId int64 = 0
var auctionGoing = false
var auctionClient proto.AuctionClient
var TimeLeftOfAuction int64 = 30

type Auction struct {
	proto.UnimplementedAuctionServer
}

func NewServer() *Auction {
	return &Auction{}
}

func main() {
	// Set up gRPC server
	listener, err := net.Listen("tcp", "localhost:50050")
	if err != nil {
		log.Fatalf("Lorte program det virker ikke", err)
	}
	grpcServer := grpc.NewServer()
	svc := NewServer()
	// Set up connection to Server 2
	conn, err := grpc.NewClient(":50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Ingen forbindelse til Server 1")
	}
	auctionClient = proto.NewAuctionClient(conn)

	proto.RegisterAuctionServer(grpcServer, svc)
	go func() {
		grpcServer.Serve(listener)
	}()
	for {
		// command line input
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		txt := strings.TrimSpace(line)
		if txt == "shutdown" {
			grpcServer.Stop()
			fmt.Println("Server", "Crashed")
			os.Exit(0)
		}
		if txt == "start" {
			highestBid = 0
			auctionGoing = true
			fmt.Println("The auction has started")
			// timer for auction
			go func() {
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
			}()
		}
	}
}

// recieve bid
func (a *Auction) Bid(ctx context.Context, BidIn *proto.BidIn) (*proto.BidAck, error) {
	if serverLogicalTime < BidIn.Ls { // && id < BidIn.Nid))
		serverLogicalTime = BidIn.Ls + 1
	}
	fmt.Println("Received Big Beautiful Bid: ", BidIn.Bid, "$$$")
	fmt.Println("From client_id: ", BidIn.ClientId)

	if BidIn.Bid > highestBid {
		highestBid = BidIn.Bid
		leadingClientId = BidIn.ClientId
		// everytime a new winner appears, we update the backup replicas
		sendDataBackup()
		return &proto.BidAck{
			Ack: "Success",
			Ls:  serverLogicalTime,
		}, nil
	} else {
		return &proto.BidAck{
			Ack: fmt.Sprintf("Your bid is lower than the highest bid, highest bid: %d\n", highestBid),
			Ls:  serverLogicalTime,
		}, nil
	}
}

func (a *Auction) Result(ctx context.Context, x *proto.Empty) (*proto.ResultOut, error) {
	return &proto.ResultOut{
		HighestBid: highestBid,
	}, nil
}

func (a *Auction) HeartBeat(ctx context.Context, x *proto.Empty) (*proto.ImAlive, error) {
	return &proto.ImAlive{
		Ack: true,
	}, nil

}

// updates backup replicas with current data
func sendDataBackup() {
	ctx := context.Background()
	data := &proto.DataBackup{
		HighestBid:        highestBid,
		Ls:                serverLogicalTime,
		AuctionGoing:      auctionGoing,
		TimeLeftOfAuction: TimeLeftOfAuction,
	}
	_, err := auctionClient.BackUpToReplicas(ctx, data)
	if err != nil {
		log.Fatalf("shit & piss backup  %v", err)
	}

}
