package types

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"sync"

	"github.com/akhil-is-watching/E2E-Bridge/Proposal-Validator/bridge"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Validator struct {
	mu            sync.Mutex
	client        *ethclient.Client
	address       string
	listenAddress common.Address
	listenChainID *big.Int
	privateKey    *ecdsa.PrivateKey
	nonce         uint64
	nonceMu       sync.Mutex // Mutex for nonce management
}

func NewValidator(
	privateKeyHex string,
	listenAddressHex string,
	listenChainIDInt int64,
	client *ethclient.Client,
) *Validator {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatal("PRIVATE KEY DERIVATION FAILED")
	}

	listenAddress := common.HexToAddress(listenAddressHex)
	listenChainID := big.NewInt(listenChainIDInt)

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("PUBLIC KEY DERIVATION FAILED")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, _ := client.PendingNonceAt(context.Background(), address)

	return &Validator{
		client:        client,
		listenAddress: listenAddress,
		listenChainID: listenChainID,
		address:       address.Hex(),
		privateKey:    privateKey,
		nonce:         nonce,
	}
}

func (p *Validator) Address() string {
	return p.address
}

func (p *Validator) GetNonce() uint64 {
	p.nonceMu.Lock()
	defer p.nonceMu.Unlock()
	return p.nonce
}

func (p *Validator) IncrementNonce() {
	p.nonceMu.Lock()
	defer p.nonceMu.Unlock()
	p.nonce++
}

func (p *Validator) VoteProposal(proposalId [32]byte) {
	p.mu.Lock()
	defer p.mu.Unlock()

	gasPrice, err := p.client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
		return
	}

	auth, _ := bind.NewKeyedTransactorWithChainID(p.privateKey, p.listenChainID)
	auth.Nonce = big.NewInt(int64(p.GetNonce()))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(300000)
	auth.GasPrice = gasPrice

	instance, err := bridge.NewBridge(p.listenAddress, p.client)
	if err != nil {
		log.Fatal(err)
		return
	}

	tx, err := instance.Vote(
		auth,
		proposalId,
	)

	if err != nil {
		log.Fatal(err)
		return
	}
	p.IncrementNonce()
	fmt.Println("[FANTOM]TransactionHash: ", tx.Hash().Hex())
	fmt.Println(" ")
}
