package extract

import (
	"fmt"
	"github.com/piux2/gnobounty7/sink"
	putil "github.com/piux2/gnobounty7/util"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	sdktypes "github.com/cosmos/cosmos-sdk/types"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/gogo/protobuf/proto" // used by cosmos-sdk
)

//prepare the data read to insert in to the databse.
type ValExtract struct {
	ValAddress string `json:"val_address"`
	AccAddress string `json:"acc_address"`
	Moniker    string `json:"moniker"`
}

// Since there are not big data out put, we sink it and out put json at the same time
func ExtractValidators(psink *sink.PsqlSink, sink bool) error {

	cms := *psink.AppStore

	stakingStoreKey := storetypes.NewKVStoreKey(stakingtypes.StoreKey)
	storeType := storetypes.StoreTypeIAVL
	cms.MountStoreWithDB(stakingStoreKey, storeType, nil)
	err := cms.LoadLatestVersion()
	if err != nil {
		panic(err)
	}

	stakingStore := cms.GetCommitKVStore(stakingStoreKey)

	if stakingStore == nil {

		fmt.Println("stakingStore is nil")

		return nil
	}

	iterator := sdktypes.KVStorePrefixIterator(stakingStore, stakingtypes.ValidatorsKey) //stakingtypes.ValidatorsKey is the key prefix

	defer iterator.Close()
	valExtracts := []ValExtract{}
	i := 0

	for ; iterator.Valid(); iterator.Next() {

		address := stakingtypes.AddressFromLastValidatorPowerKey(iterator.Key())

		value := stakingStore.Get(stakingtypes.GetValidatorKey(address))
		if value == nil {
			err := fmt.Errorf("No validator value with address %s", address)
			panic(err)
		}

		validator := stakingtypes.Validator{}

		if err := proto.Unmarshal(value, &validator); err != nil {

			panic(err)

		}

		valAddr, err := sdktypes.ValAddressFromBech32(validator.OperatorAddress)
		if err != nil {

			panic(err)
		}

		accAddr := sdktypes.AccAddress(valAddr).String()

		ve := ValExtract{}
		ve.ValAddress = valAddr.String()
		ve.AccAddress = accAddr
		ve.Moniker = validator.Description.Moniker
		valExtracts = append(valExtracts, ve)
		i++
	}

	putil.JsonLine(valExtracts)

	if sink == false {

		return nil
	}
	err = InsertValExtracts(psink, valExtracts)

	return err

}

func InsertValExtracts(psink *sink.PsqlSink, es []ValExtract) error {

	b := sink.InsertBatch{}

	b.Size = 200
	b.ValueStr = "(?,?,?)"
	b.Statement = "INSERT INTO validator (val_address, acc_address, moniker) VALUES "

	v := []string{}
	a := []interface{}{}

	for _, e := range es {

		v = append(v, b.ValueStr)
		a = append(a, e.ValAddress, e.AccAddress, e.Moniker)

		b.ValueStrings = v
		b.ValueArgs = a
		psink.Batch = b

		if len(v) >= b.Size {

			err := psink.ExecBatch()
			if err != nil {

				return fmt.Errorf(">Failed to execute batch insert %v  ", err)

			}
			v = []string{}
			a = []interface{}{}

		}

	}
	// flush the remaining elements in the batch
	err := psink.ExecBatch()

	return err

}
