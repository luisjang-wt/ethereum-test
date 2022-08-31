package privatekey

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

const (
	to  = "0xA4e25D42D9a8E83cfa60A7FE2e2345Ba4b5C8F0d"
	cmp = 117133190704000
)

var (
	client *ethclient.Client
	urls   = []string{
		"https://cloudflare-eth.com",
		//"https://http-pwemix.phnxops.in",
	}
)

func getHost() string {
	//return urls[rand.Int31n(2)]
	return urls[0]
}

func TestCheckAndTakeLoop(t *testing.T) {
	var (
		i   decimal.Decimal
		err error
	)
	// 뽑기.
	//start, err := decimal.NewFromString("45979977666192131123124453456476567568568657655498755827359801234578023489675")
	start, err := decimal.NewFromString("1")
	require.NoError(t, err)
	host := getHost()
	t.Log("host", host)
	client, err = ethclient.DialContext(context.Background(), host)
	require.NoError(t, err)
	wg := sync.WaitGroup{}
	for i = start; i.Cmp(start.Add(decimal.New(100, 0))) < 0; i = i.Add(decimal.New(1, 0)) {
		// 1. select 1 ~ 2^256
		wg.Add(1)
		num, ok := new(big.Int).SetString(fmt.Sprintf("%v", i), 10)
		require.Equal(t, true, ok)
		go func(num *big.Int) {
			if err := checkAndTake(t, num); err != nil {
				t.Log("F:", num.Uint64(), "error", err)
			} else {
				t.Log("S:", num.Uint64())
			}
			wg.Done()
		}(num)
	}
	wg.Wait()
}

func TestCheckAndTake(t *testing.T) {
	var err error
	client, err = ethclient.DialContext(context.Background(), getHost())
	require.NoError(t, err)
	err = checkAndTake(t, big.NewInt(130))
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

	gasPrice, err := client.SuggestGasPrice(ctx)
	require.NoError(t, err)
	nonceAt, err := client.NonceAt(ctx, addr, nil)
	require.NoError(t, err)
	balance, err := client.BalanceAt(context.Background(), addr, nil)
	require.NoError(t, err)
	if balance.Cmp(big.NewInt(cmp)) < 0 {
		return fmt.Errorf("insufficient fund:%v", balance)
	}
	chain, err := client.ChainID(ctx)
	require.NoError(t, err)
	require.NoError(t, err)
	t.Log("balanceOf", balance.Int64())

	toAddr := common.HexToAddress(to)
	unsignedTx := types.NewTransaction(nonceAt, toAddr, big.NewInt(100), params.TxGas, gasPrice, []byte{})
	signedTx, err := types.SignTx(unsignedTx, types.LatestSignerForChainID(chain), key)
	require.NoError(t, err)

	return client.SendTransaction(ctx, signedTx)

}
