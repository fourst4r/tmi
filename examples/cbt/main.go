package main

import (
	"flag"
	"strings"
	"tmi"
)

func main() {
	nick := flag.String("nick", "", "twitch nick")
	pass := flag.String("pass", "", "twitch oauth")
	flag.Parse()

	// rt := router.New()
	// rt.On("!echo", func(r router.Resp) router.Resp {
	// 	return r.
	// })

	c, err := tmi.NewClient(tmi.Auth(*nick, *pass))
	if err != nil {
		panic(err)
	}
	if err = c.Connect(); err != nil {
		panic(err)
	}

	c.Command() <- tmi.Join(*nick)

	for ev := range c.Events() {
		switch ev := ev.(type) {
		case tmi.PRIVMSG:
			if strings.HasPrefix(ev.Message(), "!echo") {
				// go echo(c.Command(), ev)

				// reply := strings.TrimPrefix(ev.Message(), "!echo ")
				// c.Command() <- tmi.Say(ev.Channel(), reply)
			}
		default:
			c.Default(ev)
		}
	}
}

func echo(commandCh chan<- tmi.Command) {
	// commandCh <- tmi.Say(ev.Channel(), "What should I echo?")
}
