package types_test

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/0xPolygon/polygon-edge/bls"
	"github.com/cometbft/cometbft/crypto/tmhash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/votepool"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	pk1 = ed25519.GenPrivKey().PubKey()
	pk2 = ed25519.GenPrivKey().PubKey()
)

func TestNetGenesisState(t *testing.T) {
	gen := types.NewGenesisState(nil)
	assert.NotNil(t, gen.GenTxs) // https://github.com/cosmos/cosmos-sdk/issues/5086

	gen = types.NewGenesisState(
		[]json.RawMessage{
			[]byte(`{"foo":"bar"}`),
		},
	)
	assert.Equal(t, string(gen.GenTxs[0]), `{"foo":"bar"}`)
}

func TestValidateGenesisMultipleMessages(t *testing.T) {
	desc := stakingtypes.NewDescription("testname", "", "", "", "")
	comm := stakingtypes.CommissionRates{}

	blsSecretKey1, _ := bls.GenerateBlsKey()
	blsPk1 := hex.EncodeToString(blsSecretKey1.PublicKey().Marshal())
	blsProofBuf1, _ := blsSecretKey1.Sign(tmhash.Sum(blsSecretKey1.PublicKey().Marshal()), votepool.DST)
	blsProofBts1, _ := blsProofBuf1.Marshal()
	blsProof1 := hex.EncodeToString(blsProofBts1)
	msg1, err := stakingtypes.NewMsgCreateValidator(
		sdk.AccAddress(pk1.Address()), pk1,
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 50), desc, comm, sdk.OneInt(),
		sdk.AccAddress(pk1.Address()), sdk.AccAddress(pk1.Address()),
		sdk.AccAddress(pk1.Address()), sdk.AccAddress(pk1.Address()), blsPk1, blsProof1)
	require.NoError(t, err)

	blsSecretKey2, _ := bls.GenerateBlsKey()
	blsPk2 := hex.EncodeToString(blsSecretKey2.PublicKey().Marshal())
	blsProofBuf2, _ := blsSecretKey2.Sign(tmhash.Sum(blsSecretKey2.PublicKey().Marshal()), votepool.DST)
	blsProofBts2, _ := blsProofBuf2.Marshal()
	blsProof2 := hex.EncodeToString(blsProofBts2)
	msg2, err := stakingtypes.NewMsgCreateValidator(
		sdk.AccAddress(pk2.Address()), pk2,
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 50), desc, comm, sdk.OneInt(),
		sdk.AccAddress(pk2.Address()), sdk.AccAddress(pk2.Address()),
		sdk.AccAddress(pk2.Address()), sdk.AccAddress(pk2.Address()), blsPk2, blsProof2)
	require.NoError(t, err)

	txConfig := moduletestutil.MakeTestEncodingConfig(staking.AppModuleBasic{}, genutil.AppModuleBasic{}).TxConfig
	txBuilder := txConfig.NewTxBuilder()
	require.NoError(t, txBuilder.SetMsgs(msg1, msg2))

	tx := txBuilder.GetTx()
	genesisState := types.NewGenesisStateFromTx(txConfig.TxJSONEncoder(), []sdk.Tx{tx})

	err = types.ValidateGenesis(genesisState, txConfig.TxJSONDecoder(), types.DefaultMessageValidator)
	require.Error(t, err)
}

func TestValidateGenesisBadMessage(t *testing.T) {
	desc := stakingtypes.NewDescription("testname", "", "", "", "")
	blsSecretKey1, _ := bls.GenerateBlsKey()
	blsPk1 := hex.EncodeToString(blsSecretKey1.PublicKey().Marshal())
	blsProofBuf, _ := blsSecretKey1.Sign(tmhash.Sum(blsSecretKey1.PublicKey().Marshal()), votepool.DST)
	blsProofBts, _ := blsProofBuf.Marshal()
	blsProof := hex.EncodeToString(blsProofBts)
	msg1 := stakingtypes.NewMsgEditValidator(
		sdk.AccAddress(pk1.Address()), desc, nil, nil,
		sdk.AccAddress(pk1.Address()), sdk.AccAddress(pk1.Address()), blsPk1, blsProof,
	)

	txConfig := moduletestutil.MakeTestEncodingConfig(staking.AppModuleBasic{}, genutil.AppModuleBasic{}).TxConfig
	txBuilder := txConfig.NewTxBuilder()
	err := txBuilder.SetMsgs(msg1)
	require.NoError(t, err)

	tx := txBuilder.GetTx()
	genesisState := types.NewGenesisStateFromTx(txConfig.TxJSONEncoder(), []sdk.Tx{tx})

	err = types.ValidateGenesis(genesisState, txConfig.TxJSONDecoder(), types.DefaultMessageValidator)
	require.Error(t, err)
}

func TestGenesisStateFromGenFile(t *testing.T) {
	cdc := codec.NewLegacyAmino()

	genFile := "../../../tests/fixtures/adr-024-coin-metadata_genesis.json"
	genesisState, _, err := types.GenesisStateFromGenFile(genFile)
	require.NoError(t, err)

	var bankGenesis banktypes.GenesisState
	cdc.MustUnmarshalJSON(genesisState[banktypes.ModuleName], &bankGenesis)

	require.True(t, bankGenesis.Params.DefaultSendEnabled)
	require.Equal(t, "1000nametoken,100000000stake", bankGenesis.Balances[0].GetCoins().String())
	require.Equal(t, "0x68F07419B137B1F9e36bf559502f05912C4769D0", bankGenesis.Balances[0].GetAddress().String())
	require.Equal(t, "The native staking token of the Cosmos Hub.", bankGenesis.DenomMetadata[0].GetDescription())
	require.Equal(t, "uatom", bankGenesis.DenomMetadata[0].GetBase())
	require.Equal(t, "matom", bankGenesis.DenomMetadata[0].GetDenomUnits()[1].GetDenom())
	require.Equal(t, []string{"milliatom"}, bankGenesis.DenomMetadata[0].GetDenomUnits()[1].GetAliases())
	require.Equal(t, uint32(3), bankGenesis.DenomMetadata[0].GetDenomUnits()[1].GetExponent())
}
