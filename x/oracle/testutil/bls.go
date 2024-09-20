package testutil

import (
	"fmt"

	"github.com/0xPolygon/polygon-edge/bls"
	"github.com/cometbft/cometbft/votepool"
)

type Vote struct {
	PubKey    [128]byte
	Signature [64]byte
}

func (vote *Vote) Verify(eventHash []byte) error {
	blsPubKey, err := bls.UnmarshalPublicKey(vote.PubKey[:])
	if err != nil {
		return err
	}
	sig, err := bls.UnmarshalSignature(vote.Signature[:])
	if err != nil {
		return err
	}
	if !sig.Verify(blsPubKey, eventHash, votepool.DST) {
		return fmt.Errorf("verify sig error")
	}
	return nil
}

func AggregatedSignature(votes []*Vote) ([]byte, error) {
	// Prepare aggregated vote signature
	signatures := make(bls.Signatures, 0, len(votes))
	for _, v := range votes {
		signature, _ := bls.UnmarshalSignature(v.Signature[:])
		signatures = append(signatures, signature)
	}
	return signatures.Aggregate().Marshal()
}

type VoteSigner struct {
	privKey *bls.PrivateKey
	pubKey  *bls.PublicKey
}

func NewVoteSignerV2(privkey []byte) (*VoteSigner, error) {
	privKey, err := bls.UnmarshalPrivateKey(privkey)
	if err != nil {
		return nil, err
	}
	pubKey := privKey.PublicKey()
	return &VoteSigner{
		privKey: privKey,
		pubKey:  pubKey,
	}, nil
}

// SignVote sign a vote, data is used to signed to generate the signature
func (signer *VoteSigner) SignVote(vote *Vote, data []byte) error {
	signature, _ := signer.privKey.Sign(data, votepool.DST)
	signaturebts, _ := signature.Marshal()
	copy(vote.PubKey[:], signer.pubKey.Marshal())
	copy(vote.Signature[:], signaturebts)
	return nil
}

func GenerateBlsSig(privKeys []*bls.PrivateKey, data []byte) []byte {
	privateKey1, _ := privKeys[0].Marshal()
	privateKey2, _ := privKeys[1].Marshal()
	privateKey3, _ := privKeys[2].Marshal()

	validatorSigner1, _ := NewVoteSignerV2(privateKey1)
	validatorSigner2, _ := NewVoteSignerV2(privateKey2)
	validatorSigner3, _ := NewVoteSignerV2(privateKey3)

	var vote1 Vote
	validatorSigner1.SignVote(&vote1, data)
	err := vote1.Verify(data)
	if err != nil {
		panic("verify sig error")
	}

	var vote2 Vote
	validatorSigner2.SignVote(&vote2, data)
	err = vote2.Verify(data)
	if err != nil {
		panic("verify sig error")
	}

	var vote3 Vote
	validatorSigner3.SignVote(&vote3, data)
	err = vote3.Verify(data)
	if err != nil {
		panic("verify sig error")
	}

	var votes []*Vote
	votes = append(votes, &vote1)
	votes = append(votes, &vote2)
	votes = append(votes, &vote3)

	aggregatedSignature, _ := AggregatedSignature(votes)
	return aggregatedSignature
}
