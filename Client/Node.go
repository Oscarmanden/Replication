package main

import (
	proto "Replication/grpc"
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var ls int64 = 0
var client_id int64

func main() {
	conn, err := grpc.NewClient(":50050", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Not working")
	}

	client := proto.NewAuctionClient(conn)
	ctx := context.Background()
	var Value int64 = 0
	// switch for random values or terminal input just change true to false
	var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		switch text {
		case "bid":
			Value = rng.Int63n(100000)
		case "result":
			resp, _ := client.Result(ctx, &proto.Empty{})
			fmt.Println("the highest bid is", resp.GetHighestBid())
			continue
		default:
			//
			Value, err = strconv.ParseInt(text, 10, 64)
			if err != nil {
				fmt.Fprintln(os.Stderr, "invalid integer:", err)
				// :^)
			}
		}
		fmt.Printf("Bid: %d\n", Value)

		// sending bid
		req := &proto.BidIn{Bid: Value, Ls: ls, ClientId: client_id}

		//Exception skal håndteres, eller vent på ACK
		resp, err := client.Bid(ctx, req)
		if err != nil {
			fmt.Println("Server error, please wait before retrying")
			time.Sleep(5 * time.Second)
			continue
		}
		fmt.Println(resp.Ack)

	}
}
