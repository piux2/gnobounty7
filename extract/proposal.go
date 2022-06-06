package extract

import (
	"fmt"

	"github.com/piux2/gnobounty7/sink"
	putil "github.com/piux2/gnobounty7/util"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	v1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/gogo/protobuf/proto" // used by cosmos-sdk
)

func ExtractProposals(psink *sink.PsqlSink) error {

	cms := *psink.AppStore
	govStoreKey := storetypes.NewKVStoreKey(govtypes.StoreKey)
	storeType := storetypes.StoreTypeIAVL
	cms.MountStoreWithDB(govStoreKey, storeType, nil)

	err := cms.LoadLatestVersion()
	if err != nil {
		panic(err)
	}

	govStore := cms.GetCommitKVStore(govStoreKey)

	if govStore == nil {

		err = fmt.Errorf(">>stakingStore is nil")

		return err

	}

	for i := 1; ; i++ {
		id := uint64(i)
		pKey := govtypes.ProposalKey(id)

		if pKey == nil {

			break
		}

		pBytes := govStore.Get(pKey)

		proposal := v1.Proposal{}

		if err := proto.Unmarshal(pBytes, &proposal); err != nil {
			panic(err)
		}

		anys := proposal.Messages

		for i, any := range anys {

			bytes := any.Value
			ptext := v1beta1.TextProposal{}
			err := proto.Unmarshal(bytes, &ptext)
			if err != nil {

				panic(err)

			}

			putil.PrettyJson(ptext)

			fmt.Printf("any.i %d", i)

		}
	}

	return nil

}
