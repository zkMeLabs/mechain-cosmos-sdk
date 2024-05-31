package eip712

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	txTypes "github.com/cosmos/cosmos-sdk/types/tx"

	apitypes "github.com/ethereum/go-ethereum/signer/core/apitypes"
)

type aminoMessage struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// RootCodespace is the codespace for all errors defined in this package
const RootCodespace = "cosmos"

var (
	regexChainID         = `[a-z]{1,}`
	regexEIP155Separator = `_{1}`
	regexEIP155          = `[1-9][0-9]*`
	regexEpochSeparator  = `-{1}`
	regexEpoch           = `[1-9][0-9]*`
	evmosChainID         = regexp.MustCompile(fmt.Sprintf(`^(%s)%s(%s)%s(%s)$`,
		regexChainID,
		regexEIP155Separator,
		regexEIP155,
		regexEpochSeparator,
		regexEpoch))
	// ErrInvalidChainID returns an error resulting from an invalid chain ID.
	ErrInvalidChainID = errorsmod.Register(RootCodespace, 3, "invalid chain ID")
)

// ParseChainID parses a string chain identifier's epoch to an Ethereum-compatible
// chain-id in *big.Int format. The function returns an error if the chain-id has an invalid format
func ParseChainID(chainID string) (*big.Int, error) {
	chainID = strings.TrimSpace(chainID)
	if len(chainID) > 48 {
		return nil, errorsmod.Wrapf(ErrInvalidChainID, "chain-id '%s' cannot exceed 48 chars", chainID)
	}

	matches := evmosChainID.FindStringSubmatch(chainID)
	if matches == nil || len(matches) != 4 || matches[1] == "" {
		return nil, errorsmod.Wrapf(ErrInvalidChainID, "%s: %v", chainID, matches)
	}

	// verify that the chain-id entered is a base 10 integer
	chainIDInt, ok := new(big.Int).SetString(matches[2], 10)
	if !ok {
		return nil, errorsmod.Wrapf(ErrInvalidChainID, "epoch %s must be base-10 integer format", matches[2])
	}

	return chainIDInt, nil
}

// LegacyGetEIP712BytesForMsg returns the EIP-712 object bytes for the given SignDoc bytes by decoding the bytes into
// an EIP-712 object, then converting via LegacyWrapTxToTypedData. See https://eips.ethereum.org/EIPS/eip-712 for more.
func LegacyGetEIP712BytesForMsg(signDocBytes []byte) ([]byte, error) {
	typedData, err := LegacyGetEIP712TypedDataForMsg(signDocBytes)
	if err != nil {
		return nil, err
	}

	_, rawData, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return nil, fmt.Errorf("could not get EIP-712 object bytes: %w", err)
	}

	return []byte(rawData), nil
}

// LegacyGetEIP712TypedDataForMsg returns the EIP-712 TypedData representation for either
// Amino or Protobuf encoded signature doc bytes.
func LegacyGetEIP712TypedDataForMsg(signDocBytes []byte) (apitypes.TypedData, error) {
	// Attempt to decode as both Amino and Protobuf since the message format is unknown.
	// If either decode works, we can move forward with the corresponding typed data.
	typedDataAmino, errAmino := legacyDecodeAminoSignDoc(signDocBytes)
	if errAmino == nil && isValidEIP712Payload(typedDataAmino) {
		return typedDataAmino, nil
	}
	typedDataProtobuf, errProtobuf := legacyDecodeProtobufSignDoc(signDocBytes)
	if errProtobuf == nil && isValidEIP712Payload(typedDataProtobuf) {
		return typedDataProtobuf, nil
	}

	return apitypes.TypedData{}, fmt.Errorf("could not decode sign doc as either Amino or Protobuf.\n amino: %v\n protobuf: %v", errAmino, errProtobuf)
}

// legacyDecodeAminoSignDoc attempts to decode the provided sign doc (bytes) as an Amino payload
// and returns a signable EIP-712 TypedData object.
func legacyDecodeAminoSignDoc(signDocBytes []byte) (apitypes.TypedData, error) {
	// Ensure codecs have been initialized
	if err := validateCodecInit(); err != nil {
		return apitypes.TypedData{}, err
	}

	var aminoDoc StdSignDoc
	if err := AminoCodec.UnmarshalJSON(signDocBytes, &aminoDoc); err != nil {
		return apitypes.TypedData{}, err
	}

	var fees StdFee
	if err := AminoCodec.UnmarshalJSON(aminoDoc.Fee, &fees); err != nil {
		return apitypes.TypedData{}, err
	}

	// Validate payload messages
	msgs := make([]sdk.Msg, len(aminoDoc.Msgs))
	for i, jsonMsg := range aminoDoc.Msgs {
		var m sdk.Msg
		if err := AminoCodec.UnmarshalJSON(jsonMsg, &m); err != nil {
			return apitypes.TypedData{}, fmt.Errorf("failed to unmarshal sign doc message: %w", err)
		}
		msgs[i] = m
	}

	if err := legacyValidatePayloadMessages(msgs); err != nil {
		return apitypes.TypedData{}, err
	}

	// Use first message for fee payer and type inference
	msg := msgs[0]

	// By convention, the fee payer is the first address in the list of signers.
	feePayer := msg.GetSigners()[0]
	feeDelegation := &FeeDelegationOptions{
		FeePayer: feePayer,
	}

	chainID, err := ParseChainID(aminoDoc.ChainID)
	if err != nil {
		return apitypes.TypedData{}, errors.New("invalid chain ID passed as argument")
	}

	typedData, err := LegacyWrapTxToTypedData(
		ProtoCodec,
		chainID.Uint64(),
		msg,
		signDocBytes,
		feeDelegation,
	)
	if err != nil {
		return apitypes.TypedData{}, fmt.Errorf("could not convert to EIP712 representation: %w", err)
	}

	return typedData, nil
}

