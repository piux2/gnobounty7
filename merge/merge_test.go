package merge

import (
	"flag"
	"fmt"
	"sort"
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/piux2/gnobounty7/extract"
	"github.com/stretchr/testify/assert"
)

var bfile *string
var dfile *string
var vfile *string

func init() {

	bfile = flag.String("b", "", "balances.json file name")
	dfile = flag.String("d", "", "delegations.json file name")
	vfile = flag.String("val", "", "validators.json file name")

	if *bfile == "" || *dfile == "" || *vfile == "" {

		*bfile = "test_balances.json"
		*dfile = "test_delegations.json"
		*vfile = "test_validators.json"

		fmt.Println("\nyou can load your own data to test with the flag options --b balancefile.json, --d delegationsfile.json, --v validatorsfile.json ")

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

var vshare1, _ = types.NewDecFromStr("883435145.637346060773787067")
var vratio1, _ = types.NewDecFromStr("0.999904372564569963")

var vshare2, _ = types.NewDecFromStr("4373508.000000000000000000")
var vratio2, _ = types.NewDecFromStr("1.000000000000000000")

var vshare3, _ = types.NewDecFromStr("214361133.454145305958021670")
var vratio3, _ = types.NewDecFromStr("0.998800675990079607")

var vshare4, _ = types.NewDecFromStr("500005.956625917197747696")
var vratio4, _ = types.NewDecFromStr("0.950004682354975250")

var validators = map[string]extract.ValExtract{

	"cosmosvaloper1n229vhepft6wnkt5tjpwmxdmcnwz55jv3vp7ed": extract.ValExtract{
		"cosmosvaloper1n229vhepft6wnkt5tjpwmxdmcnwz55jv3vp7ed",
		"cosmos1qtxec3ggeuwnca9mmngw7vf6ctw54cppusml9r",
		"test validator 1",
		10562840,
		types.NewInt(int64(883350665)),
		vshare1,
		vratio1,
	},
	"cosmosvaloper1tfk30mq5vgqjdly92kkhhq3raev2hnz6eete3": extract.ValExtract{
		"cosmosvaloper1tfk30mq5vgqjdly92kkhhq3raev2hnz6eete3",
		"cosmos1998928nfs697ep5d825y5jah0nq9zrtd2ms37d",
		"test validator 3",
		10562840,
		types.NewInt(int64(214104045)),
		vshare3,
		vratio3,
	},
	"cosmosvaloper196ax4vc0lwpxndu9dyhvca7jhxp70rmcvrj90c": extract.ValExtract{

		"cosmosvaloper196ax4vc0lwpxndu9dyhvca7jhxp70rmcvrj90c",
		"cosmos19v9ej55ataqrfl39v83pf4e0dm69u89rkuasx5",
		"test validator 4",
		10562840,
		types.NewInt(int64(475008)),
		vshare4,
		vratio4,
	},
	"cosmosvaloper1sjllsnramtg3ewxqwwrwjxfgc4n4ef9u2lcnj0": extract.ValExtract{
		"cosmosvaloper1sjllsnramtg3ewxqwwrwjxfgc4n4ef9u2lcnj0",
		"cosmos1zmr5mglwkkru7m3q8sxcg66gxr508v6hltcj9m",
		"test validator 2",
		10562840,
		types.NewInt(int64(4373508)),
		vshare2,
		vratio2,
	},
}

func TestAddShares(t *testing.T) {
	//"tokens": 883350665,
	//"shares": "883435145.637346060773787067",
	v1, _ := validators["cosmosvaloper1n229vhepft6wnkt5tjpwmxdmcnwz55jv3vp7ed"]
	sum1, err := addShares("1.000000000000000000", "12345678901234567.000000000000000000", v1)

	assert.Equal(t, nil, err)
	assert.Equal(t, "12344498315622600.219489215357475304", sum1)

	sum2, err := addShares("1.000000000000000001", "1.000000000000000001", v1)
	assert.Equal(t, nil, err)
	assert.Equal(t, "1.999904372564569965", sum2)

	sum3, err := addShares("1.000000000000000000", "123456789012345678.000000000000000000", v1)
	assert.Equal(t, nil, err)

	assert.Equal(t, "123444983156226001.194127134091312743", sum3)

	v2, _ := validators["cosmosvaloper1sjllsnramtg3ewxqwwrwjxfgc4n4ef9u2lcnj0"]

	sum4, err := addShares("1.000000000000000000", "12345678901234567.000000000000000000", v2)

	assert.Equal(t, nil, err)
	assert.Equal(t, "12345678901234568.000000000000000000", sum4)

	sum5, err := addShares("1.000000000000000001", "1.000000000000000001", v2)
	assert.Equal(t, nil, err)
	assert.Equal(t, "2.000000000000000002", sum5)

	sum6, err := addShares("1.000000000000000000", "123456789012345678.000000000000000000", v2)
	assert.Equal(t, nil, err)
	assert.Equal(t, "123456789012345679.000000000000000000", sum6)

}

func TestMergeRecord(t *testing.T) {

	assert := assert.New(t)

	//merge two uncorrelated recods
	m1, err := mergeRecord(b1, d1, validators)

	//prettyJson(m1)

	assert.Nil(err)
	assert.Equal("34141001", m1.Coins[0].Amount)
	assert.Equal("uatom", m1.Coins[0].Denom)
	assert.Equal(1, len(m1.Coins))

	//merge b1 with its first delegations
	m2, err := mergeRecord(b1, d2, validators)
	assert.Nil(err)
	assert.Equal("34141001", m2.Coins[0].Amount)
	assert.Equal("uatom", m2.Coins[0].Denom)
	assert.Equal("1038355.180360638740055647", m2.Coins[1].Amount)
	assert.Equal("duatom", m2.Coins[1].Denom)
	assert.Equal(2, len(m2.Coins))

	// merge the second delegation of b1
	m2b, err := mergeRecord(m2, d4, validators)

	assert.Nil(err)
	assert.Equal(2, len(m2b.Coins))
	assert.Equal("34141001", m2b.Coins[0].Amount)
	assert.Equal("uatom", m2b.Coins[0].Denom)
	assert.Equal("2263625.180360638740055647", m2b.Coins[1].Amount)
	assert.Equal("duatom", m2b.Coins[1].Denom)

	// merge b2 with its first delegation. b2 does not have coins records
	m3, err := mergeRecord(b2, d3, validators)

	assert.Nil(err)
	assert.Equal("1966509.692474798768082056", m3.Coins[0].Amount)
	assert.Equal("duatom", m3.Coins[0].Denom)
	assert.Equal(1, len(m3.Coins))
	// merge b2 with its second delegation.
	m3b, err := mergeRecord(m3, d5, validators)

	assert.Nil(err)
	assert.Equal("3698263.692474798768082056", m3b.Coins[0].Amount)
	assert.Equal("duatom", m3b.Coins[0].Denom)
	assert.Equal(1, len(m3b.Coins))

}

func TestMerge(t *testing.T) {
	assert := assert.New(t)

	balances, err := readBalances(*bfile)
	delegations, err2 := readDelegations(*dfile)
	validators, err3 := readValidators(*vfile)
	assert.Nil(err)
	assert.Nil(err2)
	assert.Nil(err3)

	m := merge(balances, delegations, validators)
	prettyJson(m)
	assert.Equal(3, len(m))

	// balance are sorted, joined and merged with summed shares
	assert.Equal("cosmos1qajhwf6gs0ezyjmx2t96qjvyxugaefuwpf7qae", m[0].Address)
	assert.Equal("11343922", m[0].Coins[0].Amount)
	assert.Equal("uatom", m[0].Coins[0].Denom)
	assert.Equal(1, len(m[0].Coins))

	assert.Equal("cosmos1qbqdng8wen7lw392slzmjdr8vdeg9ermjxejer", m[1].Address)
	assert.Equal("3698263.692474798768082056", m[1].Coins[0].Amount)
	assert.Equal("duatom", m[1].Coins[0].Denom)
	assert.Equal(1, len(m[1].Coins))

	assert.Equal("cosmos1qcqgpxeyv6w3vp76e5qg39zf6fqwledswt7l3d", m[2].Address)

	assert.Equal(2, len(m[2].Coins))
	assert.Equal("34141001", m[2].Coins[0].Amount)
	assert.Equal("uatom", m[2].Coins[0].Denom)
	assert.Equal("2263625.180360638740055647", m[2].Coins[1].Amount)
	assert.Equal("duatom", m[2].Coins[1].Denom)

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

func TestReadValidators(t *testing.T) {

	fmt.Println(*vfile)
	v, err := readValidators(*vfile)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(v))

}
