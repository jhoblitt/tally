package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

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

	fmt.Println("path to sum command:", conf.Sum)
	fmt.Println("path to bmc blob:", conf.BmcBlob)
	fmt.Println("path to bios blob:", conf.BiosBlob)

	for host, creds := range conf.Hosts {
		fmt.Println("host:", host)

		//cmd := exec.Command(conf.Sum, "-i", host, "-u", creds.User, "-p", creds.Pass, "-c", "GetBmcInfo")
		cmd := exec.Command(conf.Sum, "-i", host, "-u", creds.User, "-p", creds.Pass, "-c", "UpdateBMC", "--file", conf.BmcBlob)
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil {
			fmt.Println("could not run command: ", err)
		}
	}
}
