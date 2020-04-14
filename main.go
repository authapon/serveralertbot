package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	mc "github.com/authapon/mcryptzero"
	tg "gopkg.in/telegram-bot-api.v4"
)

type (
	host struct {
		name  string
		htype string
		state byte
		alert int64
	}

	user struct {
		name string
		id   int64
	}
)

var (
	botUP    *tg.BotAPI
	botDOWN  *tg.BotAPI
	mutex    = &sync.Mutex{}
	coreChan = make(chan string)
	hosts    = make([]host, 0)
)

func makeBot(token string, btype string) {
	for {
		bot, err := tg.NewBotAPI(token)

		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		switch btype {
		case "up":
			botUP = bot
		default:
			botDOWN = bot
		}

		go func() {
			u := tg.NewUpdate(0)
			u.Timeout = 60
			updates, err := bot.GetUpdatesChan(u)

			if err != nil {
				fmt.Printf("Error to get Channel")
				return
			}

			for update := range updates {
				if update.Message == nil {
					continue
				}

				txt := strings.TrimSpace(update.Message.Text)
				txtlow := strings.ToLower(txt)

				go func(txtdat, bbtype string, uid int64) {
					switch txtdat {
					case "myid":
						SendMsg(uid, fmt.Sprintf("Your ID is %d", uid), bbtype)
					case "status":
						coreChan <- "show status " + fmt.Sprintf("%d %s", uid, bbtype)
					case "up":
						coreChan <- "show up " + fmt.Sprintf("%d %s", uid, bbtype)
					case "down":
						coreChan <- "show down " + fmt.Sprintf("%d %s", uid, bbtype)
					case "ping":
						coreChan <- "show ping " + fmt.Sprintf("%d %s", uid, bbtype)
					case "mysql":
						coreChan <- "show mysql " + fmt.Sprintf("%d %s", uid, bbtype)
					case "web":
						coreChan <- "show web " + fmt.Sprintf("%d %s", uid, bbtype)
					case "ldap":
						coreChan <- "show ldap " + fmt.Sprintf("%d %s", uid, bbtype)
					case "dns":
						coreChan <- "show dns " + fmt.Sprintf("%d %s", uid, bbtype)
					case "watch":
						coreChan <- "show watch " + fmt.Sprintf("%d %s", uid, bbtype)
					default:
						for i := range users {
							if users[i].id == uid {
								SendMsg(uid, "Sorry! I don't understand the command.", bbtype)
							}
						}
					}
				}(txtlow, btype, int64(update.Message.From.ID))
			}
		}()
		break
	}

	SendMsgAll("NetAlertBot is *working* now.", btype)
}

func SendMsg(user int64, txt string, btype string) {
	mutex.Lock()
	msg := tg.NewMessage(user, txt)
	msg.ParseMode = "Markdown"

	switch btype {
	case "up":
		botUP.Send(msg)
	default:
		botDOWN.Send(msg)
	}

	mutex.Unlock()
}

func SendMsgAll(txt string, btype string) {
	for k, _ := range users {
		SendMsg(users[k].id, txt, btype)
	}
}

func getAllHostType() []string {
	ht := make(map[string]bool)

	for i := range hosts {
		ht[hosts[i].htype] = true
	}

	htype := make([]string, 0)

	for k, _ := range ht {
		htype = append(htype, k)
	}

	return htype
}

func timeTXT() string {
	thisTime := time.Now().UTC().Add(7 * time.Hour)
	return fmt.Sprintf("%d-%02d-%02d  %02d:%02d:%02d", thisTime.Year(), int(thisTime.Month()), thisTime.Day(), thisTime.Hour(), thisTime.Minute(), thisTime.Second())
}

func alertDown(index int) {
	SendMsgAll(fmt.Sprintf("%s\n*DOWN* -> %s\nService: *%s*", timeTXT(), hosts[index].name, strings.ToUpper(hosts[index].htype)), "down")
}

func alertUp(index int) {
	SendMsgAll(fmt.Sprintf("%s\n*UP* -> %s\nService: *%s*", timeTXT(), hosts[index].name, strings.ToUpper(hosts[index].htype)), "up")
}

func checkingHOST() {
	t := time.Now().Unix()

	for i := range hosts {
		if t > hosts[i].alert {
			go alertDown(i)
			hosts[i].state = 1
			hosts[i].alert = hosts[i].alert + repeatDuration
		}
	}
}

func uptime(htype, hostname string) {
	t := time.Now().Unix()
	found := false

	for i := range hosts {
		if hosts[i].htype == htype && hosts[i].name == hostname {
			if hosts[i].state == 1 {
				alertUp(i)
			}
			hosts[i].state = 0
			hosts[i].alert = t + downDuration
			found = true
		}
	}

	if !found {
		hostx := host{
			htype: htype,
			name:  hostname,
			state: 0,
			alert: t + downDuration,
		}

		hosts = append(hosts, hostx)
	}
}

