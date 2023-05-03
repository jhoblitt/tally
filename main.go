package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/korovkin/limiter"
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

func sum_cmd(conf TallyConf, host string, creds TallyCredsConf, cmds ...string) {
	fmt.Println("running sum command on host:", host)

	args := append([]string{"-i", host, "-u", creds.User, "-p", creds.Pass}, cmds...)
	cmd := exec.Command(conf.Sum, args...)
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		fmt.Println("could not run command: ", err)
	}
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

	limiter := limiter.NewConcurrencyLimiter(10)
	defer limiter.WaitAndClose()

	for host, creds := range conf.Hosts {
		host := host
		creds := creds

		limiter.Execute(func() {
			sum_cmd(conf, host, creds, "-c", "GetBmcInfo")
			sum_cmd(conf, host, creds, "-c", "GetBiosInfo")
			sum_cmd(conf, host, creds, "-c", "UpdateBMC", "--file", conf.BmcBlob)
			sum_cmd(conf, host, creds, "-c", "UpdateBios", "--file", conf.BiosBlob, "--reboot", "--preserve_setting", "--post_complete")
		})
	}

	limiter.WaitAndClose()
}
