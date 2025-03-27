package common

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	tmcfg "github.com/cometbft/cometbft/config"
	tmos "github.com/cometbft/cometbft/libs/os"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client/tx"
	sdkcrypto "github.com/cosmos/cosmos-sdk/crypto"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	txsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

//
//nolint:unused
type validator struct {
	chain            *Chain
	Index            int
	Moniker          string
	Mnemonic         string
	KeyInfo          keyring.Record
	privateKey       cryptotypes.PrivKey
	consensusKey     privval.FilePVKey
	consensusPrivKey cryptotypes.PrivKey
	NodeKey          p2p.NodeKey
}

type account struct {
	moniker    string //nolint:unused
	Mnemonic   string
	KeyInfo    keyring.Record
	privateKey cryptotypes.PrivKey
}

func (v *validator) InstanceName() string {
	return fmt.Sprintf("%s%d", v.Moniker, v.Index)
}

func (v *validator) ConfigDir() string {
	return fmt.Sprintf("%s/%s", v.chain.configDir(), v.InstanceName())
}

func (v *validator) createConfig() error {
	p := path.Join(v.ConfigDir(), "config")
	return os.MkdirAll(p, 0o755)
}

func (v *validator) init(genesisState map[string]json.RawMessage) error {
	if err := v.createConfig(); err != nil {
		return err
	}

	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config

	config.SetRoot(v.ConfigDir())
	config.Moniker = v.Moniker

	appState, err := json.MarshalIndent(genesisState, "", " ")
	if err != nil {
		return fmt.Errorf("failed to JSON encode app genesis state: %w", err)
	}

	appGenesis := genutiltypes.AppGenesis{
		ChainID:  v.chain.ID,
		AppState: appState,
		Consensus: &genutiltypes.ConsensusGenesis{
			Validators: nil,
		},
	}

	if err = genutil.ExportGenesisFile(&appGenesis, serverCtx.Config.GenesisFile()); err != nil {
		return fmt.Errorf("failed to export app genesis state: %w", err)
	}

	tmcfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)
	return nil
}

func (v *validator) createNodeKey() error {
	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config

	config.SetRoot(v.ConfigDir())
	config.Moniker = v.Moniker

	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		return err
	}

	v.NodeKey = *nodeKey
	return nil
}

func (v *validator) createConsensusKey() error {
	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config

	config.SetRoot(v.ConfigDir())
	config.Moniker = v.Moniker

	pvKeyFile := config.PrivValidatorKeyFile()
	if err := tmos.EnsureDir(filepath.Dir(pvKeyFile), 0o777); err != nil {
		return err
	}

	pvStateFile := config.PrivValidatorStateFile()
	if err := tmos.EnsureDir(filepath.Dir(pvStateFile), 0o777); err != nil {
		return err
	}

	filePV := privval.LoadOrGenFilePV(pvKeyFile, pvStateFile)
	v.consensusKey = filePV.Key

	return nil
}

func (v *validator) createKeyFromMnemonic(name, mnemonic string) error {
	dir := v.ConfigDir()
	kb, err := keyring.New(KeyringAppName, keyring.BackendTest, dir, nil, Cdc)
	if err != nil {
		return err
	}

	keyringAlgos, _ := kb.SupportedAlgorithms()
	algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), keyringAlgos)
	if err != nil {
		return err
	}

	info, err := kb.NewAccount(name, mnemonic, "", sdk.FullFundraiserPath, algo)
	if err != nil {
		return err
	}

	privKeyArmor, err := kb.ExportPrivKeyArmor(name, keyringPassphrase)
	if err != nil {
		return err
	}

	privKey, _, err := sdkcrypto.UnarmorDecryptPrivKey(privKeyArmor, keyringPassphrase)
	if err != nil {
		return err
	}

	v.KeyInfo = *info
	v.Mnemonic = mnemonic
	v.privateKey = privKey

	return nil
}

