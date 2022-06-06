package extract

import (
	"encoding/binary"
	"fmt"
	"strconv"

	gogotypes "github.com/gogo/protobuf/types"

	tmdb "github.com/tendermint/tm-db"

	"github.com/piux2/gnobounty7/sink"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// scan the application.db
func ProfileAppDB(psink *sink.PsqlSink) {

	appdb := psink.AppDB

	heights, err := getPruningHeights(appdb)

	if err != nil {

		fmt.Printf("no pruned heights found.\n")

	} else {

		fmt.Printf("prune height %v\n", heights)

	}

	getAppHeights(psink)

}

func getPruningHeights(db tmdb.DB) ([]int64, error) {
	pruneHeightsKey := "s/pruneheights"
	bz, err := db.Get([]byte(pruneHeightsKey))
	if err != nil {
		return nil, fmt.Errorf("failed to get pruned heights: %w", err)
	}
	if len(bz) == 0 {
		return nil, fmt.Errorf("no pruned heights found")
	}

	prunedHeights := make([]int64, len(bz)/8)
	i, offset := 0, 0
	for offset < len(bz) {
		prunedHeights[i] = int64(binary.BigEndian.Uint64(bz[offset : offset+8]))
		i++
		offset += 8
	}

	return prunedHeights, nil
}

func getLatestVersion(db tmdb.DB) int64 {
	latestVersionKey := "s/latest"

	bz, err := db.Get([]byte(latestVersionKey))

	if err != nil {
		panic(err)
	} else if bz == nil {
		return 0
	}

	var latestVersion int64

	if err := gogotypes.StdInt64Unmarshal(&latestVersion, bz); err != nil {
		panic(err)
	}

	return latestVersion
}

func getAppHeights(psink *sink.PsqlSink) {
	appdb := psink.AppDB

	height := getLatestVersion(appdb)

	prefix := "s/"
	start := Search(appdb, prefix, height, BASE)
	psink.Top = height

	fmt.Printf("application.db \t base height: %d, \t top height: %d\n", start, height)

}

// The block store might not stored the record with the same sequences of the original key sequence
// Therefore you might not get correct the base and top height from the database iterator
// We have to implement our own to find the base and top of the height.
func ProfileBlockstoreDB(psink *sink.PsqlSink) {

	db := (*psink.BlockStore).DB()
	tmdb := db.(*tmdb.GoLevelDB)

	keyPrefix := "H:"
	top := psink.Top
	base := Search(tmdb, keyPrefix, top, BASE)

	fmt.Printf("blockstore.db \t base height: %d, \t top height: %d\n", base, top)

}

func findBaseTop(db tmdb.DB, keyPrefix string) (base, top int64) {

	top = MAX_HEIGHT

	top = Search(db, keyPrefix, top, TOP) - 1
	base = Search(db, keyPrefix, top, BASE)
	return base, top

}

const MAX_HEIGHT int64 = int64(1<<63 - 1)
const (
	TOP  = true
	BASE = false
)

//this function is modified from golang's sort.Search function.
// It finds the smallest value between 0 and n when when dectector function is true.
// When max is true, it finds biggest (max) value between 0 and n
// When max is false, it search smallest (min) value,between 0 and n, pre-requisit, the n is <= max *2
// So it always to find max first and then find min.
// it find the base and top in log(n) steps
func Search(db tmdb.DB, prefix string, n int64, max bool) int64 {
	// Define f(-1) == false and f(n) == true.
	// Invariant: f(i-1) == false, f(j) == true.
	c := 0
	i, j := int64(1), n
	for i < j {
		c++

		middle := int64(uint(i+j) >> 1) // avoid overflow when computing h
		// i ≤ middle < j
		key := []byte(prefix + strconv.FormatInt(middle, 10))

		t, err := db.Has(key) // dectector function

		if err != nil {
			fmt.Printf("%s", err)
			t = false
		}

		if t == max {
			i = middle + 1 // preserves f(i-1) == false
		} else {
			j = middle // preserves f(j) == true
		}
	}
	// i == j, f(i-1) == false, and f(j) (= f(i)) == true  =>  answer is i.
	return i
}

func ProfileStateDB(psink *sink.PsqlSink) {

	db := (*psink.StateStore).DB()
	tmdb := db.(*tmdb.GoLevelDB)

	keyPrefix := "abciResponsesKey:"

	top := psink.Top
	base := Search(tmdb, keyPrefix, top, BASE)
	psink.Base = base

	fmt.Printf("state.db \t base height: %d, \t top height: %d\n", psink.Base, psink.Top)

}

func DiscoverPrefixKey(psink *sink.PsqlSink) {

	db := (*psink.StateStore).DB()
	tmdb := db.(*tmdb.GoLevelDB)
	ldb := tmdb.DB()

	fmt.Println("Discover prefix key in state.db")

	for j := 0; j <= 255; j++ {
		a := fmt.Sprint(j)
		i := ldb.NewIterator(util.BytesPrefix([]byte(a)), nil)

		if i.First() == true {
			fmt.Printf("First key %s :  \n", i.Key())

		}
		i.Release()
		err := i.Error()
		if err != nil {

			panic(err)
		}

	}

}
