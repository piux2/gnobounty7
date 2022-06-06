package extract

import (
	"fmt"
	"github.com/piux2/gnobounty7/sink"
	putil "github.com/piux2/gnobounty7/util"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/cosmos/cosmos-sdk/codec"
)

// MultisigExtract extract the main multisig account and related key holding address
type MultisigKeyExtract struct {
	MultisigAddress string   `json:"multi_address"`
	AccAddress      []string `json:"acc_address"`
	Threshold       uint     `json:"threshold"`
}

type AccMultisigKeyExtract struct {
	MultisigAddress string `json:"multi_address"`
	AccAddress      string `json:"acc_address"`
	Threshold       uint   `json:"threshold"`
}

func extractMultisig(psink *sink.PsqlSink, sink bool) error {

	cms := *psink.AppStore

	storeType := storetypes.StoreTypeIAVL
	authStoreKey := storetypes.NewKVStoreKey(authtypes.StoreKey)
	cms.MountStoreWithDB(authStoreKey, storeType, nil)
	err := cms.LoadLatestVersion()
	if err != nil {
		panic(err)
	}
	authStore := cms.GetCommitKVStore(authStoreKey)

	if authStore == nil {

		fmt.Println("authStore is nil")

	}

	iterator := sdktypes.KVStorePrefixIterator(authStore, authtypes.AddressStoreKeyPrefix)
	defer iterator.Close()

	ir := codectypes.NewInterfaceRegistry()
	authtypes.RegisterInterfaces(ir)
	vestingtypes.RegisterInterfaces(ir)
	cryptocodec.RegisterInterfaces(ir)

	marshaler := codec.NewProtoCodec(ir)

	addrCounter := 0
	multisigKeyCounter := 0
	accMultiExtracts := []AccMultisigKeyExtract{}

	var acc authtypes.AccountI
	var pubkey cryptotypes.PubKey

	for ; iterator.Valid(); iterator.Next() {

		addrCounter++

		v := iterator.Value()

		err := marshaler.UnmarshalInterface(v, &acc)
		if err != nil {

			panic(err)
		}

		accAddress := acc.GetAddress()
		pubkey = acc.GetPubKey()

		if pubkey == nil || pubkey.Type() == "secp256k1" {

			continue
		}

		if pubkey.Type() == "PubKeyMultisigThreshold" {

			multisigKeyCounter++
			mpk := pubkey.(multisig.PubKey)
			m := AccMultisigKeyExtract{}
			m.MultisigAddress = accAddress.String()
			m.Threshold = mpk.GetThreshold()

			keys := mpk.GetPubKeys()

			for _, v := range keys {

				a := sdktypes.AccAddress(v.Address()).String()

				m.AccAddress = a
				accMultiExtracts = append(accMultiExtracts, m)
			}

		}

	}
	putil.JsonLine(accMultiExtracts)
	fmt.Printf("Total addresses: %d\n", addrCounter)
	fmt.Printf("Total multisig address: %d\n", multisigKeyCounter)

	if sink == false {
		return nil
	}

	return nil

}
