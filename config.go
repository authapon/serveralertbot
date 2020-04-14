package main

import (
	"time"
)

var (
	tokenTGup   = "1234:tokenGenerateFromBotFather for Up   Alert"
	tokenTGdown = "1234:tokenGenerateFromBotFather for Down Alert"

	downDuration         = int64(120)
	checkingLoopDuration = 30 * time.Second
	repeatDuration       = int64(3600)
	port                 = ":9055"

	users = []user{
		user{name: "Group1", id: -1},
		user{name: "User1", id: 1},
	}

	secret = "Secret text that share with agentalert"
)
