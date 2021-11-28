package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type (
	configServerST struct {
		TokenTGup            string   `yaml:"tokentgup"`
		TokenTGdown          string   `yaml:"tokentgdown"`
		DownDuration         int64    `yaml:"downduration"`
		CheckingLoopDuration int64    `yaml:"checkingloop"`
		RepeatDuration       int64    `yaml:"repeatduration"`
		Port                 string   `yaml:"port"`
		Users                []userST `yaml:"users"`
		Secret               string   `yaml:"secret"`
	}
	userST struct {
		Name string `yaml:"name"`
		Id   int64  `yaml:"id"`
	}
)

var (
	tokenTGup   = ""
	tokenTGdown = ""

	downDuration         = int64(0)
	checkingLoopDuration = time.Second
	repeatDuration       = int64(0)
	port                 = ""

	users = []user{}

	secret = ""
)

func configProcess() {
	if len(os.Args) > 1 {
		yfile, err := ioutil.ReadFile(os.Args[1])
		if err != nil {
			fmt.Printf("Cannot read file config\n\n---------------\n\n")
			usage()
		}
		configDat := configServerST{}
		err2 := yaml.Unmarshal(yfile, &configDat)
		if err2 != nil {
			fmt.Printf("Error in file config\n%v\n----------------\n\n", err2)
			usage()
		}

		tokenTGup = configDat.TokenTGup
		tokenTGdown = configDat.TokenTGdown
		downDuration = configDat.DownDuration
		checkingLoopDuration = time.Duration(configDat.CheckingLoopDuration) * time.Second
		repeatDuration = configDat.RepeatDuration
		port = configDat.Port
		secret = configDat.Secret
		for _, v := range configDat.Users {
			u := user{}
			u.name = v.Name
			u.id = v.Id
			users = append(users, u)
		}

		fmt.Printf("TokenTGup = %v\n", tokenTGup)
		fmt.Printf("TokenTGdown = %v\n", tokenTGdown)
		fmt.Printf("Secret = %v\n", secret)
		fmt.Printf("DownDuration = %vs\n", downDuration)
		fmt.Printf("CheckingLoop = %v\n", checkingLoopDuration)
		fmt.Printf("RepeatDuration = %vs\n", repeatDuration)
		fmt.Printf("Port = %v\n", port)

		for _, v := range users {
			fmt.Printf("%v\n", v)
		}
		return
	}
	usage()
}

func usage() {
	fmt.Printf("usage: serveralertbot <config.yaml>\n\nExample config:\n-------------\n")
	fmt.Printf("tokentgup: \"1234:tokenGenerateFromBotFather for Up   Alert\"\n")
	fmt.Printf("tokentgdown: \"1234:tokenGenerateFromBotFather for Down Alert\"\n")
	fmt.Printf("secret: \"Secret text that share with agentalert\"\n")
	fmt.Printf("downduration: 120\n")
	fmt.Printf("checkingloop: 30\n")
	fmt.Printf("repeatduration: 3600\n")
	fmt.Printf("port: \":9055\"\n")
	fmt.Printf("users:\n")
	fmt.Printf("  - name: \"Group1\"\n")
	fmt.Printf("    id: -1\n")
	fmt.Printf("  - name: \"User1\"\n")
	fmt.Printf("    id: 1\n")
	os.Exit(0)
}
