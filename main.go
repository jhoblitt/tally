package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/jhoblitt/tally/conf"
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
	c := conf.ParseFile("tally.yaml")

	fmt.Println("path to sum command:", c.Sum)
	fmt.Println("path to bmc blob:", c.BmcBlob)
	fmt.Println("path to bios blob:", c.BiosBlob)

	limiter := limiter.NewConcurrencyLimiter(10)
	defer limiter.WaitAndClose()

	for host, creds := range c.Hosts {
		host := host
		creds := creds

		limiter.Execute(func() {
			sum_cmd(c, host, creds, "-c", "GetBmcInfo")
			sum_cmd(c, host, creds, "-c", "GetBiosInfo")
			sum_cmd(c, host, creds, "-c", "UpdateBMC", "--file", c.BmcBlob)
			sum_cmd(c, host, creds, "-c", "UpdateBios", "--file", c.BiosBlob, "--reboot", "--preserve_setting", "--post_complete")
		})
	}
}
