package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"strings"
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
	}

	rooms := []hipchat.Room{}
	for _, id := range s.config.RoomIDs {
		if r, ok := tmp[id]; ok {
			rooms = append(rooms, r)
			fmt.Println(r.ID, r.Name)
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
			if reflect.TypeOf(m.From) == reflect.TypeOf("") {
				continue
			}
			tindex := strings.Index(m.Date, ".")
			time := m.Date[:tindex]
			name := m.From.(map[string]interface{})["name"].(string)
			findex := strings.Index(name, "(")
			from := name[:findex]
			message := strings.Replace(m.Message, "\n", " ", -1)

			light := false
			for _, h := range s.config.HighLight {
				if strings.Contains(message, h) {
					light = true
					break
				}
			}
			out := "[" + e.Name + "] " + time + " " + from + " " + message
			if light {
				fmt.Printf("\x1b[31m%s\x1b[0m", out)
			} else {
				fmt.Print(out)
			}
			fmt.Println()
		}
	}
}

func (s *Swallow) History(c *hipchat.Client, room hipchat.Room) {
	t := time.NewTicker(time.Duration(s.config.Period) * time.Second)

	items, latest := s.LatestHistory(c, room, "")
	for _, m := range items {
		s.queue <- Event{room.Name, m}
	}

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-t.C:
			items, latest = s.LatestHistory(c, room, latest)
			for _, m := range items {
				s.queue <- Event{room.Name, m}
			}
		}
	}
}

func (s *Swallow) LatestHistory(c *hipchat.Client, room hipchat.Room, latest string) ([]hipchat.Message, string) {
	h, _, err := c.Room.History(room.Name, &hipchat.HistoryOptions{EndDate: latest})
	if err != nil {
		fmt.Println("Swallow.History", err)
		return nil, latest
	}

	items := []hipchat.Message{}
	for _, m := range h.Items {
		items = append(items, m)
		if m.Date == latest {
			items = []hipchat.Message{}
		}
	}

	return items, h.Items[len(h.Items)-1].Date
}
