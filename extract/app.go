package extract

import (
	"github.com/gnolang/gno/pkgs/command"
	"github.com/piux2/gnobounty7/config"
	"github.com/piux2/gnobounty7/sink"
)

const STATE_HEIGHT = int64(10562840)

func ExtractApp(cmd *command.Command, args []string, iopts interface{}) error {

	cf := iopts.(config.ConfigFile)

	var c *config.Config

	if cf.FileName == "" {

		c = config.LoadConfig(config.DefaultConfigFile)

	} else {

		c = config.LoadConfig(cf)
	}

	s := sink.NewPsqlSink(c)

	ProfileAppDB(s)

	ProfileBlockstoreDB(s)
	ProfileStateDB(s)

	//////////
	height := STATE_HEIGHT

	err := ExtractValidators(s, height, true)

	if err != nil {

		panic(err)
	}

	err = ExtractVotes(s, true)
	if err != nil {

		panic(err)
	}

	return nil
}
