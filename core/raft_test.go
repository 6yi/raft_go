package core

import (
	"context"
	"testing"
	"time"
)

func Test(t *testing.T) {
	addrs := map[int]string{
		0: "127.0.0.1:6811",
		1: "127.0.0.1:6816",
		2: "127.0.0.1:6812",
	}

	raft := NewRaft(addrs, 0)
	HeatBeat(context.Background(), raft)
	time.Sleep(1 * time.Hour)
}

func Test2(t *testing.T) {
	addrs := map[int]string{
		0: "127.0.0.1:6811",
		1: "127.0.0.1:6816",
		2: "127.0.0.1:6812",
	}

	raft := NewRaft(addrs, 1)
	HeatBeat(context.Background(), raft)
	time.Sleep(1 * time.Hour)
}

func Test3(t *testing.T) {
	addrs := map[int]string{
		0: "127.0.0.1:6811",
		1: "127.0.0.1:6816",
		2: "127.0.0.1:6812",
	}

	raft := NewRaft(addrs, 2)
	HeatBeat(context.Background(), raft)
	time.Sleep(1 * time.Hour)
}