func startHost(htype, hostname string) {
	t := time.Now().Unix()
	found := false

	for i := range hosts {
		if hosts[i].htype == htype && hosts[i].name == hostname {
			found = true
		}
	}

	if !found {
		hostx := host{
			htype: htype,
			name:  hostname,
			state: 0,
			alert: t + downDuration,
		}

		hosts = append(hosts, hostx)
	}
}

func ShowService(uid int64, service, bot string) {
	hserviceTXT := ""

	for i := range hosts {
		if hosts[i].htype == service {
			if hosts[i].state == 0 {
				hserviceTXT += "UP   -> " + hosts[i].name + "\n"
			} else {
				hserviceTXT += "*DOWN* -> " + hosts[i].name + "\n"
			}
		}
	}

	if hserviceTXT == "" {
		hserviceTXT = "no service\n"
	}

	SendMsg(uid, fmt.Sprintf("%s\nService: *%s*\n\n%s", timeTXT(), service, hserviceTXT), bot)
}

func ShowServiceState(uid int64, service, bot string, state byte) {
	hserviceTXT := ""
	stateTXT := "UP"

	if state == 1 {
		stateTXT = "*DOWN*"
	}

	for i := range hosts {
		if hosts[i].htype == service {
			if hosts[i].state == state {
				hserviceTXT += stateTXT + " -> " + hosts[i].name + "\n"
			}
		}
	}

	if hserviceTXT == "" {
		hserviceTXT = "nothing\n"
	}

	SendMsg(uid, fmt.Sprintf("%s\nService: *%s*\n\n%s", timeTXT(), service, hserviceTXT), bot)
}

func coreLoop() {
	for {
		data := <-coreChan
		arg := strings.Split(data, " ")

		switch arg[0] {
		case "show":
			go func(arg1, arg2, arg3 string) {
				uid, _ := strconv.ParseInt(arg2, 10, 64)

				switch arg1 {
				case "status":
					allHtype := getAllHostType()
					if len(allHtype) == 0 {
						SendMsg(uid, "No any service", arg3)
					} else {
						for i := range allHtype {
							go ShowService(uid, allHtype[i], arg3)
						}
					}
				case "up":
					allHtype := getAllHostType()

					if len(allHtype) == 0 {
						SendMsg(uid, "No any service", arg3)
					} else {
						for i := range allHtype {
							go ShowServiceState(uid, allHtype[i], arg3, byte(0))
						}
					}

				case "down":
					allHtype := getAllHostType()

					if len(allHtype) == 0 {
						SendMsg(uid, "No any service", arg3)
					} else {
						for i := range allHtype {
							go ShowServiceState(uid, allHtype[i], arg3, byte(1))
						}
					}
				default:
					go ShowService(uid, arg1, arg3)
				}
			}(arg[1], arg[2], arg[3])
		case "checking":
			fmt.Printf("Checking\n")
			go checkingHOST()
		default:
			datax := strings.SplitN(data, " ", 3)
			switch datax[0] {
			case "up":
				go uptime(datax[1], datax[2])
			case "start":
				go startHost(datax[1], datax[2])
			}
		}
	}
}

func UDPserver() {
	udpAddr, err := net.ResolveUDPAddr("udp", port)

	if err != nil {
		fmt.Printf("Error in UDP server!!!\n")
		panic(err)
	}

	conUDP, err := net.ListenUDP("udp", udpAddr)

	if err != nil {
		fmt.Printf("Error to listen UDP!!!\n")
		panic(err)
	}

	for {
		buffer := make([]byte, 1024)
		n, _, err := conUDP.ReadFromUDP(buffer)
		if err != nil {
			continue
		}

		dcrypt := string(buffer[:n])
		dcryptsplit := strings.SplitN(dcrypt, ":", 2)

		if len(dcryptsplit) != 2 {
			fmt.Printf("Got wrong UDP data\n")
			continue
		}

		data := string(mc.Decrypt([]byte(dcryptsplit[1]), []byte(dcryptsplit[0]+secret+dcryptsplit[0])))
		go func(datax string) {
			coreChan <- datax
			fmt.Printf("%s\n", datax)
		}(data)
	}
}

func checkLoop() {
	c := time.Tick(checkingLoopDuration)

	for _ = range c {
		coreChan <- "checking"
	}
}

func main() {
	go coreLoop()
	go UDPserver()
	go makeBot(tokenTGup, "up")
	go makeBot(tokenTGdown, "down")
	checkLoop()
}
