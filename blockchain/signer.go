package blockchain

import(
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
)

type Signer interface {
	GetAddress() ethereum.Address
	GetTransactOpts() *bind.TransactOpts
	Sign(tx *types.Transaction) (*types.Transaction, error)
}
