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

type HostUpdate struct {
	Name          string
	TargetBmcInfo *sum.SumBMC
	Creds         *conf.TallyCredsConf
	Conf          *conf.TallyConf
	Sum           *sum.Sum
	Logger        *log.Logger
}

func bmc_update(host HostUpdate) error {
	l := host.Logger

	l.Println("checking bmc info")

	out, err := host.Sum.Command(host.Creds, "-i", host.Name, "-c", "GetBmcInfo")
	if err != nil {
		return fmt.Errorf("could not run command: %w", err)
	}
	bmc_current, err := sum.ParseBmcInfo(string(out))
	if err != nil {
		return fmt.Errorf("could not parse bmc info: %w", err)
	}

	l.Println("bmc info:", bmc_current)

	if bmc_current.Type != host.TargetBmcInfo.Type {
		return fmt.Errorf("incompatible bmc types")
	}
	if bmc_current.Version == host.TargetBmcInfo.Version {
		l.Println("bmc version already in sync")
		return nil
	}

	l.Println("bmc firmware will be upgraded")

	out, _ = host.Sum.Command(host.Creds, "-i", host.Name, "-c", "UpdateBMC", "--file", host.Conf.BmcBlob)
	if err != nil {
		return fmt.Errorf("could not run command: %w", err)
	}
	l.Println(string(out))

	l.Println("bmc firmware upgrade complete")

	return nil
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
		log.Fatal("could not run command: ", err)
	}
	bmc_target, err := sum.ParseBmcInfo(string(out))
	if err != nil {
		log.Fatal("could not parse bmc info: ", err)
	}
	fmt.Println("bmc info:", bmc_target)

	limiter := limiter.NewConcurrencyLimiter(10)
	defer limiter.WaitAndClose()

	for host, creds := range c.Hosts {
		host := host
		creds := creds
		// prefix log messages with host name
		l := log.New(os.Stdout, host+": ", 0)

		// conf file creds take precedence over op creds
		if creds.User == "" || creds.Pass == "" {
			if *use_op {
				item := op.ItemGet(host)
				creds = op.Item2TallyCreds(item)
			} else {
				l.Println("no credentials for host:", host)
				continue
			}
		}

		limiter.Execute(func() {
			host := HostUpdate{
				Name:          host,
				TargetBmcInfo: &bmc_target,
				Creds:         &creds,
				Conf:          &c,
				Sum:           s,
				Logger:        l,
			}

			err = bmc_update(host)
			if err != nil {
				l.Println("could not update bmc:", err)
				return
			}

			/*
				sum_cmd(c, host, creds, "-c", "UpdateBMC", "--file", c.BmcBlob)
				sum_cmd(c, host, creds, "-c", "UpdateBios", "--file", c.BiosBlob, "--reboot", "--preserve_setting", "--post_complete")
			*/
		})
	}
}
