package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

type TallyCredsConf struct {
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
}

type TallyConf struct {
	Hosts map[string]TallyCredsConf `yaml:"hosts"`
}

func main() {
	raw_conf, err := ioutil.ReadFile("tally.yaml")
	if err != nil {
		log.Fatal(err)
	}

	var conf TallyConf
	err = yaml.Unmarshal(raw_conf, &conf)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(conf.Hosts)
}
