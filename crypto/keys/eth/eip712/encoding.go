// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE
package eip712

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txTypes "github.com/cosmos/cosmos-sdk/types/tx"

	apitypes "github.com/ethereum/go-ethereum/signer/core/apitypes"
)

var (
	ProtoCodec codec.ProtoCodecMarshaler
	AminoCodec *codec.LegacyAmino
)

// GetEIP712BytesForMsg returns the EIP-712 object bytes for the given SignDoc bytes by decoding the bytes into
// an EIP-712 object, then converting via WrapTxToTypedData. See https://eips.ethereum.org/EIPS/eip-712 for more.
func GetEIP712BytesForMsg(signDocBytes []byte) ([]byte, error) {
	typedData, err := GetEIP712TypedDataForMsg(signDocBytes)
	if err != nil {
		return nil, err
	}

	_, rawData, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return nil, fmt.Errorf("could not get EIP-712 object bytes: %w", err)
	}

	return []byte(rawData), nil
}

// GetEIP712TypedDataForMsg returns the EIP-712 TypedData representation for either
// Amino or Protobuf encoded signature doc bytes.
func GetEIP712TypedDataForMsg(signDocBytes []byte) (apitypes.TypedData, error) {
	// Attempt to decode as both Amino and Protobuf since the message format is unknown.
	// If either decode works, we can move forward with the corresponding typed data.
	typedDataAmino, errAmino := decodeAminoSignDoc(signDocBytes)
	if errAmino == nil && isValidEIP712Payload(typedDataAmino) {
		return typedDataAmino, nil
	}
	typedDataProtobuf, errProtobuf := decodeProtobufSignDoc(signDocBytes)
	if errProtobuf == nil && isValidEIP712Payload(typedDataProtobuf) {
		return typedDataProtobuf, nil
	}

	return apitypes.TypedData{}, fmt.Errorf("could not decode sign doc as either Amino or Protobuf.\n amino: %v\n protobuf: %v", errAmino, errProtobuf)
}

// isValidEIP712Payload ensures that the given TypedData does not contain empty fields from
// an improper initialization.
func isValidEIP712Payload(typedData apitypes.TypedData) bool {
	return len(typedData.Message) != 0 && len(typedData.Types) != 0 && typedData.PrimaryType != "" && typedData.Domain != apitypes.TypedDataDomain{}
}

// decodeAminoSignDoc attempts to decode the provided sign doc (bytes) as an Amino payload
// and returns a signable EIP-712 TypedData object.
func decodeAminoSignDoc(signDocBytes []byte) (apitypes.TypedData, error) {
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

	if err := validatePayloadMessages(msgs); err != nil {
		return apitypes.TypedData{}, err
	}

	chainID, err := ParseChainID(aminoDoc.ChainID)
	if err != nil {
		return apitypes.TypedData{}, errors.New("invalid chain ID passed as argument")
	}

	typedData, err := WrapTxToTypedData(
		chainID.Uint64(),
		signDocBytes,
	)
	if err != nil {
		return apitypes.TypedData{}, fmt.Errorf("could not convert to EIP712 representation: %w", err)
	}

	return typedData, nil
}

