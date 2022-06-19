package merge

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"sort"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/errors"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/piux2/gnobounty7/extract"
)

// mergeApp section

type mergeOptions struct {
	BalanceFile    string `flag:"b" help:"balances.js"`
	DelegationFile string `flag:"d" help:"delegations.json"`
	ValidatorFile  string `flag:"val" help:"validators.json"`
	VoteFile       string `flag:"vote" help:"votes.json"`
}

var DefaultMergeOptions = mergeOptions{

	BalanceFile:    "", // required
	DelegationFile: "", // required
	ValidatorFile:  "", // required
	VoteFile:       "", // required

}

type Delegation struct {
	DelegatorAddress string `json:"delegator_address"`
	Shares           string `json:"shares"`
	ValidatorAddress string `json:"validator_address"`
}

type Balance struct {
	Address string `json:"address"`
	Coins   []Coin `json:"coins"`
	Vote    string `json:"vote"`
}
type Coin struct {
	Amount string `json:"amount"`
	Denom  string `json:"denom"`
}

func MergeApp(cmd *command.Command, args []string, iopts interface{}) error {

	opts := iopts.(mergeOptions)

	if len(args) != 0 {

		cmd.ErrPrintfln("Usage: extractor merge --b balances.json --d delegations.json --val validators.json\n")

	}

	if opts.BalanceFile == "" {
		cmd.ErrPrintfln("Usage: extractor merge --b balances.json --d delegations.json --val validators.json --vote votes.json\n")

		return errors.New("--b file not specified\n")

	}

	if opts.DelegationFile == "" {
		cmd.ErrPrintfln("Usage: extractor merge --b balances.json --d delegations.json --val validators.json --vote votes.json\n")

		return errors.New("--d file not specified\n")
	}

	if opts.ValidatorFile == "" {
		cmd.ErrPrintfln("Usage: extractor merge --b balances.json --d delegations.json --val validators.json --vote votes.json\n")

		return errors.New("--val file not specified\n")
	}

	if opts.VoteFile == "" {
		cmd.ErrPrintfln("Usage: extractor merge --b balances.json --d delegations.json --val validators.json --vote votes.json\n")

		return errors.New("--vote file not specified\n")
	}

	bfname := opts.BalanceFile
	dfname := opts.DelegationFile
	vfname := opts.ValidatorFile
	votefname := opts.VoteFile

	b, err := readBalances(bfname)

	if err != nil {

		return errors.New("can not parse balance file", err.Error())
	}

	d, err2 := readDelegations(dfname)

	if err2 != nil {

		return errors.New("can not parse delegations file", err2.Error())
	}

	v, err3 := readValidators(vfname)

	if err3 != nil {

		return errors.New("can not parse validators file", err3.Error())
	}

	vote, err4 := readVotes(votefname)

	if err4 != nil {

		return errors.New("can not parse votes file", err4.Error())
	}

	m := mergeBalanceAndDelegations(b, d, v)
	m = mergeBalanceAndVotes(m, vote)

	prettyJson(m)

	return nil

}

