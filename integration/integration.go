package app

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/jomei/notionapi"
)

type Event interface {
	OnTicker() error
	Close() error
}

type Integration struct {
	wg     sync.WaitGroup
	events []Event
}

type Opt func(*Integration)

func WithEvent(events ...Event) Opt {
	return func(i *Integration) {
		for _, e := range events {
			i.events = append(i.events, e)
		}
	}
}

func NewIntegration(client *notionapi.Client, opts ...Opt) *Integration {
	i := &Integration{}

	for _, opt := range opts {
		i.wg.Add(1)
		opt(i)
	}

	return i
}

func (integrion *Integration) gracefulShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan

	for i := len(integrion.events) - 1; i >= 0; i-- {
		if err := integrion.events[i].Close(); err != nil {
			log.Fatal(err)
		}

		integrion.wg.Done()
	}

	integrion.wg.Wait()
}

func (integrion *Integration) Run() {
	for _, e := range integrion.events {
		go func(e Event) {
			if err := e.OnTicker(); err != nil {
				log.Fatal(err)
			}
		}(e)
	}
}
