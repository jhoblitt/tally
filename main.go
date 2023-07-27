package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jhoblitt/tally/conf"
	"github.com/jhoblitt/tally/op"
	"github.com/jhoblitt/tally/sum"
	"github.com/korovkin/limiter"
)

type HostUpdate struct {
	Name           string
	TargetBmcInfo  *sum.SumBmcInfo
	TargetBiosInfo *sum.SumBiosInfo
	Creds          *conf.TallyCredsConf
	Conf           *conf.TallyConf
	Sum            *sum.Sum
	Logger         *log.Logger
	Noop           bool
}

func bmc_update(host HostUpdate) error {
	l := host.Logger

	l.Println("checking bmc info")

	out, err := host.Sum.Command(host.Creds, "-i", host.Name, "-c", "GetBmcInfo")
	if err != nil {
		return fmt.Errorf("could not run sum command: %w", err)
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

	if host.Noop {
		l.Println("bmc firmware would be upgraded (noop)")
		return nil
	}

	l.Println("bmc firmware will be upgraded")

	out, _ = host.Sum.Command(host.Creds, "-i", host.Name, "-c", "UpdateBMC", "--file", host.Conf.BmcBlob)
	if err != nil {
		return fmt.Errorf("could not run sum command: %w", err)
	}
	l.Println(string(out))

	l.Println("bmc firmware upgrade complete")

	return nil
}

func bios_update(host HostUpdate) error {
	l := host.Logger

	l.Println("checking bios info")

	out, err := host.Sum.Command(host.Creds, "-i", host.Name, "-c", "GetBiosInfo")
	if err != nil {
		return fmt.Errorf("could not run sum command: %w", err)
	}
	bios_current, err := sum.ParseBiosInfo(string(out))
	if err != nil {
		return fmt.Errorf("could not parse bios info: %w", err)
	}

	l.Println("bios info:", bios_current)

	if bios_current.BoardID != host.TargetBiosInfo.BoardID {
		return fmt.Errorf("incompatible Board Ids")
	}
	if bios_current.BuildDate == host.TargetBiosInfo.BuildDate {
		l.Println("bios build date already in sync")
		return nil
	}

	if host.Noop {
		l.Println("bios would be upgraded (noop)")
		return nil
	}

	l.Println("bios will be upgraded")

	out, _ = host.Sum.Command(host.Creds, "-i", host.Name, "-c", "UpdateBIOS", "--file", host.Conf.BiosBlob, "--reboot", "--preserve_setting", "--post_complete")
	if err != nil {
		return fmt.Errorf("could not run sum command: %w", err)
	}
	l.Println(string(out))

	l.Println("bios upgrade complete")

	return nil
}

func main() {
	use_op := flag.Bool("op", false, "use op (1password cli) to get credentials")
	noop := flag.Bool("noop", false, "do not apply firmware updates")
	parallelism := flag.Int("p", 10, "number of hosts to update in parallel")
	conf_file := flag.String("conf", "tally.yaml", "tally configuration file (YAML)")
	flag.Parse()

	c := conf.ParseFile(*conf_file)

	fmt.Println("path to sum command:", c.Sum)
	fmt.Println("path to bmc blob:", c.BmcBlob)
	fmt.Println("path to bios blob:", c.BiosBlob)

	s := sum.NewSum(c.Sum)
	out, err := s.Command(nil, "-c", "GetBmcInfo", "--file", c.BmcBlob, "--file_only")
	if err != nil {
		log.Fatal("could not run sum command: ", err)
	}
	bmc_target, err := sum.ParseBmcInfo(string(out))
	if err != nil {
		log.Fatal("could not parse bmc info: ", err)
	}
	fmt.Println("bmc info:", bmc_target)

	out, err = s.Command(nil, "-c", "GetBiosInfo", "--file", c.BiosBlob, "--file_only")
	if err != nil {
		log.Fatal("could not run sum command: ", err)
	}
	bios_target, err := sum.ParseBiosInfo(string(out))
	if err != nil {
		log.Fatal("could not parse bios info: ", err)
	}
	fmt.Println("bios info:", bios_target)

	limiter := limiter.NewConcurrencyLimiter(*parallelism)
	defer limiter.WaitAndClose()

	for host, creds := range c.Hosts {
		host := host
		creds := creds
		// prefix log messages with host name
		l := log.New(os.Stdout, host+": ", 0)

		// conf file creds take precedence over op creds
		if creds.User == "" || creds.Pass == "" {
			if *use_op {
				item, err := op.ItemGet(host)
				if err != nil {
					l.Println("could not get credentials from op:", err)
					continue
				}

				creds = op.Item2TallyCreds(item)
			} else {
				l.Println("no credentials for host:", host)
				continue
			}
		}

		limiter.Execute(func() {
			host := HostUpdate{
				Name:           host,
				TargetBmcInfo:  &bmc_target,
				TargetBiosInfo: &bios_target,
				Creds:          &creds,
				Conf:           &c,
				Sum:            s,
				Logger:         l,
				Noop:           *noop,
			}

			err = bmc_update(host)
			if err != nil {
				l.Println("could not update bmc:", err)
				return
			}

			err = bios_update(host)
			if err != nil {
				l.Println("could not update bios:", err)
				return
			}
		})
	}
}
