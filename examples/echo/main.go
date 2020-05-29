package main

import (
	"flag"
	"strings"
	"tmi"
)

func main() {
	nick := flag.String("nick", "justinfan77777", "")
	pass := flag.String("pass", "oauth:", "")
	flag.Parse()

	c, err := tmi.NewClient(tmi.Auth(*nick, *pass))
	if err != nil {
		panic(err)
	}
	if err = c.Connect(); err != nil {
		panic(err)
	}

	c.Command() <- tmi.Join(*nick)
	c.Command() <- tmi.Join("nymn", "esfandtv", "sodapoppin", "greekgodx", "mizkif",
		"zoil", "liquidwifi", "vadikus007", "xqcow", "forsen", "dreamhackcs", "nmplol",
		"gofur", "asmongold", "m0xyy",
	)

	for ev := range c.Events() {
		switch ev := ev.(type) {
		case tmi.PRIVMSG:
			if strings.HasPrefix(ev.Message(), "!echo ") {
				reply := strings.TrimPrefix(ev.Message(), "!echo ")
				c.Command() <- tmi.Say(ev.Channel(), reply)
			}
		default:
			c.Default(ev)
		}
	}
}
