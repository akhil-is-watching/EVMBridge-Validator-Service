package service

import (
	"context"
	"log"
	"math/big"
	"math/rand"
	"time"

	"github.com/akhil-is-watching/E2E-Bridge/Proposal-Validator/bridge"
	"github.com/akhil-is-watching/E2E-Bridge/Proposal-Validator/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type ValidatorService struct {
	listenClient  *ethclient.Client
	listenChainID *big.Int
	listenAddress common.Address
	listenBridge  *bridge.Bridge
	validator     *types.Validator
	cache         *types.BoundedCache
}

func NewValidatorService(
	listenClient *ethclient.Client,
	listenChainIDInt int64,
	listenAddressHex string,
	validator *types.Validator,
) *ValidatorService {

	listenChainID := big.NewInt(listenChainIDInt)
	listenAddress := common.HexToAddress(listenAddressHex)
	listenBridge, err := bridge.NewBridge(listenAddress, listenClient)
	if err != nil {
		log.Fatal(err)
	}

	return &ValidatorService{
		listenClient:  listenClient,
		listenChainID: listenChainID,
		listenAddress: listenAddress,
		listenBridge:  listenBridge,
		validator:     validator,
		cache:         types.NewBoundedCache(15),
	}
}

func (service *ValidatorService) ListenForProposal(startBlock uint64) {

	for {

		latestBlock, err := service.listenClient.BlockNumber(context.Background())
		if err != nil {
			log.Println("Error getting latest block:", err)
			continue
		}

		if startBlock == 0 {
			startBlock = latestBlock
		}

		for blockNumber := startBlock; blockNumber <= latestBlock; {
			endBlock := blockNumber + 5
			if endBlock > latestBlock {
				endBlock = latestBlock
			}

			it, err := service.listenBridge.FilterProposalEvent(&bind.FilterOpts{
				Start: blockNumber,
				End:   &endBlock,
			})

			if err != nil {
				log.Println("Error getting latest block:", err)
				continue
			}

			for it.Next() {
				event := it.Event
				if service.cache.Contains(event.Raw.TxHash) {
					continue // Skip processing if the event has been processed
				}

				log.Println("[FANTOM] PROPOSAL EVENT FOUND")
				log.Println("[FANTOM] Block:", it.Event.Raw.BlockNumber)
				log.Println("[FANTOM] TransactionHash:", it.Event.Raw.TxHash.Hex())
				log.Println("[SERVICE] INITIATING VOTING TRANSACTION")
				randomSleep := rand.Intn(5)
				time.Sleep(time.Duration(randomSleep))
				go service.validator.VoteProposal(event.ProposalID)
				service.cache.Add(event.Raw.TxHash)
			}
			blockNumber = endBlock + 1 // Move to the next block after the last processed block
		}

		startBlock = latestBlock
		time.Sleep(5 * time.Second)
	}

}

// func (service *ValidatorService) ListenForProposal(startBlock uint64) {

// 	for {

// 		latestBlock, err := service.listenClient.BlockNumber(context.Background())
// 		if err != nil {
// 			log.Println("Error getting latest block:", err)
// 			continue
// 		}

// 		if startBlock == 0 {
// 			startBlock = latestBlock
// 		}

// 		for blockNumber := startBlock; blockNumber <= latestBlock; blockNumber++ {

// 			it, _ := service.listenBridge.FilterProposalEvent(&bind.FilterOpts{
// 				Start: blockNumber,
// 				End:   &blockNumber,
// 			})

// 			for it.Next() {
// 				event := it.Event
// 				log.Println("[FANTOM] PROPOSAL EVENT FOUND")
// 				log.Println("[FANTOM] Block:", it.Event.Raw.BlockNumber)
// 				log.Println("[FANTOM] TransactionHash:", it.Event.Raw.TxHash.Hex())
// 				log.Println("[SERVICE] INITIATING VOTING TRANSACTION")
// 				randomSleep := rand.Intn(5)
// 				time.Sleep(time.Duration(randomSleep))
// 				go service.validator.VoteProposal(event.ProposalID)
// 			}
// 			time.Sleep(3 * time.Second)
// 		}
// 		startBlock = latestBlock
// 		time.Sleep(5 * time.Second)
// 	}

// }
