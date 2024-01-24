# Ethereum Transactions and NFT Transfers Crawler

This project is a simple API that allows users to view Ethereum blockchain transaction data and NFT transfer events associated with a specific wallet address.

## Installation

1. Install Go: https://golang.org/doc/install
2. Clone this repository: `git clone https://github.com/yourusername/yourrepository.git`
3. Navigate to the project directory: `cd yourrepository`
4. Install the dependencies: `go get`

## Usage

1. Start the server: `go run main.go`
2. Open a web browser and navigate to `http://localhost:8080/transactions?address=youraddress&startBlock=yourstartblock` to view transaction data, or `http://localhost:8080/nft-transfers?address=youraddress&contractAddress=yourcontractaddress&startBlock=yourstartblock` to view NFT transfer events.