// legacyDecodeProtobufSignDoc attempts to decode the provided sign doc (bytes) as a Protobuf payload
// and returns a signable EIP-712 TypedData object.
func legacyDecodeProtobufSignDoc(signDocBytes []byte) (apitypes.TypedData, error) {
	// Ensure codecs have been initialized
	if err := validateCodecInit(); err != nil {
		return apitypes.TypedData{}, err
	}

	signDoc := &txTypes.SignDoc{}
	if err := signDoc.Unmarshal(signDocBytes); err != nil {
		return apitypes.TypedData{}, err
	}

	authInfo := &txTypes.AuthInfo{}
	if err := authInfo.Unmarshal(signDoc.AuthInfoBytes); err != nil {
		return apitypes.TypedData{}, err
	}

	body := &txTypes.TxBody{}
	if err := body.Unmarshal(signDoc.BodyBytes); err != nil {
		return apitypes.TypedData{}, err
	}

	// Until support for these fields is added, throw an error at their presence
	if body.TimeoutHeight != 0 || len(body.ExtensionOptions) != 0 || len(body.NonCriticalExtensionOptions) != 0 {
		return apitypes.TypedData{}, errors.New("body contains unsupported fields: TimeoutHeight, ExtensionOptions, or NonCriticalExtensionOptions")
	}

	if len(authInfo.SignerInfos) != 1 {
		return apitypes.TypedData{}, fmt.Errorf("invalid number of signer infos provided, expected 1 got %v", len(authInfo.SignerInfos))
	}

	// Validate payload messages
	msgs := make([]sdk.Msg, len(body.Messages))
	for i, protoMsg := range body.Messages {
		var m sdk.Msg
		if err := ProtoCodec.UnpackAny(protoMsg, &m); err != nil {
			return apitypes.TypedData{}, fmt.Errorf("could not unpack message object with error %w", err)
		}
		msgs[i] = m
	}

	if err := legacyValidatePayloadMessages(msgs); err != nil {
		return apitypes.TypedData{}, err
	}

	// Use first message for fee payer and type inference
	msg := msgs[0]

	signerInfo := authInfo.SignerInfos[0]

	chainID, err := ParseChainID(signDoc.ChainId)
	if err != nil {
		return apitypes.TypedData{}, fmt.Errorf("invalid chain ID passed as argument: %w", err)
	}

	stdFee := &StdFee{
		Amount: authInfo.Fee.Amount,
		Gas:    authInfo.Fee.GasLimit,
	}

	feePayer := msg.GetSigners()[0]
	feeDelegation := &FeeDelegationOptions{
		FeePayer: feePayer,
	}

	tip := authInfo.Tip

	// WrapTxToTypedData expects the payload as an Amino Sign Doc
	signBytes := StdSignBytes(
		signDoc.ChainId,
		signDoc.AccountNumber,
		signerInfo.Sequence,
		body.TimeoutHeight,
		*stdFee,
		msgs,
		body.Memo,
		tip,
	)

	typedData, err := LegacyWrapTxToTypedData(
		ProtoCodec,
		chainID.Uint64(),
		msg,
		signBytes,
		feeDelegation,
	)
	if err != nil {
		return apitypes.TypedData{}, err
	}

	return typedData, nil
}

// validatePayloadMessages ensures that the transaction messages can be represented in an EIP-712
// encoding by checking that messages exist, are of the same type, and share a single signer.
func legacyValidatePayloadMessages(msgs []sdk.Msg) error {
	if len(msgs) == 0 {
		return errors.New("unable to build EIP-712 payload: transaction does contain any messages")
	}

	var msgType string
	var msgSigner sdk.AccAddress

	for i, m := range msgs {
		t, err := getMsgType(m)
		if err != nil {
			return err
		}

		if len(m.GetSigners()) != 1 {
			return errors.New("unable to build EIP-712 payload: expect exactly 1 signer")
		}

		if i == 0 {
			msgType = t
			msgSigner = m.GetSigners()[0]
			continue
		}

		if t != msgType {
			return errors.New("unable to build EIP-712 payload: different types of messages detected")
		}

		if !msgSigner.Equals(m.GetSigners()[0]) {
			return errors.New("unable to build EIP-712 payload: multiple signers detected")
		}
	}

	return nil
}

// getMsgType returns the message type prefix for the given Cosmos SDK Msg
func getMsgType(msg sdk.Msg) (string, error) {
	jsonBytes, err := AminoCodec.MarshalJSON(msg)
	if err != nil {
		return "", err
	}

	var jsonMsg aminoMessage
	if err := json.Unmarshal(jsonBytes, &jsonMsg); err != nil {
		return "", err
	}

	// Verify Type was successfully filled in
	if jsonMsg.Type == "" {
		return "", errors.New("could not decode message: type is missing")
	}

	return jsonMsg.Type, nil
}
