package tmi

import (
	"reflect"
	"testing"
)

func Test_parsePacket(t *testing.T) {
	type args struct {
		p []byte
	}
	tests := []struct {
		name    string
		args    args
		want    Packet
		wantErr bool
	}{
		{
			name: "command",
			args: args{[]byte("PRIVMSG")},
			want: Packet{
				Command: "PRIVMSG",
				Params:  []string{},
			},
		},
		{
			name: "command middle",
			args: args{[]byte("PRIVMSG #nymn")},
			want: Packet{
				Command: "PRIVMSG",
				Params:  []string{"#nymn"},
			},
		},
		{
			name: "command middle trailing",
			args: args{[]byte("PRIVMSG #nymn :nobody knows")},
			want: Packet{
				Command: "PRIVMSG",
				Params:  []string{"#nymn", "nobody knows"},
			},
		},
		{
			name: "command middle empty trailing",
			args: args{[]byte("PRIVMSG #nymn :")},
			want: Packet{
				Command: "PRIVMSG",
				Params:  []string{"#nymn", ""},
			},
		},
		{
			name: "command trailing",
			args: args{[]byte("PRIVMSG :nobody knows")},
			want: Packet{
				Command: "PRIVMSG",
				Params:  []string{"nobody knows"},
			},
		},
		{
			name: "command empty trailing",
			args: args{[]byte("PRIVMSG :")},
			want: Packet{
				Command: "PRIVMSG",
				Params:  []string{""},
			},
		},
		{
			name: "number command",
			args: args{[]byte(":tmi.twitch.tv 421 justinfan64537 WHO :Unknown command")},
			want: Packet{
				Prefix: struct {
					Nick, User, Host string
				}{"", "", "tmi.twitch.tv"},
				Command: "421",
				Params:  []string{"justinfan64537", "WHO", "Unknown command"},
			},
		},
		{
			name: "host",
			args: args{[]byte(":tmi.twitch.tv CLEARMSG #dallas :HeyGuys")},
			want: Packet{
				Prefix: struct {
					Nick, User, Host string
				}{"", "", "tmi.twitch.tv"},
				Command: "CLEARMSG",
				Params:  []string{"#dallas", "HeyGuys"},
			},
		},
		{
			name: "tags",
			args: args{[]byte("@badge-info=;badges=staff/1;color=#0D4200;display-name=ronni;emote-sets=0,33,50,237,793,2126,3517,4578,5569,9400,10337,12239;mod=1;subscriber=1;turbo=1;user-type=staff :tmi.twitch.tv USERSTATE #dallas")},
			want: Packet{
				Tags: map[string]string{
					"badge-info":   "",
					"badges":       "staff/1",
					"color":        "#0D4200",
					"display-name": "ronni",
					"emote-sets":   "0,33,50,237,793,2126,3517,4578,5569,9400,10337,12239",
					"mod":          "1",
					"subscriber":   "1",
					"turbo":        "1",
					"user-type":    "staff",
				},
				Prefix: struct {
					Nick, User, Host string
				}{"", "", "tmi.twitch.tv"},
				Command: "USERSTATE",
				Params:  []string{"#dallas"},
			},
		},
		{
			name: "tags missing value",
			args: args{[]byte("@login;target-msg-id :tmi.twitch.tv CLEARMSG #dallas :HeyGuys")},
			want: Packet{
				Tags: map[string]string{},
				Prefix: struct {
					Nick, User, Host string
				}{"", "", "tmi.twitch.tv"},
				Command: "CLEARMSG",
				Params:  []string{"#dallas", "HeyGuys"},
			},
		},
		{
			name: "nick user host",
			args: args{[]byte(":justinfan64537!justinfan64537@justinfan64537.tmi.twitch.tv JOIN #nymn")},
			want: Packet{
				Prefix: struct {
					Nick, User, Host string
				}{"justinfan64537", "justinfan64537", "justinfan64537.tmi.twitch.tv"},
				Command: "JOIN",
				Params:  []string{"#nymn"},
			},
		},
		{
			name: "nick user",
			args: args{[]byte(":justinfan64537!justinfan64537 JOIN #nymn")},
			want: Packet{
				Prefix: struct {
					Nick, User, Host string
				}{"justinfan64537", "justinfan64537", ""},
				Command: "JOIN",
				Params:  []string{"#nymn"},
			},
		},
		{
			name: "nick host",
			args: args{[]byte(":justinfan64537@justinfan64537.tmi.twitch.tv JOIN #nymn")},
			want: Packet{
				Prefix: struct {
					Nick, User, Host string
				}{"justinfan64537", "", "justinfan64537.tmi.twitch.tv"},
				Command: "JOIN",
				Params:  []string{"#nymn"},
			},
		},
		{
			name: "malformed emote",
			args: args{[]byte("@badge-info=subscriber/52;badges=moderator/1,subscriber/48;color=#2E8B57;display-name=pajbot;emotes=80481_/3:7-14;flags=;id=1ec936d3-7853-4113-9984-664ac5c42694;mod=1;room-id=11148817;subscriber=1;tmi-sent-ts=1589640131796;turbo=0;user-id=82008718;user-type=mod :pajbot!pajbot@pajbot.tmi.twitch.tv PRIVMSG #pajlada :󠀀-tags pajaW_/3.0")},
			want: Packet{
				Tags: map[string]string{
					"badge-info":   "subscriber/52",
					"badges":       "moderator/1,subscriber/48",
					"color":        "#2E8B57",
					"display-name": "pajbot",
					"emotes":       "80481_/3:7-14",
					"flags":        "",
					"id":           "1ec936d3-7853-4113-9984-664ac5c42694",
					"mod":          "1",
					"room-id":      "11148817",
					"subscriber":   "1",
					"tmi-sent-ts":  "1589640131796",
					"turbo":        "0",
					"user-id":      "82008718",
					"user-type":    "mod",
				},
				Prefix: struct {
					Nick, User, Host string
				}{"pajbot", "pajbot", "pajbot.tmi.twitch.tv"},
				Command: "PRIVMSG",
				Params:  []string{"#pajlada", "󠀀-tags pajaW_/3.0"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePacket(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePacket() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parsePacket() = %v, want %v", got, tt.want)
			}
		})
	}
}