func (c *Chain) AddAccountFromMnemonic(counts int) error {
	val0ConfigDir := c.Validators[0].ConfigDir()
	kb, err := keyring.New(KeyringAppName, keyring.BackendTest, val0ConfigDir, nil, Cdc)
	if err != nil {
		return err
	}

	keyringAlgos, _ := kb.SupportedAlgorithms()
	algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), keyringAlgos)
	if err != nil {
		return err
	}

	for i := 0; i < counts; i++ {
		name := fmt.Sprintf("acct-%d", i)
		mnemonic, err := CreateMnemonic()
		if err != nil {
			return err
		}
		info, err := kb.NewAccount(name, mnemonic, "", sdk.FullFundraiserPath, algo)
		if err != nil {
			return err
		}

		privKeyArmor, err := kb.ExportPrivKeyArmor(name, keyringPassphrase)
		if err != nil {
			return err
		}

		privKey, _, err := sdkcrypto.UnarmorDecryptPrivKey(privKeyArmor, keyringPassphrase)
		if err != nil {
			return err
		}
		acct := account{}
		acct.KeyInfo = *info
		acct.Mnemonic = mnemonic
		acct.privateKey = privKey
		c.GenesisAccounts = append(c.GenesisAccounts, &acct)
	}

	return nil
}

func (v *validator) createKey(name string) error {
	mnemonic, err := CreateMnemonic()
	if err != nil {
		return err
	}

	return v.createKeyFromMnemonic(name, mnemonic)
}

func (v *validator) BuildCreateValidatorMsg(amount sdk.Coin) (sdk.Msg, error) {
	description := stakingtypes.NewDescription(v.Moniker, "", "", "", "")
	commissionRates := stakingtypes.CommissionRates{
		Rate:          math.LegacyMustNewDecFromStr("0.1"),
		MaxRate:       math.LegacyMustNewDecFromStr("0.2"),
		MaxChangeRate: math.LegacyMustNewDecFromStr("0.01"),
	}

	valPubKey, err := cryptocodec.FromCmtPubKeyInterface(v.consensusKey.PubKey)
	if err != nil {
		return nil, err
	}

	addr, err := v.KeyInfo.GetAddress()
	if err != nil {
		return nil, err
	}

	return stakingtypes.NewMsgCreateValidator(
		sdk.ValAddress(addr).String(),
		valPubKey,
		amount,
		description,
		commissionRates,
	)
}

func (v *validator) SignMsg(msgs ...sdk.Msg) (*sdktx.Tx, error) {
	txBuilder := EncodingConfig.TxConfig.NewTxBuilder()

	if err := txBuilder.SetMsgs(msgs...); err != nil {
		return nil, err
	}

	txBuilder.SetMemo(fmt.Sprintf("%s@%s:26656", v.NodeKey.ID(), v.InstanceName()))
	txBuilder.SetFeeAmount(sdk.NewCoins())
	txBuilder.SetGasLimit(200000)

	// For SIGN_MODE_DIRECT, calling SetSignatures calls setSignerInfos on
	// TxBuilder under the hood, and SignerInfos is needed to generate the sign
	// bytes. This is the reason for setting SetSignatures here, with a nil
	// signature.
	//
	// Note: This line is not needed for SIGN_MODE_LEGACY_AMINO, but putting it
	// also doesn't affect its generated sign bytes, so for code's simplicity
	// sake, we put it here.
	pk, err := v.KeyInfo.GetPubKey()
	if err != nil {
		return nil, err
	}

	sig := txsigning.SignatureV2{
		PubKey: pk,
		Data: &txsigning.SingleSignatureData{
			SignMode:  txsigning.SignMode_SIGN_MODE_DIRECT,
			Signature: nil,
		},
		Sequence: 0,
	}

	if err := txBuilder.SetSignatures(sig); err != nil {
		return nil, err
	}

	pk, err = v.KeyInfo.GetPubKey()
	if err != nil {
		return nil, err
	}

	signerData := authsigning.SignerData{
		Address:       sdk.AccAddress(pk.Bytes()).String(),
		ChainID:       v.chain.ID,
		AccountNumber: 0,
		Sequence:      0,
		PubKey:        pk,
	}
	sig, err = tx.SignWithPrivKey(
		context.TODO(), txsigning.SignMode_SIGN_MODE_DIRECT, signerData,
		txBuilder, v.privateKey, TxConfig, 0)
	if err != nil {
		return nil, err
	}

	if err := txBuilder.SetSignatures(sig); err != nil {
		return nil, err
	}

	signedTx := txBuilder.GetTx()
	bz, err := EncodingConfig.TxConfig.TxEncoder()(signedTx)
	if err != nil {
		return nil, err
	}

	return DecodeTx(bz)
}
