package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/jhoblitt/tally/conf"
	"github.com/jhoblitt/tally/op"
	"github.com/jhoblitt/tally/sum"
	"github.com/korovkin/limiter"
)

func sum_cmd(conf conf.TallyConf, host string, creds conf.TallyCredsConf, cmds ...string) {
	fmt.Println("running sum command on host:", host)

	args := append([]string{"-i", host, "-u", creds.User, "-p", creds.Pass}, cmds...)
	cmd := exec.Command(conf.Sum, args...)
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		fmt.Println("could not run command: ", err)
	}
}

func main() {
	var use_op = flag.Bool("op", false, "use op (1password cli) to get credentials")
	flag.Parse()

	c := conf.ParseFile("tally.yaml")

	fmt.Println("path to sum command:", c.Sum)
	fmt.Println("path to bmc blob:", c.BmcBlob)
	fmt.Println("path to bios blob:", c.BiosBlob)

	s := sum.NewSum(c.Sum)
	out, err := s.Command(nil, "-c", "GetBmcInfo", "--file", c.BmcBlob, "--file_only")
	if err != nil {
		fmt.Println("could not run command: ", err)
	}
	bmc_target, err := sum.ParseBmcInfo(string(out))
	if err != nil {
		fmt.Println("could not parse bmc info: ", err)
	}
	fmt.Println("bmc info:", bmc_target)

	limiter := limiter.NewConcurrencyLimiter(10)
	defer limiter.WaitAndClose()

	for host, creds := range c.Hosts {
		host := host
		creds := creds

		// conf file creds take precedence over op creds
		if creds.User == "" || creds.Pass == "" {
			if *use_op {
				item := op.ItemGet(host)
				creds = op.Item2TallyCreds(item)
			} else {
				fmt.Println("no credentials for host:", host)
				continue
			}
		}

		limiter.Execute(func() {
			//log := log.New(os.Stdout, host, log.LstdFlags)
			log := log.New(os.Stdout, host+": ", 0)

			out, _ := s.Command(&creds, "-i", host, "-c", "GetBmcInfo")
			if err != nil {
				log.Println("could not run command: ", err)
				return
			}

			bmc_current, err := sum.ParseBmcInfo(string(out))
			if err != nil {
				log.Println("could not parse bmc info: ", err)
				return
			}

			log.Println("bmc info:", bmc_current)

			if bmc_current.Type != bmc_target.Type {
				log.Println("incompatible bmc types, skipping")
				return
			}

			if bmc_current.Version == bmc_target.Version {
				log.Println("bmc version already in sync")
				return
			}

			log.Println("bmc firmware will be upgraded")

			out, _ = s.Command(&creds, "-i", host, "-c", "UpdateBMC", "--file", c.BmcBlob)
			if err != nil {
				log.Println("could not run command: ", err)
			}
			log.Println(string(out))

			log.Println("bmc firmware upgrade complete")

			/*
				sum_cmd(c, host, creds, "-c", "UpdateBMC", "--file", c.BmcBlob)
				sum_cmd(c, host, creds, "-c", "UpdateBios", "--file", c.BiosBlob, "--reboot", "--preserve_setting", "--post_complete")
			*/
		})
	}
}
