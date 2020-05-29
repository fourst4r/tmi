package tmi

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

var (
	errNoCommandsCap   = errors.New("tmi: no commands capability")
	errNoMembershipCap = errors.New("tmi: no membership capability")
	errNoTagsCap       = errors.New("tmi: no tags capability")
)

type Event interface{}

type (
	UNKNOWN Packet

	CLEARCHAT  Packet // Purge a user’s message(s), typically after a user is banned from chat or timed out.
	CLEARMSG   Packet // Single message removal on a channel. This is triggered via /delete <target-msg-id> on IRC.
	HOSTTARGET Packet // Channel starts or stops host mode.
	NOTICE     Packet // General notices from the server.
	PING       Packet
	PRIVMSG    Packet
	RECONNECT  Packet // Rejoin channels after a restart.
	ROOMSTATE  Packet // Identifies the channel’s chat settings (e.g., slow mode duration).
	USERNOTICE Packet // Announces Twitch-specific events to the channel (e.g., a user’s subscription notification).
	USERSTATE  Packet // Identifies a user’s chat settings or properties (e.g., chat color).
	WHISPER    Packet
)

func (p *CLEARCHAT) Channel() string { return p.Params[0][1:] }
func (p *CLEARCHAT) Nick() string {
	if len(p.Params) > 1 {
		return p.Params[1]
	}
	return ""
}

func (p *CLEARMSG) Login() (string, error)       { return tagorerr(p, "login") }
func (p *CLEARMSG) TargetMsgID() (string, error) { return tagorerr(p, "target-msg-id") }
func (p *CLEARMSG) Channel() string              { return p.Params[0][1:] }
func (p *CLEARMSG) Message() string              { return p.Params[1] }

func (p *HOSTTARGET) HostingChannel() string { return p.Params[0][1:] }
func (p *HOSTTARGET) Channel() string        { return strings.SplitN(p.Params[1], " ", 1)[0] }
func (p *HOSTTARGET) NumViewers() (int, error) {
	split := strings.SplitN(p.Params[1], " ", 1)
	if len(split) > 1 {
		return strconv.Atoi(split[1])
	}
	return 0, nil
}

func (p *NOTICE) Channel() string        { return p.Params[0][1:] }
func (p *NOTICE) Message() string        { return p.Params[1] }
func (p *NOTICE) MsgID() (string, error) { return tagorerr(p, "msg-id") }

func (p *PRIVMSG) Channel() string { return p.Params[0][1:] }
func (p *PRIVMSG) Message() string { return p.Params[1] }
func (p *PRIVMSG) Author() string  { return p.Prefix.Nick }

func (p *ROOMSTATE) Channel() string { return p.Params[0][1:] }

func (p *USERNOTICE) Channel() string { return p.Params[0][1:] }
func (p *USERNOTICE) Message() string { return p.Params[1] }

func (p *USERSTATE) Channel() string { return p.Params[0][1:] }

func tagorerr(p interface{}, tag string) (string, error) {
	packet := p.(*Packet)
	if packet.Tags == nil {
		return "", errNoTagsCap
	}
	return packet.Tags[tag], nil
}

func toevent(p Packet) Event {
	switch p.Command {
	case "CLEARCHAT":
		return CLEARCHAT(p)
	case "CLEARMSG":
		return CLEARMSG(p)
	case "HOSTTARGET":
		return HOSTTARGET(p)
	case "NOTICE":
		return NOTICE(p)
	case "PING":
		return PING(p)
	case "PRIVMSG":
		return PRIVMSG(p)
	case "RECONNECT":
		return RECONNECT(p)
	case "ROOMSTATE":
		return ROOMSTATE(p)
	case "USERNOTICE":
		return USERNOTICE(p)
	case "USERSTATE":
		return USERSTATE(p)
	case "WHISPER":
		return WHISPER(p)
	default:
		return UNKNOWN(p)
	}
}

const (
	url           = "irc.chat.twitch.tv:6667"
	urlssl        = "irc.chat.twitch.tv:6697"
	maxpacketsize = 510
	Delim         = "\r\n"
)

