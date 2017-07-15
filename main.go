package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/tbruyelle/hipchat-go/hipchat"
)

func main() {
	token := os.Getenv("SWALLOW_TOKEN")
	if len(token) == 0 {
		fmt.Println("export SWALLOW_TOKEN=<YOUR_HIPCHAT_TOKEN>")
		return
	}
	ids := os.Getenv("SWALLOW_ROOM_ID")
	if len(ids) == 0 {
		fmt.Println("export SWALLOW_ROOM_ID=<ROOM_ID>,<ROOM_ID>,...")
		return
	}

	c := hipchat.NewClient(token)
	opt := &hipchat.RoomsListOptions{}
	list, _, err := c.Room.List(opt)
	if err != nil {
		fmt.Println(err)
		return
	}

	rooms := []hipchat.Room{}
	for _, r := range list.Items {
		for _, id := range strings.Split(ids, ",") {
			i, _ := strconv.Atoi(id)
			if r.ID == i {
				rooms = append(rooms, r)
			}
		}
	}

	for _, room := range rooms {
		fmt.Println(room.ID, room.Name)
		h, _, err := c.Room.History(room.Name, &hipchat.HistoryOptions{})
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, m := range h.Items {
			fmt.Println(m.Date, m.From, m.Message)
		}
	}
}
