package ethereum

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"cosmossdk.io/math"

	"github.com/strangelove-ventures/interchaintest/v8/testutil"

	"github.com/srdtrk/solidity-ibc-eureka/e2e/v8/testvalues"
)

type Ethereum struct {
	ChainID         *big.Int
	RPC             string
	EthAPI          EthAPI
	BeaconAPIClient *BeaconAPIClient
	RPCClient       *ethclient.Client

	Faucet *ecdsa.PrivateKey
}

func NewEthereum(ctx context.Context, rpc string, beaconAPIClient *BeaconAPIClient, faucet *ecdsa.PrivateKey) (Ethereum, error) {
	ethClient, err := ethclient.Dial(rpc)
	if err != nil {
		return Ethereum{}, err
	}
	chainID, err := ethClient.ChainID(ctx)
	if err != nil {
		return Ethereum{}, err
	}
	ethAPI, err := NewEthAPI(rpc)
	if err != nil {
		return Ethereum{}, err
	}

	return Ethereum{
		ChainID:         chainID,
		RPC:             rpc,
		EthAPI:          ethAPI,
		BeaconAPIClient: beaconAPIClient,
		RPCClient:       ethClient,
		Faucet:          faucet,
	}, nil
}

// BroadcastMessages broadcasts the provided messages to the given chain and signs them on behalf of the provided user.
// Once the transaction is mined, the receipt is returned.
func (e *Ethereum) BroadcastTx(ctx context.Context, userKey *ecdsa.PrivateKey, gasLimit uint64, address ethcommon.Address, txBz []byte) (*ethtypes.Receipt, error) {
	txOpts, err := e.GetTransactOpts(userKey)
	if err != nil {
		return nil, err
	}

	tx := ethtypes.NewTransaction(
		txOpts.Nonce.Uint64(),
		address,
		txOpts.Value,
		gasLimit,
		txOpts.GasPrice,
		txBz,
	)

	signedTx, err := txOpts.Signer(txOpts.From, tx)
	if err != nil {
		return nil, err
	}

	err = e.RPCClient.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, err
	}

	receipt, err := e.GetTxReciept(ctx, signedTx.Hash())
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

func (e Ethereum) ForgeScript(deployer *ecdsa.PrivateKey, solidityContract string, args ...string) ([]byte, error) {
	args = append(args, "script", "--rpc-url", e.RPC, "--private-key",
		hex.EncodeToString(deployer.D.Bytes()), "--broadcast",
		"--non-interactive", "-vvvv", solidityContract,
	)
	cmd := exec.Command(
		"forge", args...,
	)

	faucetAddress := crypto.PubkeyToAddress(e.Faucet.PublicKey)
	extraEnv := []string{
		fmt.Sprintf("%s=%s", testvalues.EnvKeyE2EFacuetAddress, faucetAddress.Hex()),
	}

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, extraEnv...)

	var stdoutBuf bytes.Buffer

	// Create a MultiWriter to write to both os.Stdout and the buffer
	multiWriter := io.MultiWriter(os.Stdout, &stdoutBuf)

	// Set the command's stdout to the MultiWriter
	cmd.Stdout = multiWriter
	cmd.Stderr = os.Stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		fmt.Println("Error start command", cmd.Args, err)
		return nil, err
	}

	// Get the output as byte slices
	stdoutBytes := stdoutBuf.Bytes()

	return stdoutBytes, nil
}

func (e Ethereum) CreateAndFundUser() (*ecdsa.PrivateKey, error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	address := crypto.PubkeyToAddress(key.PublicKey).Hex()
	if err := e.FundUser(address, testvalues.StartingEthBalance); err != nil {
		return nil, err
	}

	return key, nil
}

func (e Ethereum) FundUser(address string, amount math.Int) error {
	return e.SendEth(e.Faucet, address, amount)
}

func (e Ethereum) SendEth(key *ecdsa.PrivateKey, toAddress string, amount math.Int) error {
	cmd := exec.Command(
		"cast",
		"send",
		toAddress,
		"--value", amount.String(),
		"--private-key", fmt.Sprintf("0x%s", ethcommon.Bytes2Hex(key.D.Bytes())),
		"--rpc-url", e.RPC,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send eth with %s: %w", strings.Join(cmd.Args, " "), err)
	}

	return nil
}

func (e *Ethereum) Height() (int64, error) {
	cmd := exec.Command("cast", "block-number", "--rpc-url", e.RPC)
	stdout, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSpace(string(stdout)), 10, 64)
}

func (e *Ethereum) GetTxReciept(ctx context.Context, hash ethcommon.Hash) (*ethtypes.Receipt, error) {
	var receipt *ethtypes.Receipt
	err := testutil.WaitForCondition(time.Second*40, time.Second, func() (bool, error) {
		var err error
		receipt, err = e.RPCClient.TransactionReceipt(ctx, hash)
		if err != nil {
			return false, nil
		}

		return receipt != nil, nil
	})
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

func (e *Ethereum) GetTransactOpts(key *ecdsa.PrivateKey) (*bind.TransactOpts, error) {
	fromAddress := crypto.PubkeyToAddress(key.PublicKey)
	nonce, err := e.RPCClient.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		nonce = 0
	}

	gasPrice, err := e.RPCClient.SuggestGasPrice(context.Background())
	if err != nil {
		panic(err)
	}

	txOpts, err := bind.NewKeyedTransactorWithChainID(key, e.ChainID)
	if err != nil {
		return nil, err
	}

	txOpts.Nonce = big.NewInt(int64(nonce))
	txOpts.GasPrice = gasPrice

	return txOpts, nil
}
