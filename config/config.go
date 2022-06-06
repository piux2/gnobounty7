package config

import (
	"io/ioutil"

	"github.com/pelletier/go-toml"
)

var defaultFile = "./config.toml"

type ConfigFile struct {
	FileName string `flag:"file" help:"config.toml"`
}

var DefaultConfigFile = ConfigFile{

	FileName: defaultFile, // file name with path

}

type Config struct {

	// Full path or relative to the parent director of the applicatoin.db, state.db and blockstore.db
	//  path/data/ , make sure "/" is included at the end of the dir
	DBDir string `toml:"db_dir"`

	DBBackend string `toml:"db_backend"`

	AppDB string `toml:"app_bd"`

	ChainID string `toml:"chain_id"`

	PsqlHost     string `toml:"psql_host"`
	PsqlPort     int    `toml:"psql_port"`
	PsqlUser     string `toml:"psql_user"`
	PsqlPassword string `toml:"psql_password"`
	PsqlDBName   string `toml:"psql_dbname"`
	SinkOn       bool   `toml:"sink_on"`
}

// postgres database connection information

var DefaultConfig = &Config{

	DBDir:        "~/.gaiad/data/", // the cosmos data directory is located. "/" at the end is needed.
	DBBackend:    "goleveldb",
	AppDB:        "application", // name only, do not add extention .db
	ChainID:      "cosmos-hub4",
	PsqlHost:     "localhost",
	PsqlPort:     5432,
	PsqlUser:     "app",
	PsqlPassword: "psink",
	PsqlDBName:   "cosmos_hub4",
	SinkOn:       false,
}

func LoadConfig(file ConfigFile) *Config {

	bz, err := ioutil.ReadFile(file.FileName)
	if err != nil {
		// write the default to the file system
		WriteConfig(file.FileName, DefaultConfig)

		return DefaultConfig
	}

	config := Config{}

	err = toml.Unmarshal(bz, &config)
	if err != nil {
		panic(err)
	}
	return &config

}

func WriteConfig(configFile string, config *Config) {
	c, err := toml.Marshal(config)

	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(configFile, c, 0644)

	if err != nil {
		panic(err)
	}

}