func mergeBalanceAndDelegations(balances []Balance, delegations []Delegation, validators map[string]extract.ValExtract) []Balance {

	// quick sort O(nlogn) to O(n^2)
	// sort balances by address
	sort.Sort(balanceSort(balances))

	// sort delegations by delegator address
	sort.Sort(delegationSort(delegations))

	lenBal := len(balances)
	lenDel := len(delegations)

	indexB := 0
	indexD := 0

	// loop through  balance delegation and vote list at the same time

	for indexB < lenBal && indexD < lenDel {

		b := balances[indexB]
		d := delegations[indexD]

		bAddress := b.Address
		dAddress := d.DelegatorAddress

		if bAddress < dAddress {

			indexB++

			continue

		} else if bAddress > dAddress {

			if indexB == 0 { // add unmatched delegation before first balance address

				newBalance, nextD := addUnmatchedDelegations(delegations, indexD, validators)
				indexD = nextD
				balances = append(balances, newBalance)
				continue

			}

			lastBalance := balances[indexB-1]

			// this condition means the delgation address never matched with
			// the delegation address was skipt we need to added create an new balance entry and
			// append it to the end of balances array
			if lastBalance.Address != dAddress {
				newBalance, nextIndex := addUnmatchedDelegations(delegations, indexD, validators)
				indexD = nextIndex
				balances = append(balances, newBalance)
				continue

			}

			indexD++

			continue

		} else if bAddress == dAddress {

			newBalance, err := mergeRecord(b, d, validators)

			// continue without interrupting merging process
			if err != nil {

				fmt.Println("failed to merge", err.Error())
			}

			balances[indexB] = newBalance

			indexD++
			continue
		}

	}
	// add all unmatched delegation after the last balance address

	for indexD < lenDel {

		newBalance, nextD := addUnmatchedDelegations(delegations, indexD, validators)
		indexD = nextD
		balances = append(balances, newBalance)

	}

	return balances

}
func addUnmatchedDelegations(delegations []Delegation, indexD int, validators map[string]extract.ValExtract) (Balance, int) {

	d := delegations[indexD]
	b := Balance{Address: d.DelegatorAddress, Coins: []Coin{}}

	newBalance, err := mergeRecord(b, d, validators)
	// continue without interrupting merging process
	if err != nil {

		fmt.Println("failed to merge", err.Error())
	}

	// merge all delegations of a address missed in balance array
	// since delegations are sorted by addresses, we can assume all delegation from
	// the same delegator address are next to each other.

	indexD++

	if indexD == len(delegations) {

		return newBalance, indexD

	}

	nextD := delegations[indexD]

	for b.Address == nextD.DelegatorAddress {

		newBalance, err = mergeRecord(newBalance, nextD, validators)
		// continue without interrupting merging process
		if err != nil {

			fmt.Println("failed to merge", err.Error())
		}
		indexD++
		nextD = delegations[indexD]
	}

	return newBalance, indexD

}
func mergeBalanceAndVotes(balances []Balance, votes []extract.VoteExtract) []Balance {

	// quick sort O(nlogn) to O(n^2)
	// sort balances by address
	sort.Sort(balanceSort(balances))

	// sort votes by voter address
	sort.Sort(voteSort(votes))

	lenBal := len(balances)
	lenVote := len(votes)

	indexB := 0
	indexVote := 0
	// loop through  balance and vote list at the same time

	for indexB < lenBal && indexVote < lenVote {

		b := balances[indexB]
		vote := votes[indexVote]

		bAddress := b.Address
		voterAddress := vote.Sender

		if bAddress == voterAddress {

			b.Vote = vote.Option
			//update the updated balance entity.
			balances[indexB] = b

			indexB++
			indexVote++

		} else if bAddress > voterAddress {

			indexVote++

		} else if bAddress < voterAddress {

			indexB++

		}

	}

	return balances

}

func mergeRecord(b Balance, d Delegation, validators map[string]extract.ValExtract) (Balance, error) {

	// if we can not join records by address, ignore it.do nothing
	if b.Address != d.DelegatorAddress {

		return b, nil
	}

	validator, ok := validators[d.ValidatorAddress]

	if ok == false {

		return b, fmt.Errorf("%s does not exist in the validator map", d.ValidatorAddress)
	}

	// loop through []coins and if there is "duatom" denom, sum it,
	// otherwise add a Coin{amount: duatom, denom: "duatom"}
	// NOTE: duatom is delegation shares coverted back to token in uatom.
	// we might not get 1:1 from share back to uatom because the validator might got slashed after
	// a user delegated the tokens.

	//if the balance has the "duatom" coin, we sum them, otherwise we add a new "duatom" coin in the Coins[]
	hasDuatom := false

	for i, bcoin := range b.Coins {

		if bcoin.Denom == "duatom" {

			hasDuatom = true

			sum, err := addShares(bcoin.Amount, d.Shares, validator)

			if err != nil {

				return b, errors.New("can not add %s and %s ", bcoin.Amount, d.Shares)

			}
			// share is a float64 with 18 decical

			b.Coins[i].Amount = sum

		}
	}

	if hasDuatom == false {

		sum, err := addShares("0", d.Shares, validator)
		if err != nil {

			return b, errors.New("can not add %s and %s ", "0", d.Shares)

		}

		b.Coins = append(b.Coins, Coin{Amount: sum, Denom: "duatom"})

	}

	return b, nil

}

