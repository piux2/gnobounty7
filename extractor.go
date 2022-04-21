package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sort"
	"strconv"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/errors"
)

//Assumption:
// Exported states are accurate and correct
// Balances contain all the addresses, and addresseses are unique in balances.
// We igonre the delegator addresseses that do not appear in Balances address.

func main() {

	cmd := command.NewStdCommand()
	exec := os.Args[0]
	args := os.Args[1:]
	err := runMain(cmd, exec, args)

	if err != nil {

		cmd.ErrPrintfln("%s", err.Error())

		return

	}

}

type AppItem = command.AppItem
type AppList = command.AppList

var mainApps AppList = []AppItem{

	{App: mergeApp, Name: "merge", Desc: "merge balances and delegations from exported genesis states", Defaults: DefaultMergeOptions},
}

func runMain(cmd *command.Command, exec string, args []string) error {

	if len(args) == 0 || args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		cmd.Println("available subcommands: ")
		for _, appItem := range mainApps {

			cmd.Printf("  %s - %s\n", appItem.Name, appItem.Desc)

			return nil
		}

	}

	for _, appItem := range mainApps {

		if appItem.Name == args[0] {

			err := cmd.Run(appItem.App, args[1:], appItem.Defaults)
			return err
		}

	}

	return errors.New("unknown command: " + args[0])

}

// mergeApp section

type mergeOptions struct {
	BalanceFile    string `flag:"b" help:"balances.js"`
	DelegationFile string `flag:"d" help:"delegations.json"`
}

var DefaultMergeOptions = mergeOptions{

	BalanceFile:    "", // required
	DelegationFile: "", // required

}

type Delegation struct {
	DelegatorAddress string `json:"delegator_address"`
	Shares           string `json:"shares"`
	ValidatorAddress string `json:"validator_address"`
}

type Balance struct {
	Address string `json:"address"`
	Coins   []Coin `json:"coins"`
}
type Coin struct {
	Amount string `json:"amount"`
	Denom  string `json:"denom"`
}

func mergeApp(cmd *command.Command, args []string, iopts interface{}) error {

	opts := iopts.(mergeOptions)

	if len(args) != 0 {

		cmd.ErrPrintfln("Usage: extractor merge --b balances.json --d delegations.json\n")

	}

	if opts.BalanceFile == "" {
		cmd.ErrPrintfln("Usage: extractor merge --b balances.json --d delegations.json\n")

		return errors.New("--b file not specified\n")

	}

	if opts.DelegationFile == "" {
		cmd.ErrPrintfln("Usage: extractor merge --b balances.json --d delegations.json\n")

		return errors.New("--d file not specified\n")
	}

	bfname := opts.BalanceFile
	dfname := opts.DelegationFile

	b := readBalances(bfname)
	d := readDelegations(dfname)
	m := merge(b, d)

	v, err := json.MarshalIndent(m, "", " ")

	if err != nil {

		return err

	}
	cmd.Println(string(v))

	return nil

}

// Merge balances and delegation by address and calculate total uatom

func merge(balances []Balance, delegations []Delegation) []Balance {

	// quick sort O(nlogn) to O(n^2)
	// sort balances by address
	sort.Sort(balanceSort(balances))

	// sort delegations by delegator address
	sort.Sort(delegationSort(delegations))

	lenBal := len(balances)
	lenDel := len(delegations)

	indexB := 0
	indexD := 0
	// loop through both balance and delegation list at the same time

	for indexB < lenBal && indexD < lenDel {

		b := balances[indexB]
		d := delegations[indexD]
		bAddress := b.Address
		dAddress := d.DelegatorAddress

		if bAddress < dAddress {

			indexB++

			continue

		} else if bAddress > dAddress {

			indexD++

			continue

		} else if bAddress == dAddress {

			newBalance, err := mergeRecord(b, d)

			if err != nil {

				errors.New("failed to merge", err.Error())
			}

			balances[indexB] = newBalance

			indexD++
			continue
		}

	}

	return balances
}

func mergeRecord(b Balance, d Delegation) (Balance, error) {

	// loop through []coins and if there is "share" denom, sum it,
	// otherwise add a Coin{amount: shares, denom: "share"}

	//if the balance has the "shares" coin, we sum them, otherwise we add a new "shares" coin in the Coins[]
	hasShares := false

	for i, bcoin := range b.Coins {

		if bcoin.Denom == "shares" {

			hasShares = true

			sum, err := addShares(bcoin.Amount, d.Shares)

			if err != nil {

				return b, errors.New("can not add %s and %s ", bcoin.Amount, d.Shares)

			}
			// share is a float64 with 18 decical

			b.Coins[i].Amount = sum

		}
	}

	if hasShares == false {

		b.Coins = append(b.Coins, Coin{Amount: d.Shares, Denom: "shares"})

	}

	return b, nil

}

// TODO: There is an issue adding two numbers with large decical. see TestAddShares
// "1234567890.100000000000000000", "0.200000000000000000"
func addShares(a string, b string) (string, error) {

	afloat, err := strconv.ParseFloat(a, 64)

	if err != nil {

		return "", errors.New(a, err.Error())

	}

	bfloat, err := strconv.ParseFloat(b, 64)

	if err != nil {

		return "", errors.New(b, err.Error())

	}

	sum := afloat + bfloat
	s := strconv.FormatFloat(sum, 'f', 18, 64)

	return s, err

}

func readBalances(bfname string) []Balance {

	bf, err := ioutil.ReadFile(bfname)

	if err != nil {

		panic(err)

	}

	balances := []Balance{}
	json.Unmarshal(bf, &balances)

	return balances

}

func readDelegations(dfname string) []Delegation {

	df, err := ioutil.ReadFile(dfname)

	if err != nil {

		panic(err)

	}
	delegations := []Delegation{}

	json.Unmarshal(df, &delegations)

	return delegations
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
