package privatekey

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
)

const to = "0xA4e25D42D9a8E83cfa60A7FE2e2345Ba4b5C8F0d"

func TestCheckAndTakeLoop(t *testing.T) {
	var i int64
	for i = 1000; i < 1100; i++ {
		// 1. select 1 ~ 2^256
		num := big.NewInt(i)
		t.Log("length", len(common.LeftPadBytes(num.Bytes()[:], 32)))

		checkAndTake(t, num)
	}
}

func TestCheckAndTake(t *testing.T) {
	err := checkAndTake(t, big.NewInt(999999))
	require.NoError(t, err)
}

func checkAndTake(t *testing.T, num *big.Int) error {
	ctx := context.Background()

	// 2. gen private key
	key, err := crypto.ToECDSA(common.LeftPadBytes(num.Bytes()[:], 32))
	require.NoError(t, err)

	// get pub key and address
	addr := crypto.PubkeyToAddress(key.PublicKey)
	t.Log("address", addr)

	client, err := ethclient.DialContext(context.Background(), "https://cloudflare-eth.com")
	require.NoError(t, err)
	gasPrice, err := client.SuggestGasPrice(ctx)
	require.NoError(t, err)
	nonceAt, err := client.NonceAt(ctx, addr, nil)
	require.NoError(t, err)
	balance, err := client.BalanceAt(context.Background(), addr, nil)
	require.NoError(t, err)
	chain, err := client.ChainID(ctx)
	require.NoError(t, err)
	require.NoError(t, err)
	t.Log("balanceOf", balance.Int64())

	toAddr := common.HexToAddress(to)
	unsignedTx := types.NewTransaction(nonceAt, toAddr, big.NewInt(100), params.TxGas, gasPrice, []byte{})
	signedTx, err := types.SignTx(unsignedTx, types.LatestSignerForChainID(chain), key)
	require.NoError(t, err)

	if err = client.SendTransaction(ctx, signedTx); err != nil {
		t.Log("fail", num.Uint64())
		return err
	} else {
		t.Log("success", num.Uint64())
		return err
	}

}
