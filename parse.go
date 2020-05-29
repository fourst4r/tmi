package tmi

import (
	"fmt"
)

func isLetter(r int) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'
}

func isNumber(r int) bool {
	return r >= '0' && r <= '9'
}

func isWhite(r int) bool {
	return r == 0x20 || r == 0x00 || r == 0x0d || r == 0x0a
}

func isSpecial(r int) bool {
	switch r {
	case '-', '[', ']', '\\', '`', '^', '{', '}':
		return true
	default:
		return false
	}
}

const (
	stBegin = iota
	// <tag> [';'<tag>]*
	// <key> ['='<escaped_value>]
	// [<client_prefix>] [<vendor>'/'] <key_name>
	stTagKey
	stTagValue
	stKey
	// <host> | <nick> ['!'<user>] ['@'<host>]
	stPrefix
	stPrefixNick
	stPrefixUser
	stPrefixHost
	// <letter> {<letter>} | <number> <number> <number>
	stCommand
	stCommandLetter
	stCommandNumber1
	stCommandNumber2
	// ['@'<tags> <space>] [':'<prefix> <space>] <command> <params> <crlf>
	stMessage
	// <space> [':'<trailing> | <middle> <params>]
	stParams
	stParamsMiddle
	stParamsTrailing
	stEnd
)

type Packet struct {
	Tags   map[string]string
	Prefix struct {
		Nick, User, Host string
	}
	Command string
	Params  []string
}

func parsePacket(p []byte) (Packet, error) {
	const eof = -1
	const maxtagsize = 8191

	// Tag values are UTF8-encoded.
	// nextUTF8 := func(p []byte) rune {
	// 	r, size := utf8.DecodeRune(p[i:])
	// 	i += size
	// 	return r
	// }

	var packet Packet
	packet.Params = make([]string, 0)
	var start int
	var key string
	var b int
	state := stBegin
	i := -1

	errUnexpected := func(want string) error {
		return fmt.Errorf("parsePacket: want %q, got %q at byte %d", want, b, i)
	}

next:
	i++
	if i >= len(p) {
		b = eof
	} else {
		b = int(p[i])
	}

	switch state {
	case stBegin:
		switch b {
		case ':':
			state = stPrefixNick
			start = i + 1
		case '@':
			packet.Tags = make(map[string]string)
			state = stTagKey
			start = i + 1
		default:
			if isLetter(b) {
				state = stCommandLetter
				start = i
			} else if isNumber(b) {
				state = stCommandNumber1
				start = i
			} else {
				return packet, errUnexpected("begin")
			}
		}
	case stTagKey:
		switch b {
		default:
			// next
		case '+':
		case '=':
			key = string(p[start:i])
			state = stTagValue
			start = i + 1
		case ';':
			// missing
			start = i + 1
		case ' ':
			// missing
			state = stBegin
		case eof:
			return packet, errUnexpected("tag key")
		}
	case stTagValue:
		switch b {
		default:
			// next
		case ';':
			packet.Tags[key] = string(p[start:i])
			state = stTagKey
			start = i + 1
		case ' ':
			packet.Tags[key] = string(p[start:i])
			state = stBegin
		case 0x00, '\r', '\n', eof:
			return packet, errUnexpected("tag value")
		}
	case stPrefixNick:
		switch b {
		default:
			// TODO: handle invalid chars
			// next
		case '!':
			packet.Prefix.Nick = string(p[start:i])
			state = stPrefixUser
			start = i + 1
		case '@':
			packet.Prefix.Nick = string(p[start:i])
			state = stPrefixHost
			start = i + 1
		case ' ':
			packet.Prefix.Host = string(p[start:i])
			state = stCommand
		case eof:
			return packet, errUnexpected("prefix nick")
		}
	case stPrefixUser:
		switch b {
		default:
			// next
		case '@':
			packet.Prefix.User = string(p[start:i])
			state = stPrefixHost
			start = i + 1
		case ' ':
			packet.Prefix.User = string(p[start:i])
			state = stCommand
		case eof:
			return packet, errUnexpected("prefix user")
		}
	case stPrefixHost:
		switch b {
		default:
			// next
		case ' ':
			packet.Prefix.Host = string(p[start:i])
			state = stCommand
		case eof:
			return packet, errUnexpected("prefix host")
		}
	case stCommand:
		if isLetter(b) {
			state = stCommandLetter
			start = i
		} else if isNumber(b) {
			state = stCommandNumber1
			start = i
		} else {
			return packet, errUnexpected("command")
		}
	case stCommandLetter:
		if !isLetter(b) {
			switch b {
			case ' ':
				packet.Command = string(p[start:i])
				state = stParams
			case eof:
				packet.Command = string(p[start:])
			default:
				return packet, errUnexpected("command letter")
			}
		}
	case stCommandNumber1:
		if isNumber(b) {
			state = stCommandNumber2
		} else {
			return packet, errUnexpected("command number")
		}
	case stCommandNumber2:
		if isNumber(b) {
			packet.Command = string(p[start : start+3])
			state = stParams
		} else {
			return packet, errUnexpected("command number")
		}
	case stParams:
		// params   :: <space> [':'<trailing> | <middle> <params>]
		// middle   :: <Any *non-empty* sequence of octets not including SPACE
		//             or NUL or CR or LF, the first of which may not be ':'>
		// trailing :: <Any, possibly *empty*, sequence of octets not including
		//             NUL or CR or LF>
		switch b {
		case ' ':
			// skip
		case 0x00, '\r', '\n':
			return packet, errUnexpected("params")
		case ':':
			state = stParamsTrailing
			start = i + 1 // next is the start of the param
		default:
			state = stParamsMiddle
			start = i // this is the start of the param
		}
	case stParamsMiddle:
		switch b {
		case ' ':
			packet.Params = append(packet.Params, string(p[start:i]))
			state = stParams
		case eof:
			packet.Params = append(packet.Params, string(p[start:]))
		case 0x00, '\r', '\n':
			return packet, errUnexpected("middle")
		default:
			// next
		}
	case stParamsTrailing:
		switch b {
		case eof:
			packet.Params = append(packet.Params, string(p[start:]))
		case 0x00, '\r', '\n':
			return packet, errUnexpected("trailing")
		default:
			// next
		}
	default:
		panic("unreachable")
	}

	if b != eof {
		goto next
	}

	return packet, nil
}
