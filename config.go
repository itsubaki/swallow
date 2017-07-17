package main

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Token     string
	RoomIDs   []int
	Period    int
	HighLight []string
}

func NewConfig() (*Config, error) {
	token := os.Getenv("SWALLOW_TOKEN")
	if len(token) == 0 {
		return nil, errors.New("export SWALLOW_TOKEN=<YOUR_HIPCHAT_TOKEN>")
	}

	room_id := os.Getenv("SWALLOW_ROOM_ID")
	if len(room_id) == 0 {
		return nil, errors.New("export SWALLOW_ROOM_ID=<ROOM_ID>,<ROOM_ID>,...")
	}

	period := os.Getenv("SWALLOW_PERIOD")
	if len(period) == 0 {
		period = "60"
	}

	highlight := os.Getenv("SWALLOW_HIGHLIGHT")

	ids := []int{}
	for _, id := range strings.Split(room_id, ",") {
		i, err := strconv.Atoi(id)
		if err != nil {
			continue
		}
		ids = append(ids, i)
	}

	p, err := strconv.Atoi(period)
	if err != nil {
		return nil, err
	}

	return &Config{token, ids, p, strings.Split(highlight, ",")}, nil
}
