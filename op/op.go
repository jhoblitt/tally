package op

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"

	"github.com/jhoblitt/tally/conf"
)

type OpItem struct {
	Id     string    `json:"id"`
	Fields []OpField `json:"fields"`
}

type OpField struct {
	Id    string `json:"id"`
	Value string `json:"value"`
}

func ItemGet(name string) (*OpItem, error) {
	cmd := exec.Command("/usr/bin/op", "item", "get", name, "--format", "json")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("could not find item: %s - %w", name, err)
	}

	var item OpItem
	json.Unmarshal(out, &item)

	return &item, nil
}

func Item2TallyCreds(item *OpItem) conf.TallyCredsConf {
	var cred conf.TallyCredsConf
	for _, field := range item.Fields {
		if field.Id == "username" {
			cred.User = field.Value
		}
		if field.Id == "password" {
			cred.Pass = field.Value
		}
	}

	if cred.User == "" {
		log.Fatal("could not get username")
	}

	if cred.Pass == "" {
		log.Fatal("could not get password")
	}

	return cred
}
