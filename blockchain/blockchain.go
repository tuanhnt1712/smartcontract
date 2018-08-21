package blockchain

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"time"

	ether "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type Blockchain struct {
	client    *rpc.Client
	ethclient *ethclient.Client
	signer    Signer
	abi       abi.ABI
	contract  ethereum.Address
}

func NewBlockchain(client *rpc.Client,
	ethclient *ethclient.Client,
	signer Signer, contract, abiPath string) *Blockchain {

	_, fileLocation, _, _ := runtime.Caller(1)
	abiPath = filepath.Join(fileLocation, abiPath)
	file, err := os.Open(abiPath)
	if err != nil {
		panic(err)
	}
	parsed, err := abi.JSON(file)
	if err != nil {
		panic(err)
	}

	return &Blockchain{
		client:    client,
		ethclient: ethclient,
		signer:    signer,
		abi:       parsed,
		contract:  ethereum.HexToAddress(contract),
	}
}

func (self *Blockchain) GetTreasure() {
	input, err := self.abi.Pack("getTreasure", self.signer.GetAddress())
	if err != nil {
		log.Println(err)
	}
	value := big.NewInt(0)
	from := self.signer.GetAddress()
	msg := ether.CallMsg{From: from, To: &self.contract, Value: value, Data: input}
	result, err := self.ethclient.CallContract(context.Background(), msg, nil)
	if err != nil {
		log.Println(err)
	}
	// log.Println("treasure: ", string(result))
	log.Println("treasure: ", result)

	var trueResult interface{}
	err = self.abi.Unpack(&trueResult, "getTreasure", result)
	if err != nil {
		log.Println(err)
	}

	log.Println("result get treasure: ", trueResult)
}

func (self *Blockchain) AddTreasure(owner string, amount *big.Int) (*types.Transaction, error) {

	opts, cancel, err := self.getTransactOpts(nil, nil)
	defer cancel()

	if err != nil {
		return nil, err
	} else {
		tx, err := self.buildTx(
			opts,
			"addTreasure",
			self.signer.GetAddress(),
			amount)
		if err != nil {
			return nil, err
		} else {
			// fmt.Println("raw tx: ", tx)
			return self.signAndBroadcast(tx, self.signer)
		}
	}
	return nil, errors.New("add done")
}

func (self *Blockchain) signAndBroadcast(tx *types.Transaction, singer Signer) (*types.Transaction, error) {
	if tx == nil {
		panic(errors.New("Nil tx is forbidden here"))
	} else {
		signedTx, err := singer.Sign(tx)
		if err != nil {
			return nil, err
		}
		// log.Println("raw tx: ", signedTx)
		ctx := context.Background()
		err = self.ethclient.SendTransaction(ctx, signedTx)
		if err != nil {
			log.Println("failed to broadcast tx: ", err)
		}
		log.Println("send done!")
		return signedTx, nil
	}
}

func (self *Blockchain) buildTx(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	input, err := self.abi.Pack(method, params...)
	if err != nil {
		return nil, err
	}
	log.Println("build tx: ", opts.From.Hex())
	return self.transactTx(opts, &self.contract, input)
}

func ensureContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.TODO()
	}
	return ctx
}

func (self *Blockchain) transactTx(opts *bind.TransactOpts, contract *ethereum.Address, input []byte) (*types.Transaction, error) {
	var err error
	// Ensure a valid value field and resolve the account nonce
	value := opts.Value
	if value == nil {
		value = new(big.Int)
	}
	var nonce uint64
	if opts.Nonce == nil {
		return nil, errors.New("nonce must be specified")
	} else {
		nonce = opts.Nonce.Uint64()
	}
	// Figure out the gas allowance and gas price values
	gasPrice := opts.GasPrice
	if gasPrice == nil {
		return nil, errors.New("gas price must be specified")
	}
	gasLimit := opts.GasLimit
	if gasLimit == 0 {
		// Gas estimation cannot succeed without code for method invocations
		if contract != nil {
			if code, err := self.ethclient.PendingCodeAt(ensureContext(opts.Context), self.contract); err != nil {
				return nil, err
			} else if len(code) == 0 {
				return nil, bind.ErrNoCode
			}
		}
		// If the contract surely has code (or code is not needed), estimate the transaction
		msg := ether.CallMsg{From: opts.From, To: contract, Value: value, Data: input}
		gasLimit, err = self.ethclient.EstimateGas(ensureContext(opts.Context), msg)
		if err != nil {
			return nil, fmt.Errorf("failed to estimate gas needed: %v", err)
		}
		// add gas limit by 50K gas
		// gasLimit.Add(gasLimit, big.NewInt(50000))
		gasLimit = gasLimit + 50000
	}
	// Create the transaction, sign it and schedule it for execution
	var rawTx *types.Transaction
	if contract == nil {
		rawTx = types.NewContractCreation(nonce, value, gasLimit, gasPrice, input)
	} else {
		rawTx = types.NewTransaction(nonce, self.contract, value, gasLimit, gasPrice, input)
	}
	return rawTx, nil
}

func donothing() {}

func (self *Blockchain) getTransactOpts(nonce *big.Int, gasPrice *big.Int) (*bind.TransactOpts, context.CancelFunc, error) {
	shared := self.signer.GetTransactOpts()
	var err error
	if nonce == nil {
		nonce, err = self.GetNonceFromNode()
	}
	if err != nil {
		return nil, donothing, err
	}
	if gasPrice == nil {
		gasPrice = big.NewInt(50100000000)
	}
	timeout, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	result := bind.TransactOpts{
		shared.From,
		nonce,
		shared.Signer,
		shared.Value,
		gasPrice,
		shared.GasLimit,
		timeout,
	}
	return &result, cancel, nil
}

func (self *Blockchain) GetNonceFromNode() (*big.Int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	nonce, err := self.ethclient.PendingNonceAt(ctx, self.signer.GetAddress())
	return big.NewInt(int64(nonce)), err
}
