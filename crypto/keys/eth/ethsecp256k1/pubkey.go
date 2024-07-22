package ethsecp256k1

import (
	"bytes"
	"fmt"

	errorsmod "cosmossdk.io/errors"

	cmtcrypto "github.com/cometbft/cometbft/crypto"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/eth/eip712"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	_ cryptotypes.PubKey   = &PubKey{}
	_ codec.AminoMarshaler = &PubKey{}
)

// Address returns the address of the ECDSA public key.
// The function will return an empty address if the public key is invalid.
func (pubKey *PubKey) Address() cmtcrypto.Address {
	pubkey, err := crypto.DecompressPubkey(pubKey.Key)
	if err != nil {
		return nil
	}

	return cmtcrypto.Address(crypto.PubkeyToAddress(*pubkey).Bytes())
}

// Bytes returns the raw bytes of the ECDSA public key.
func (pubKey *PubKey) Bytes() []byte {
	if pubKey == nil {
		return nil
	}
	bz := make([]byte, len(pubKey.Key))
	copy(bz, pubKey.Key)

	return bz
}

// String implements the fmt.Stringer interface.
func (pubKey *PubKey) String() string {
	return fmt.Sprintf("EthPubKeySecp256k1{%X}", pubKey.Key)
}

// Type returns eth_secp256k1
func (pubKey *PubKey) Type() string {
	return KeyType
}

// Equals returns true if the pubkey type is the same and their bytes are deeply equal.
func (pubKey *PubKey) Equals(other cryptotypes.PubKey) bool {
	return pubKey.Type() == other.Type() && bytes.Equal(pubKey.Bytes(), other.Bytes())
}

// MarshalAmino overrides Amino binary marshalling.
func (pubKey PubKey) MarshalAmino() ([]byte, error) {
	return pubKey.Key, nil
}

// UnmarshalAmino overrides Amino binary marshalling.
func (pubKey *PubKey) UnmarshalAmino(bz []byte) error {
	if len(bz) != PubKeySize {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidPubKey, "invalid pubkey size, expected %d, got %d", PubKeySize, len(bz))
	}
	pubKey.Key = bz

	return nil
}

// MarshalAminoJSON overrides Amino JSON marshalling.
func (pubKey PubKey) MarshalAminoJSON() ([]byte, error) {
	// When we marshal to Amino JSON, we don't marshal the "key" field itself,
	// just its contents (i.e. the key bytes).
	return pubKey.MarshalAmino()
}

// UnmarshalAminoJSON overrides Amino JSON marshalling.
func (pubKey *PubKey) UnmarshalAminoJSON(bz []byte) error {
	return pubKey.UnmarshalAmino(bz)
}

// VerifySignature verifies that the ECDSA public key created a given signature over
// the provided message. It will calculate the Keccak256 hash of the message
// prior to verification and approve verification if the signature can be verified
// from either the original message or its EIP-712 representation.
//
// CONTRACT: The signature should be in [R || S] format.
func (pubKey *PubKey) VerifySignature(msg, sig []byte) bool {
	return pubKey.verifySignatureECDSA(msg, sig) || pubKey.verifySignatureAsEIP712(msg, sig)
}

// Verifies the signature as an EIP-712 signature by first converting the message payload
// to EIP-712 object bytes, then performing ECDSA verification on the hash. This is to support
// signing a Cosmos payload using EIP-712.
func (pubKey *PubKey) verifySignatureAsEIP712(msg, sig []byte) bool {
	eip712Bytes, err := eip712.GetEIP712BytesForMsg(msg)
	if err != nil {
		return false
	}

	if pubKey.verifySignatureECDSA(eip712Bytes, sig) {
		return true
	}

	// Try verifying the signature using the legacy EIP-712 encoding
	legacyEIP712Bytes, err := eip712.LegacyGetEIP712BytesForMsg(msg)
	if err != nil {
		return false
	}

	return pubKey.verifySignatureECDSA(legacyEIP712Bytes, sig)
}

// Perform standard ECDSA signature verification for the given raw bytes and signature.
func (pubKey *PubKey) verifySignatureECDSA(msg, sig []byte) bool {
	if len(sig) == crypto.SignatureLength {
		// remove recovery ID (V) if contained in the signature
		sig = sig[:len(sig)-1]
	}

	// the signature needs to be in [R || S] format when provided to VerifySignature
	return crypto.VerifySignature(pubKey.Key, crypto.Keccak256Hash(msg).Bytes(), sig)
}