func addShares(uatom string, shares string, v extract.ValExtract) (string, error) {

	a, err := types.NewDecFromStr(uatom)

	if err != nil {

		return "", err
	}

	s, err2 := types.NewDecFromStr(shares)

	if err2 != nil {

		return "", err2

	}
	//tokens from validators delegator_share x validator.tokens/validator.shares
	c := (s.MulInt(v.Tokens)).Quo(v.Shares)

	return a.Add(c).String(), err

}

func readBalances(bfname string) ([]Balance, error) {

	bf, err := ioutil.ReadFile(bfname)

	if err != nil {

		return nil, errors.New("can not read balance file", bfname, err.Error())

	}

	balances := []Balance{}
	err = json.Unmarshal(bf, &balances)

	if err != nil {

		return nil, errors.New("can not parse balance file", bfname, err.Error())

	}

	return balances, nil

}

func readDelegations(dfname string) ([]Delegation, error) {

	df, err := ioutil.ReadFile(dfname)

	if err != nil {

		return nil, errors.New("can not read delegation file", dfname, err.Error())

	}
	delegations := []Delegation{}

	err = json.Unmarshal(df, &delegations)

	if err != nil {
		return nil, errors.New("can not parse delegation json file", dfname, err.Error())
	}

	return delegations, nil
}

func readValidators(vfname string) (map[string]extract.ValExtract, error) {

	vf, err := ioutil.ReadFile(vfname)

	if err != nil {

		return nil, errors.New("can not read validator file", vfname, err.Error())

	}
	validators := []extract.ValExtract{}

	err = json.Unmarshal(vf, &validators)

	if err != nil {
		return nil, errors.New("can not parse validator json file", vfname, err.Error())
	}

	vmap := map[string]extract.ValExtract{}
	for _, v := range validators {

		vmap[v.ValAddress] = v
	}

	return vmap, nil

}

func readVotes(fname string) ([]extract.VoteExtract, error) {

	f, err := ioutil.ReadFile(fname)

	if err != nil {

		return nil, errors.New("can not read votes file", fname, err.Error())

	}
	votes := []extract.VoteExtract{}

	err = json.Unmarshal(f, &votes)

	if err != nil {
		return nil, errors.New("can not parse votes json file", fname, err.Error())
	}

	return votes, nil
}

// balance sort by Address
type balanceSort []Balance

func (a balanceSort) Len() int { return len(a) }

func (a balanceSort) Less(i, j int) bool {

	return a[i].Address < a[j].Address

}

func (a balanceSort) Swap(i, j int) {

	a[i], a[j] = a[j], a[i]

}

// delegation sorter by  delegation address
type delegationSort []Delegation

func (d delegationSort) Len() int { return len(d) }

func (d delegationSort) Less(i, j int) bool {

	return d[i].DelegatorAddress < d[j].DelegatorAddress

}

func (d delegationSort) Swap(i, j int) {

	d[i], d[j] = d[j], d[i]
}

// balance sort by Address
type voteSort []extract.VoteExtract

func (a voteSort) Len() int { return len(a) }

func (a voteSort) Less(i, j int) bool {

	return a[i].Sender < a[j].Sender

}

func (a voteSort) Swap(i, j int) {

	a[i], a[j] = a[j], a[i]

}

// visual review of the merged list.
func prettyJson(a interface{}) error {

	v, err := json.MarshalIndent(a, "", " ")

	if err != nil {
		return err
	}

	fmt.Println(string(v))

	return nil
}
