package main

import (
	"context"
	firebase "firebase.google.com/go"
	"fmt"
	"strconv"

	"firebase.google.com/go/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
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

var client *db.Client

func main() {
	clientEF, err := ethclient.Dial("https://mainnet.infura.io/v3/160b897ccac244c888cad39a5529303d")
	if err != nil {
		log.Fatalln(err)
	}
	var array []*big.Int
	index := 0
	for {
		header, err := clientEF.HeaderByNumber(context.Background(), nil)
		if err != nil {
			log.Fatal(err)
		}
		headerNumber := header.Number
		if !isElementInArray(headerNumber, array) {
			array = append(array, headerNumber)
			fmt.Println(headerNumber)
			block, err := clientEF.BlockByNumber(context.Background(), headerNumber)
			if err != nil {
				continue
			}
			blockDB := getBlockDB(block)
			addBlockToDB(blockDB, index)
			index++
		}
	}
}

func isElementInArray(element *big.Int, array []*big.Int) bool {
	for _, elem := range array {
		if elem.Cmp(element) == 0 {
			return true
		}
	}
	return false
}

func getBlockDB(block *types.Block) BlockDB {
	blockDB := BlockDB{
		block.Number().Uint64(),
		block.Time(),
		block.Difficulty().Uint64(),
		block.Hash().Hex(),
		getTransactionsDB(block.Transactions()),
	}
	return blockDB
}

func getTransactionsDB(transactions types.Transactions) []TransactionDB {
	var transactionsDB []TransactionDB
	for _, transaction := range transactions {
		transactionDB := TransactionDB{
			transaction.ChainId().String(),
			transaction.Hash(),
			transaction.Value().String(),
			transaction.Cost().String(),
			transaction.To(),
			transaction.Gas(),
			transaction.GasPrice().String(),
		}
		transactionsDB = append(transactionsDB, transactionDB)
	}

	return transactionsDB
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

func addBlockToDB(blockDB BlockDB, index int) {

	// create ref at path user_scores/:userId
	ref := client.NewRef("eth/blocks/" + strconv.Itoa(index))
	if err := ref.Set(context.TODO(), blockDB); err != nil {
		log.Fatal(err)
	}
	fmt.Println("score added/updated successfully!")
}
