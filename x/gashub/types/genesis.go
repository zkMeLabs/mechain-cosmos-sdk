package types

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
)

// Validate performs basic validation of supply genesis data returning an
// error for any failed validation criteria.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	seenMsgGasParams := make(map[string]bool)

	for _, mgp := range gs.GetMsgGasParams() {
		if _, exists := seenMsgGasParams[mgp.MsgTypeUrl]; exists {
			return fmt.Errorf("duplicate msg gas params found: '%s'", mgp.MsgTypeUrl)
		}
		if err := mgp.Validate(); err != nil {
			return err
		}
		seenMsgGasParams[mgp.MsgTypeUrl] = true
	}

	return nil
}

// NewGenesisState - Create a new genesis state
func NewGenesisState(params Params, msgGasParamsSet []MsgGasParams) *GenesisState {
	return &GenesisState{
		Params:       params,
		MsgGasParams: msgGasParamsSet,
	}
}

// DefaultGenesisState - Return a default genesis state
func DefaultGenesisState() *GenesisState {
	defaultMsgGasParamsSet := []MsgGasParams{
		*NewMsgGasParamsWithFixedGas("/cosmos.auth.v1beta1.MsgUpdateParams", 0),
		*NewMsgGasParamsWithFixedGas("/cosmos.bank.v1beta1.MsgUpdateParams", 0),
		*NewMsgGasParamsWithFixedGas("/cosmos.consensus.v1.MsgUpdateParams", 0),
		*NewMsgGasParamsWithFixedGas("/cosmos.crisis.v1.MsgUpdateParams", 0),
		*NewMsgGasParamsWithFixedGas("/cosmos.crosschain.v1.MsgUpdateParams", 0),
		*NewMsgGasParamsWithFixedGas("/cosmos.crosschain.v1.MsgUpdateChannelPermissions", 0),
		*NewMsgGasParamsWithFixedGas("/cosmos.distribution.v1beta1.MsgUpdateParams", 0),
		*NewMsgGasParamsWithFixedGas("/cosmos.gashub.v1beta1.MsgUpdateParams", 0),
		*NewMsgGasParamsWithFixedGas("/cosmos.gov.v1.MsgUpdateParams", 0),
		*NewMsgGasParamsWithFixedGas("/cosmos.mint.v1beta1.MsgUpdateParams", 0),
		*NewMsgGasParamsWithFixedGas("/cosmos.oracle.v1.MsgUpdateParams", 0),
		*NewMsgGasParamsWithFixedGas("/cosmos.slashing.v1beta1.MsgUpdateParams", 0),
		*NewMsgGasParamsWithFixedGas("/cosmos.staking.v1beta1.MsgUpdateParams", 0),
		*NewMsgGasParamsWithFixedGas("/mechain.bridge.MsgUpdateParams", 0),
		*NewMsgGasParamsWithFixedGas("/mechain.sp.MsgUpdateParams", 0),
		*NewMsgGasParamsWithFixedGas("/mechain.storage.MsgUpdateParams", 0),
		*NewMsgGasParamsWithFixedGas("/mechain.payment.MsgUpdateParams", 0),
		*NewMsgGasParamsWithFixedGas("/mechain.challenge.MsgUpdateParams", 0),
		*NewMsgGasParamsWithFixedGas("/mechain.permission.MsgUpdateParams", 0),
		*NewMsgGasParamsWithFixedGas("/cosmos.authz.v1beta1.MsgExec", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/cosmos.authz.v1beta1.MsgRevoke", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/cosmos.bank.v1beta1.MsgSend", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/cosmos.distribution.v1beta1.MsgSetWithdrawAddress", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommission", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/cosmos.feegrant.v1beta1.MsgRevokeAllowance", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/cosmos.gov.v1.MsgDeposit", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/cosmos.gov.v1.MsgSubmitProposal", 2e6),
		*NewMsgGasParamsWithFixedGas("/cosmos.gov.v1.MsgVote", 2e6),
		*NewMsgGasParamsWithFixedGas("/cosmos.gov.v1.MsgVoteWeighted", 2e6),
		*NewMsgGasParamsWithFixedGas("/cosmos.oracle.v1.MsgClaim", 1e3),
		*NewMsgGasParamsWithFixedGas("/cosmos.slashing.v1beta1.MsgUnjail", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/cosmos.staking.v1beta1.MsgBeginRedelegate", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/cosmos.staking.v1beta1.MsgCancelUnbondingDelegation", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/cosmos.staking.v1beta1.MsgCreateValidator", 2e6),
		*NewMsgGasParamsWithFixedGas("/cosmos.staking.v1beta1.MsgDelegate", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/cosmos.staking.v1beta1.MsgEditValidator", 2e6),
		*NewMsgGasParamsWithFixedGas("/cosmos.staking.v1beta1.MsgUndelegate", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.bridge.MsgTransferOut", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.sp.MsgCreateStorageProvider", 2e6),
		*NewMsgGasParamsWithFixedGas("/mechain.sp.MsgDeposit", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.sp.MsgEditStorageProvider", 2e6),
		*NewMsgGasParamsWithFixedGas("/mechain.sp.MsgUpdateSpStoragePrice", 2e6),
		*NewMsgGasParamsWithFixedGas("/mechain.sp.MsgUpdateStorageProviderStatus", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.storage.MsgCreateBucket", 2.4e3),
		*NewMsgGasParamsWithFixedGas("/mechain.storage.MsgDeleteBucket", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.storage.MsgMirrorBucket", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.storage.MsgUpdateBucketInfo", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.storage.MsgCreateObject", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.storage.MsgSealObject", 1.2e2),
		*NewMsgGasParamsWithFixedGas("/mechain.storage.MsgMirrorObject", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.storage.MsgRejectSealObject", 1.2e4),
		*NewMsgGasParamsWithFixedGas("/mechain.storage.MsgDeleteObject", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.storage.MsgCopyObject", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.storage.MsgCancelCreateObject", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.storage.MsgUpdateObjectInfo", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.storage.MsgDiscontinueObject", 2.4e3),
		*NewMsgGasParamsWithFixedGas("/mechain.storage.MsgDiscontinueBucket", 2.4e3),
		*NewMsgGasParamsWithFixedGas("/mechain.storage.MsgCreateGroup", 2.4e3),
		*NewMsgGasParamsWithFixedGas("/mechain.storage.MsgDeleteGroup", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.storage.MsgLeaveGroup", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.storage.MsgUpdateGroupMember", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.storage.MsgUpdateGroupExtra", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.storage.MsgRenewGroupMember", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.storage.MsgMirrorGroup", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.storage.MsgPutPolicy", 2.4e3),
		*NewMsgGasParamsWithFixedGas("/mechain.storage.MsgDeletePolicy", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.storage.MsgMigrateBucket", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.storage.MsgCancelMigrateBucket", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.storage.MsgCompleteMigrateBucket", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.storage.MsgRejectMigrateBucket", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.payment.MsgCreatePaymentAccount", 2e5),
		*NewMsgGasParamsWithFixedGas("/mechain.payment.MsgDeposit", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.payment.MsgWithdraw", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.payment.MsgDisableRefund", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.challenge.MsgSubmit", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.challenge.MsgAttest", 1e2),
		*NewMsgGasParamsWithFixedGas("/mechain.virtualgroup.MsgCreateGlobalVirtualGroup", 1e6),
		*NewMsgGasParamsWithFixedGas("/mechain.virtualgroup.MsgDeleteGlobalVirtualGroup", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.virtualgroup.MsgDeposit", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.virtualgroup.MsgWithdraw", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.virtualgroup.MsgSettle", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.virtualgroup.MsgSwapOut", 2.4e4),
		*NewMsgGasParamsWithFixedGas("/mechain.virtualgroup.MsgCompleteSwapOut", 2.4e4),
		*NewMsgGasParamsWithFixedGas("/mechain.virtualgroup.MsgCancelSwapOut", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.virtualgroup.MsgStorageProviderExit", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.virtualgroup.MsgCompleteStorageProviderExit", 1.2e3),
		*NewMsgGasParamsWithFixedGas("/mechain.virtualgroup.MsgUpdateParams", 1.2e3),
		*NewMsgGasParamsWithDynamicGas(
			"/cosmos.authz.v1beta1.MsgGrant",
			&MsgGasParams_GrantType{
				GrantType: &MsgGasParams_DynamicGasParams{
					FixedGas:   8e2,
					GasPerItem: 8e2,
				},
			},
		),
		*NewMsgGasParamsWithDynamicGas(
			"/cosmos.bank.v1beta1.MsgMultiSend",
			&MsgGasParams_MultiSendType{
				MultiSendType: &MsgGasParams_DynamicGasParams{
					FixedGas:   8e2,
					GasPerItem: 8e2,
				},
			},
		),
		*NewMsgGasParamsWithDynamicGas(
			"/cosmos.feegrant.v1beta1.MsgGrantAllowance",
			&MsgGasParams_GrantAllowanceType{
				GrantAllowanceType: &MsgGasParams_DynamicGasParams{
					FixedGas:   8e2,
					GasPerItem: 8e2,
				},
			},
		),
	}
	return NewGenesisState(DefaultParams(), defaultMsgGasParamsSet)
}

// GetGenesisStateFromAppState returns x/gashub GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.Codec, appState map[string]json.RawMessage) GenesisState {
	var genesisState GenesisState

	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return genesisState
}
