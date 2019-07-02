package conf

import (
	"encoding/json"
	"io/ioutil"
)

type Redis struct {
	Network string `json:"network"`
	Address string `json:"address"`
}

type RethinDB struct {
	Address  string `json:"dns_rethinkdb"`
	Database string `json:"database_rethinkdb"`
}

type conf struct {
	DnsOracle string   `json:"dns_oracle"`
	Redis     Redis    `json:"redis"`
	RethinDB  RethinDB `json:"rethinkdb"`
}

var Conf conf

func Load(file string) error {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bytes, &Conf)
	if err != nil {
		return err
	}
	return nil
}
