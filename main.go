package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/sha3"
)

type TransactionData struct {
	TxHash  string
	Address string
	Data    string
}

type NFTTransferEvent struct {
	From    string
	To      string
	TokenId string
}

func getTransactionData(client *ethclient.Client, address common.Address, startBlock *big.Int) ([]TransactionData, error) {
	query := ethereum.FilterQuery{
		FromBlock: startBlock,
		ToBlock:   nil,
		Addresses: []common.Address{
			address,
		},
	}

	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		return nil, err
	}

	var transactions []TransactionData
	for _, vLog := range logs {
		transactions = append(transactions, TransactionData{
			TxHash:  vLog.TxHash.Hex(),
			Address: vLog.Address.Hex(),
			Data:    string(vLog.Data),
		})
	}

	return transactions, nil
}

func getNFTTransferEvents(client *ethclient.Client, contractAddress common.Address, userAddress common.Address, startBlock *big.Int) ([]NFTTransferEvent, error) {
	query := ethereum.FilterQuery{
		FromBlock: startBlock,
		Addresses: []common.Address{
			contractAddress,
		},
	}

	transferFnSignature := []byte("Transfer(address,address,uint256)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	hashed := hash.Sum(nil)
	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		return nil, err
	}

	var events []NFTTransferEvent
	for _, vLog := range logs {
		if vLog.Topics[0].Hex() == common.Bytes2Hex(hashed) && (vLog.Topics[1].Hex() == userAddress.Hex() || vLog.Topics[2].Hex() == userAddress.Hex()) {
			var transferEvent struct {
				From    common.Address
				To      common.Address
				TokenId *big.Int
			}

			// Define the ABI of the contract
			const contractABI = `[{"constant":true,"inputs":[{"name":"tokenId","type":"uint256"}],"name":"ownerOf","outputs":[{"name":"owner","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":true,"name":"tokenId","type":"uint256"}],"name":"Transfer","type":"event"}]`

			parsedABI, err := abi.JSON(strings.NewReader(contractABI))
			if err != nil {
				return nil, err
			}

			err = parsedABI.UnpackIntoInterface(&transferEvent, "Transfer", vLog.Data)
			if err != nil {
				return nil, err
			}

			events = append(events, NFTTransferEvent{
				From:    transferEvent.From.Hex(),
				To:      transferEvent.To.Hex(),
				TokenId: transferEvent.TokenId.String(),
			})
		}
	}

	return events, nil
}

func getBalanceAtDate(client *ethclient.Client, address common.Address, date time.Time) (*big.Int, error) {
	blockNumber, err := client.BlockNumber(context.Background())
	if err != nil {
		return nil, err
	}

	for i := blockNumber; i > 0; i-- {
		block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(i)))
		if err != nil {
			return nil, err
		}

		if time.Unix(int64(block.Time()), 0).Before(date) {
			return client.BalanceAt(context.Background(), address, big.NewInt(int64(i)))
		}
	}

	return nil, fmt.Errorf("no blocks found before the specified date")
}

func main() {
	client, err := ethclient.Dial("https://mainnet.infura.io")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	router := gin.Default()

	router.GET("/balance", func(c *gin.Context) {
		address := common.HexToAddress(c.Query("address"))
		date, _ := time.Parse("2006-01-02", c.Query("date"))

		balance, err := getBalanceAtDate(client, address, date)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"balance": balance.String()})
	})

	router.GET("/transactions", func(c *gin.Context) {
		address := common.HexToAddress(c.Query("address"))
		startBlock, _ := new(big.Int).SetString(c.Query("startBlock"), 10)
		transactions, err := getTransactionData(client, address, startBlock)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"transactions": transactions})
	})

	router.GET("/nft-transfers", func(c *gin.Context) {
		address := common.HexToAddress(c.Query("address"))
		contractAddress := common.HexToAddress(c.Query("contractAddress"))
		startBlock, _ := new(big.Int).SetString(c.Query("startBlock"), 10)
		events, err := getNFTTransferEvents(client, contractAddress, address, startBlock)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"nft_transfers": events})
	})

	router.Run(":8080")
}
