package conf

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

type TallyCredsConf struct {
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
}

type TallyConf struct {
	Sum      string                    `yaml:"sum"`
	BiosBlob string                    `yaml:"bios_blob"`
	BmcBlob  string                    `yaml:"bmc_blob"`
	Hosts    map[string]TallyCredsConf `yaml:"hosts"`
}

func ParseFile(conf_file string) TallyConf {
	raw_conf, err := ioutil.ReadFile(conf_file)
	if err != nil {
		log.Fatal(err)
	}

	var conf TallyConf
	err = yaml.Unmarshal(raw_conf, &conf)
	if err != nil {
		log.Fatal(err)
	}

	return conf
}
