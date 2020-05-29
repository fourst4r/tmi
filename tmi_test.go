package tmi

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestClient_Connect(t *testing.T) {
	c, err := NewClient()
	if err != nil {
		t.Error(err)
	}
	if err = c.Connect(); err != nil {
		t.Error(err)
	}
	for event := range c.Events() {
		switch event.(type) {
		case CLEARMSG:
			fmt.Println("clearmsg")
		default:
			fmt.Println("default")
		}
	}

	fmt.Println("blocking main")
	<-time.After(time.Second * 20)
}

func TestNewClient(t *testing.T) {
	type args struct {
		options []Option
	}
	tests := []struct {
		name    string
		args    args
		want    *Client
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewClient(tt.args.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewClient() = %v, want %v", got, tt.want)
			}
		})
	}
}