// decodeProtobufSignDoc attempts to decode the provided sign doc (bytes) as a Protobuf payload
// and returns a signable EIP-712 TypedData object.
func decodeProtobufSignDoc(signDocBytes []byte) (apitypes.TypedData, error) {
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

	if err := validatePayloadMessages(msgs); err != nil {
		return apitypes.TypedData{}, err
	}

	signerInfo := authInfo.SignerInfos[0]

	chainID, err := ParseChainID(signDoc.ChainId)
	if err != nil {
		return apitypes.TypedData{}, fmt.Errorf("invalid chain ID passed as argument: %w", err)
	}

	stdFee := &StdFee{
		Amount: authInfo.Fee.Amount,
		Gas:    authInfo.Fee.GasLimit,
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

	typedData, err := WrapTxToTypedData(
		chainID.Uint64(),
		signBytes,
	)
	if err != nil {
		return apitypes.TypedData{}, err
	}

	return typedData, nil
}

// StdTip is the tips used in a tipped transaction.
type StdTip struct {
	Amount sdk.Coins `json:"amount" yaml:"amount"`
	Tipper string    `json:"tipper" yaml:"tipper"`
}

// StdFee includes the amount of coins paid in fees and the maximum
// gas to be used by the transaction. The ratio yields an effective "gasprice",
// which must be above some miminum to be accepted into the mempool.
// [Deprecated]
type StdFee struct {
	Amount  sdk.Coins `json:"amount" yaml:"amount"`
	Gas     uint64    `json:"gas" yaml:"gas"`
	Payer   string    `json:"payer,omitempty" yaml:"payer"`
	Granter string    `json:"granter,omitempty" yaml:"granter"`
}

// Bytes returns the encoded bytes of a StdFee.
func (fee StdFee) Bytes() []byte {
	if len(fee.Amount) == 0 {
		fee.Amount = sdk.NewCoins()
	}

	bz, err := codec.NewLegacyAmino().MarshalJSON(fee)
	if err != nil {
		panic(err)
	}

	return bz
}

// LegacyMsg defines the old interface a message must fulfill, containing
// Amino signing method and legacy router info.
// Deprecated: Please use `Msg` instead.
type LegacyMsg interface {
	sdk.Msg

	// Get the canonical byte representation of the Msg.
	GetSignBytes() []byte

	// Return the message type.
	// Must be alphanumeric or empty.
	Route() string

	// Returns a human-readable string for the message, intended for utilization
	// within tags
	Type() string
}

// StdSignDoc is replay-prevention structure.
// It includes the result of msg.GetSignBytes(),
// as well as the ChainID (prevent cross chain replay)
// and the Sequence numbers for each signature (prevent
// inchain replay and enforce tx ordering per account).
type StdSignDoc struct {
	AccountNumber uint64            `json:"account_number" yaml:"account_number"`
	Sequence      uint64            `json:"sequence" yaml:"sequence"`
	TimeoutHeight uint64            `json:"timeout_height,omitempty" yaml:"timeout_height"`
	ChainID       string            `json:"chain_id" yaml:"chain_id"`
	Memo          string            `json:"memo" yaml:"memo"`
	Fee           json.RawMessage   `json:"fee" yaml:"fee"`
	Msgs          []json.RawMessage `json:"msgs" yaml:"msgs"`
	Tip           *StdTip           `json:"tip,omitempty" yaml:"tip"`
}

// StdSignBytes returns the bytes to sign for a transaction.
func StdSignBytes(chainID string, accnum, sequence, timeout uint64, fee StdFee, msgs []sdk.Msg, memo string, tip *txTypes.Tip) []byte {
	msgsBytes := make([]json.RawMessage, 0, len(msgs))
	for _, msg := range msgs {
		legacyMsg, ok := msg.(LegacyMsg)
		if !ok {
			panic(fmt.Errorf("expected %T when using amino JSON", (*LegacyMsg)(nil)))
		}

		msgsBytes = append(msgsBytes, json.RawMessage(legacyMsg.GetSignBytes()))
	}

	var stdTip *StdTip
	if tip != nil {
		if tip.Tipper == "" {
			panic(fmt.Errorf("tipper cannot be empty"))
		}

		stdTip = &StdTip{Amount: tip.Amount, Tipper: tip.Tipper}
	}

	bz, err := codec.NewLegacyAmino().MarshalJSON(StdSignDoc{
		AccountNumber: accnum,
		ChainID:       chainID,
		Fee:           json.RawMessage(fee.Bytes()),
		Memo:          memo,
		Msgs:          msgsBytes,
		Sequence:      sequence,
		TimeoutHeight: timeout,
		Tip:           stdTip,
	})
	if err != nil {
		panic(err)
	}

	return sdk.MustSortJSON(bz)
}

// validateCodecInit ensures that both Amino and Protobuf encoding codecs have been set on app init,
// so the module does not panic if either codec is not found.
func validateCodecInit() error {
	if AminoCodec == nil || ProtoCodec == nil {
		return errors.New("missing codec: codecs have not been properly initialized using SetEncodingConfig")
	}

	return nil
}

// validatePayloadMessages ensures that the transaction messages can be represented in an EIP-712
// encoding by checking that messages exist and share a single signer.
func validatePayloadMessages(msgs []sdk.Msg) error {
	if len(msgs) == 0 {
		return errors.New("unable to build EIP-712 payload: transaction does contain any messages")
	}

	var msgSigner sdk.AccAddress

	for i, m := range msgs {
		if len(m.GetSigners()) != 1 {
			return errors.New("unable to build EIP-712 payload: expect exactly 1 signer")
		}

		if i == 0 {
			msgSigner = m.GetSigners()[0]
			continue
		}

		if !msgSigner.Equals(m.GetSigners()[0]) {
			return errors.New("unable to build EIP-712 payload: multiple signers detected")
		}
	}

	return nil
}
