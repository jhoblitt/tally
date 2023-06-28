package sum

import (
	"regexp"
	"strings"
)

type SumBMC struct {
	UFFN    string
	Type    string
	Version string
	Date    string
}

func ParseBMC(text string) (SumBMC, error) {
	var bmc SumBMC

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
