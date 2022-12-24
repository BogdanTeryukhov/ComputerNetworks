package main

import (
	"context"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/db"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type TransactionDB struct {
	ChainId  string
	Hash     common.Hash
	Value    string
	Cost     string
	To       *common.Address
	Gas      uint64
	GasPrice string
}

type BlockDB struct {
	Number       uint64
	Time         uint64
	Difficulty   uint64
	Hash         string
	Transactions []TransactionDB
}

type Blocks struct {
	Blocks []BlockDB `json:"blocks"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var client *db.Client

func main() {
	server := gin.New()

	server.GET("/", UsersOnlineHandler)

	err := server.Run(":8091")
	if err != nil {
		log.Fatalln(err)
	}
}
func UsersOnlineHandler(ctx *gin.Context) {
	ws, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ws.Close()
	for {
		blocks := getBlocks()

		//log.Println(blocks)
		err = ws.WriteJSON(blocks)
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}

func init() {
	ctx := context.Background()
	conf := &firebase.Config{
		DatabaseURL: "https://bogdan-13765-default-rtdb.firebaseio.com/",
	}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Fatalf("firebase.NewApp: %v", err)
	}
	client, err = app.Database(ctx)
	if err != nil {
		log.Fatalf("app.Firestore: %v", err)
	}
}

func getBlocks() []BlockDB {
	ref := client.NewRef("eth/blocks")
	var blockDB []BlockDB
	if err := ref.Get(context.TODO(), &blockDB); err != nil {
		log.Println(err)
	}
	return blockDB
}
