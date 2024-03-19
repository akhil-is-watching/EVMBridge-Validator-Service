package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/akhil-is-watching/E2E-Bridge/Proposal-Validator/service"
	"github.com/akhil-is-watching/E2E-Bridge/Proposal-Validator/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {

	listenRpcUrl := flag.String("rpc", "", "")
	listenChainID := flag.Int64("chain", 0, "")
	listenContract := flag.String("address", "", "")
	privateKey := flag.String("privateKey", "", "")
	startBlock := flag.Int64("block", 0, "")

	flag.Parse()
	if *listenRpcUrl == "" {
		log.Fatalf("Missing rpc")
		os.Exit(0)
	}

	if *listenChainID == 0 {
		log.Fatalf("Missing chain")
		os.Exit(0)
	}

	if *listenContract == "" {
		log.Fatalf("Missing address")
		os.Exit(0)
	}

	if *privateKey == "" {
		log.Fatalf("Missing privateKey")
		os.Exit(0)
	}

	listenClient, _ := ethclient.DialContext(context.Background(), *listenRpcUrl)

	validator := types.NewValidator(
		*privateKey,
		*listenContract,
		*listenChainID,
		listenClient,
	)

	proposal_service := service.NewValidatorService(
		listenClient,
		*listenChainID,
		*listenContract,
		validator,
	)

	if *startBlock != 0 {
		proposal_service.ListenForProposal(uint64(*startBlock))
	} else {
		proposal_service.ListenForProposal(0)
	}
}
