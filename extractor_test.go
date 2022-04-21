package main

import (
	"flag"
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

var bfile *string
var dfile *string

func init() {

	bfile = flag.String("b", "", "balances.json file name")
	dfile = flag.String("d", "", "delegations.json file name")

	if *bfile == "" || *dfile == "" {

		*bfile = "test_balances.json"
		*dfile = "test_delegations.json"

		fmt.Println("\nyou can load your own data to test with the flag options -b balancefile.json and -d delegationsfile.json ")

	}

}

// The testing data are not real data on chain
// with uatom
var b1 = Balance{
	Address: "cosmos1qcqgpxeyv6w3vp76e5qg39zf6fqwledswt7l3d",
	Coins:   []Coin{Coin{Amount: "34141001", Denom: "uatom"}},
}

// empty coin[]
var b2 = Balance{Address: "cosmos1qbqdng8wen7lw392slzmjdr8vdeg9ermjxejer", Coins: nil}

// no delegations
var b3 = Balance{
	Address: "cosmos1qajhwf6gs0ezyjmx2t96qjvyxugaefuwpf7qae",
	Coins:   []Coin{Coin{Amount: "11343922", Denom: "uatom"}},
}

//no wallet address found, this should not ever happen on chain state. but let's assume
// There is discrepancy in the exported state data.
var d1 = Delegation{
	DelegatorAddress: "cosmos1qrsxstyxns0eld4fwsslz3canzezwv8waa2jew",
	Shares:           "173300.000000000000000000",
	ValidatorAddress: "cosmosvaloper1n229vhepft6wnkt5tjpwmxdmcnwz55jv3vp7ed",
}

//same delegator adddress as b1
var d2 = Delegation{
	DelegatorAddress: "cosmos1qcqgpxeyv6w3vp76e5qg39zf6fqwledswt7l3d",
	Shares:           "1039602.000000000000000000",
	ValidatorAddress: "cosmosvaloper1tfk30mq5vgqjdly92kkhhq3raev2hnz6eete3",
}

//the first delegation of b2
var d3 = Delegation{
	DelegatorAddress: "cosmos1qbqdng8wen7lw392slzmjdr8vdeg9ermjxejer",
	Shares:           "2070000.000000000000000000",
	ValidatorAddress: "cosmosvaloper196ax4vc0lwpxndu9dyhvca7jhxp70rmcvrj90c",
}

//the second delegation of b1
var d4 = Delegation{
	DelegatorAddress: "cosmos1qcqgpxeyv6w3vp76e5qg39zf6fqwledswt7l3d",
	Shares:           "1225270.000000000000000000",
	ValidatorAddress: "cosmosvaloper1sjllsnramtg3ewxqwwrwjxfgc4n4ef9u2lcnj0",
}

//the second delegation of b2
var d5 = Delegation{
	DelegatorAddress: "cosmos1qbqdng8wen7lw392slzmjdr8vdeg9ermjxejer",
	Shares:           "1731754.000000000000000000",
	ValidatorAddress: "cosmosvaloper1sjllsnramtg3ewxqwwrwjxfgc4n4ef9u2lcnj0",
}

func TestAddShares(t *testing.T) {

	sum1, err := addShares("1.000000000000000000", "12345678901234567.000000000000000000")

	assert.Equal(t, nil, err)
	assert.Equal(t, "12345678901234568.000000000000000000", sum1)

	// TODO: float64 decimal overflow need to be taken care of
	//sum2, err := addShares("1.00000000000000001", "1.00000000000000001")
	//sum3, err := addShares("1.000000000000000000", "123456789012345678.000000000000000000")
	//assert.Equal(t, nil, err)

	//assert.Equal(t, "2.00000000000000002", sum2)
	//assert.Equal(t, "123456789012345679.000000000000000000", sum3)

}

func TestMergeRecord(t *testing.T) {

	assert := assert.New(t)

	//merge two uncorrelated recods
	m1, err := mergeRecord(b1, d1)

	//prettyJson(m1)

	assert.Nil(err)
	assert.Equal("34141001", m1.Coins[0].Amount)
	assert.Equal("uatom", m1.Coins[0].Denom)
	assert.Equal(1, len(m1.Coins))

	//merge b1 with its first delegations
	m2, err := mergeRecord(b1, d2)
	assert.Nil(err)
	assert.Equal("34141001", m2.Coins[0].Amount)
	assert.Equal("uatom", m2.Coins[0].Denom)
	assert.Equal("1039602.000000000000000000", m2.Coins[1].Amount)
	assert.Equal("shares", m2.Coins[1].Denom)
	assert.Equal(2, len(m2.Coins))

	// merge the second delegation of b1
	m2b, err := mergeRecord(m2, d4)

	assert.Nil(err)
	assert.Equal(2, len(m2b.Coins))
	assert.Equal("34141001", m2b.Coins[0].Amount)
	assert.Equal("uatom", m2b.Coins[0].Denom)
	assert.Equal("2264872.000000000000000000", m2b.Coins[1].Amount)
	assert.Equal("shares", m2b.Coins[1].Denom)

	// merge b2 with its first delegation. b2 does not have coins records
	m3, err := mergeRecord(b2, d3)

	assert.Nil(err)
	assert.Equal("2070000.000000000000000000", m3.Coins[0].Amount)
	assert.Equal("shares", m3.Coins[0].Denom)
	assert.Equal(1, len(m3.Coins))
	// merge b2 with its second delegation.
	m3b, err := mergeRecord(m3, d5)

	assert.Nil(err)
	assert.Equal("3801754.000000000000000000", m3b.Coins[0].Amount)
	assert.Equal("shares", m3b.Coins[0].Denom)
	assert.Equal(1, len(m3b.Coins))

}

func TestMerge(t *testing.T) {
	assert := assert.New(t)

	balances, err := readBalances(*bfile)
	delegations, err2 := readDelegations(*dfile)
	assert.Nil(err)
	assert.Nil(err2)

	m := merge(balances, delegations)
	prettyJson(m)
	assert.Equal(3, len(m))

	// balance are sorted, joined and merged with summed shares
	assert.Equal("cosmos1qajhwf6gs0ezyjmx2t96qjvyxugaefuwpf7qae", m[0].Address)
	assert.Equal("11343922", m[0].Coins[0].Amount)
	assert.Equal("uatom", m[0].Coins[0].Denom)
	assert.Equal(1, len(m[0].Coins))

	assert.Equal("cosmos1qbqdng8wen7lw392slzmjdr8vdeg9ermjxejer", m[1].Address)
	assert.Equal("3801754.000000000000000000", m[1].Coins[0].Amount)
	assert.Equal("shares", m[1].Coins[0].Denom)
	assert.Equal(1, len(m[1].Coins))

	assert.Equal("cosmos1qcqgpxeyv6w3vp76e5qg39zf6fqwledswt7l3d", m[2].Address)

	assert.Equal(2, len(m[2].Coins))
	assert.Equal("34141001", m[2].Coins[0].Amount)
	assert.Equal("uatom", m[2].Coins[0].Denom)
	assert.Equal("2264872.000000000000000000", m[2].Coins[1].Amount)
	assert.Equal("shares", m[2].Coins[1].Denom)

}
func TestSort(t *testing.T) {

	balances, err := readBalances(*bfile)
	delegations, err2 := readDelegations(*dfile)

	assert.Nil(t, err)
	assert.Nil(t, err2)

	//	// sort balances by address
	sort.Sort(balanceSort(balances))

	assert.Equal(t, true, sort.IsSorted(balanceSort(balances)))

	// sort delegations by delegator address
	sort.Sort(delegationSort(delegations))

	assert.Equal(t, true, sort.IsSorted(balanceSort(balances)))

}
func TestReadBalances(t *testing.T) {

	b, err := readBalances(*bfile)

	assert.Nil(t, err)

	assert.Equal(t, 3, len(b))

}

func TestReadDelegations(t *testing.T) {

	fmt.Println(*dfile)

	d, err := readDelegations(*dfile)

	assert.Nil(t, err)

	assert.Equal(t, 5, len(d))

}