// Command executes a command to the server.
// Be wary that there is a potential for race conditions if you access
// non-thread-safe variables.
type Command func(io.Writer)

// Join a twitch channel.
func Join(channels ...string) Command {
	// return func(w io.Writer) {
	// 	var buf bytes.Buffer
	// 	var size int
	// 	buf.WriteString("JOIN #")
	// 	buf.WriteString(channels[0])
	// 	for _, channel := range channels[1:] {
	// 		size += len(channel)
	// 		if size > maxpacketsize {
	// 			buf.WriteString(Delim)
	// 			buf.WriteString("JOIN #")
	// 			buf.WriteString(channel)
	// 			size = len(channel)
	// 			continue
	// 		}
	// 		buf.WriteString(",#")
	// 		buf.WriteString(channel)
	// 	}
	// 	buf.WriteString(Delim)
	// 	// err := w.Flush()
	// 	_, err := buf.WriteTo(w)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// }
	// TODO: split into multiple packets when exceeding IRC packet size limit.
	//       is it even necessary? it works when you exceed 512.
	return Line("JOIN #" + strings.Join(channels, ",#"))
}

// Part from a twitch channel.
func Part(channel string) Command { return Line("PART #" + channel) }

// Say something in a channel.
func Say(channel, message string) Command {
	return Line("PRIVMSG #" + channel + " :" + message)
}

// Pong is a reply to PING.
func Pong() Command { return Line("PONG :tmi.twitch.tv") }

// Line writes a line to the server.
func Line(packet string) Command {
	return func(w io.Writer) {
		fmt.Println("->", packet)
		w.Write(append([]byte(packet), Delim...))
	}
}

type Client struct {
	conn         *net.TCPConn
	nick, pass   string
	capabilities []string
	events       chan Event
	commands     chan Command
}

func NewClient(options ...Option) (*Client, error) {
	var c Client

	// Set default options
	Auth("justinfan77777", "oauth:ThisIsAnAnonymousAuth_forsenPls")(&c)
	Cap(CapCommands, CapMembership, CapTags)(&c)

	for _, option := range options {
		option(&c)
	}

	return &c, nil
}

// Connect to Twitch chat.
func (c *Client) Connect() error {
	var err error

	addr, err := net.ResolveTCPAddr("tcp", url)
	if err != nil {
		return err
	}

	c.conn, err = net.DialTCP("tcp", nil, addr)
	if err != nil {
		return err
	}

	// TODO: should these be buffered?
	c.events = make(chan Event)
	c.commands = make(chan Command)

	go func() {
		conn := c.conn
		events := c.events
		r := bufio.NewReader(conn)
		for {
			line, _, err := r.ReadLine() // TODO: can packets contain \n without \r?
			if err != nil {
				if err == io.EOF {
					break
				}
				panic(err)
			}
			fmt.Println("<-", string(line))
			p, err := parsePacket(line)
			if err != nil {
				// just log it for now, not sure what to do here 🤔
				fmt.Println("failed to parse packet: ", err)
				continue
			}
			if len(events) == cap(events)-1 {
				fmt.Println("events channel is full, about to block")
			}
			events <- toevent(p)
		}
		fmt.Println("!!!!!!!!!!!!! exited read loop")
	}()

	go func() {
		conn := c.conn
		commands := c.commands
		w := bufio.NewWriter(conn)
		// w := io.MultiWriter(conn, NewPrefixer(os.Stdout, func() string { return "-> " }))

		for command := range commands {
			command(w)
			w.Flush()
		}
		fmt.Println("!!!!!!!!!!!!! exited write loop")
	}()

	c.commands <- Line("PASS " + c.pass)
	c.commands <- Line("NICK " + c.nick)
	c.commands <- Line("CAP REQ :" + strings.Join(c.capabilities, " "))

	return nil
}

// Close the connection.
func (c *Client) Close() error {
	close(c.commands)
	close(c.events)
	return c.conn.Close()
}

// Default handling of events.
func (c *Client) Default(event Event) {
	switch event.(type) {
	case PING:
		c.commands <- Pong()
	}
}

func (c *Client) Events() <-chan Event    { return c.events }
func (c *Client) Command() chan<- Command { return c.commands }
