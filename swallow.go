package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"github.com/tbruyelle/hipchat-go/hipchat"
)

type Swallow struct {
	config *Config
	queue  chan Event
	ctx    context.Context
	cancel func()
}

type Event struct {
	Name    string
	Message hipchat.Message
}

func NewSwallow(config *Config) *Swallow {
	ctx, cancel := context.WithCancel(context.Background())
	return &Swallow{
		config,
		make(chan Event, 1024),
		ctx,
		cancel,
	}
}

func (s *Swallow) ShutdownHook() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c // blocking
		s.cancel()
		os.Exit(0)
	}()
}

func (s *Swallow) Run() {
	c := hipchat.NewClient(s.config.Token)
	list, _, err := c.Room.List(&hipchat.RoomsListOptions{})
	if err != nil {
		fmt.Println("List", err)
		return
	}

	tmp := make(map[int]hipchat.Room)
	for _, r := range list.Items {
		tmp[r.ID] = r
		fmt.Println(r.ID, r.Name)
	}

	rooms := []hipchat.Room{}
	for _, id := range s.config.RoomIDs {
		if r, ok := tmp[id]; ok {
			rooms = append(rooms, r)
		}
	}

	for _, room := range rooms {
		go s.History(c, room)
	}

	s.Display()
}

func (s *Swallow) Display() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case e := <-s.queue:
			m := e.Message
			if reflect.TypeOf(m.From) != reflect.TypeOf("") {
				from := m.From.(map[string]interface{})["name"]
				message := m.Message
				fmt.Println("["+e.Name+"]", from, message)
			}
		}
	}
}

func (s *Swallow) History(c *hipchat.Client, room hipchat.Room) {
	t := time.NewTicker(60 * time.Second)
	var latest string
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-t.C:
			h, _, err := c.Room.History(room.Name, &hipchat.HistoryOptions{EndDate: latest})
			if err != nil {
				fmt.Println("History", err)
				return
			}
			items := []hipchat.Message{}
			for _, m := range h.Items {
				items = append(items, m)
				if m.Date == latest {
					items = []hipchat.Message{}
				}
			}
			latest = h.Items[len(h.Items)-1].Date

			for _, m := range items {
				s.queue <- Event{room.Name, m}
			}
		}
	}
}
