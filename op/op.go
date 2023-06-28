package op

import (
	"encoding/json"
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

func ItemGet(name string) OpItem {
	cmd := exec.Command("/usr/bin/op", "item", "get", name, "--format", "json")
	out, err := cmd.CombinedOutput()
	//fmt.Printf("%s\n", out)
	if err != nil {
		log.Fatal("could not run command: ", err)
	}

	var item OpItem
	json.Unmarshal(out, &item)

	return item
}

func Item2TallyCreds(item OpItem) conf.TallyCredsConf {
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
