package tmi

type Option func(*Client)

func SSL(c *Client) error { return nil }

func Auth(nick, pass string) Option {
	return func(c *Client) {
		c.nick, c.pass = nick, pass
	}
}

const (
	// CapMembership adds membership state event data.
	CapMembership = "twitch.tv/membership"
	// CapTags adds IRC V3 message tags to several commands.
	CapTags = "twitch.tv/tags"
	// CapCommands enables several Twitch-specific commands.
	CapCommands = "twitch.tv/commands"
)

func Cap(caps ...string) Option {
	return func(c *Client) {
		c.capabilities = append(c.capabilities, caps...)
	}
}
