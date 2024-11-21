package sum

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/jhoblitt/tally/conf"
)

type SumBmcInfo struct {
	UFFN    string
	Type    string
	Version string
	Date    string
}

type SumBiosInfo struct {
	BoardID   string
	BuildDate string
}

type Sum struct {
	ExecCommand func(string, ...string) *exec.Cmd
	Path        string
}

func ParseBmcInfo(text string) (SumBmcInfo, error) {
	var bmc SumBmcInfo

	re, err := regexp.Compile(`\.{3,}(\S+)$`)
	if err != nil {
		return bmc, err
	}

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		m := re.FindStringSubmatch(line)

		if strings.Contains(line, "BMC UFFN...") {
			bmc.UFFN = m[1]
		} else if strings.Contains(line, "BMC type...") {
			bmc.Type = m[1]
		} else if strings.Contains(line, "BMC version...") {
			bmc.Version = m[1]
		} else if strings.Contains(line, "BMC build date...") {
			bmc.Date = m[1]
		}
	}

	return bmc, nil
}

func ParseBiosInfo(text string) (SumBiosInfo, error) {
	var bios SumBiosInfo

	re, err := regexp.Compile(`\.{3,}(\S+)$`)
	if err != nil {
		return bios, err
	}

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		m := re.FindStringSubmatch(line)

		if strings.Contains(line, "Board ID...") {
			bios.BoardID = m[1]
		} else if strings.Contains(line, "BIOS build date...") {
			bios.BuildDate = m[1]
		}
	}

	return bios, nil
}

func NewSum(path string) *Sum {
	s := &Sum{
		Path:        path,
		ExecCommand: exec.Command,
	}
	return s
}

func (s *Sum) Command(creds *conf.TallyCredsConf, arg ...string) ([]byte, error) {
	if creds != nil {
		// put the password in a temp file to avoid leaking it on the command line
		f, err := os.CreateTemp("", "tally")
		if err != nil {
			return nil, fmt.Errorf("failed running %s: %w", arg, err)
		}
		defer os.Remove(f.Name())

		if _, err := f.Write([]byte(creds.Pass)); err != nil {
			return nil, fmt.Errorf("failed running %s: %w", arg, err)
		}
		if err := f.Close(); err != nil {
			return nil, fmt.Errorf("failed running %s: %w", arg, err)
		}

		arg = append(arg, "-u", creds.User, "-f", f.Name())
	}

	log.Printf("running: %s %s", s.Path, strings.Join(arg, " "))

	cmd := s.ExecCommand(s.Path, arg...)
	if cmd == nil {
		return nil, fmt.Errorf("failed running %s", arg)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed running %s: %w", arg, err)
	}

	return out, nil
}
