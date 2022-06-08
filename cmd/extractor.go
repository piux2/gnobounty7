package main

import (
	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/errors"
	"github.com/piux2/gnobounty7/config"
	"github.com/piux2/gnobounty7/extract"
	"github.com/piux2/gnobounty7/merge"
	"os"
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

//TODO: add --profile only flag for extract app

var mainApps AppList = []AppItem{

	{App: merge.MergeApp, Name: "merge", Desc: "merge balances and delegations from exported genesis states", Defaults: merge.DefaultMergeOptions},
	{App: extract.ExtractApp, Name: "extract", Desc: "extract votes, proposal and transaction information", Defaults: config.DefaultConfigFile},
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

	return errors.New("unknown command: %s\n try: extractor help", args[0])

}
