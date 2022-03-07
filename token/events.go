package token

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type LogTransfer struct {
	From   common.Address
	To     common.Address
	Tokens *big.Int
}

type LogApproval struct {
	TokenOwner common.Address
	Spender    common.Address
	Tokens     *big.Int
}
