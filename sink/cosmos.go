package sink

import (
	"fmt"

	tmdb "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/piux2/gnobounty7/config"
)

func LoadAppDB(cfg *config.Config) tmdb.DB {

	fmt.Printf("Load %s ... ", cfg.DBDir+"application.db")

	appdb, err := tmdb.NewDB(cfg.AppDB, tmdb.BackendType(cfg.DBBackend), cfg.DBDir)

	if err != nil {

		panic(err)
	}

	fmt.Printf("Done.\n")
	return appdb
}

func LoadAppStore(appdb tmdb.DB) *storetypes.CommitMultiStore {

	//load cms commited multistores from application.db

	cms := store.NewCommitMultiStore(appdb)

	//	stakingStoreKey := storetypes.NewKVStoreKey(stakingtypes.StoreKey)
	//authStoreKey := storetypes.NewKVStoreKey(authtypes.StoreKey)
	bankStoreKey := storetypes.NewKVStoreKey(banktypes.StoreKey)

	// others might use storetypes.StoreTypeDB
	// QuickSync's database uses IAVL
	storeType := storetypes.StoreTypeIAVL

	//cms.MountStoreWithDB(stakingStoreKey, storeType, nil)
	//cms.MountStoreWithDB(authStoreKey, storeType, nil)
	cms.MountStoreWithDB(bankStoreKey, storeType, nil)

	fmt.Printf("Load application state at the latest height ... ")
	err := cms.LoadLatestVersion()
	if err != nil {
		panic(err)
	}

	// you have to use the same storeKey to mount and get store.
	// because the key is a pointer so every time you new a storeKey it is defferent
	// Therefor the mount are different. If you mount a store with DB with one key  and
	// create a new store key to retrive the store, you will get nil
	// Don't know why it will use a struct pointer as key, and a the mean time KVStore still
	// check duplication of name in stead.

	bankStore := cms.GetCommitKVStore(bankStoreKey)

	if bankStore == nil {

		fmt.Println("bankStore is nil")

	}

	fmt.Printf("Done\n")

	return &cms
}
