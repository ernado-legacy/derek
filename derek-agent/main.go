package main

import (
	"flag"
	"net/http"

	"github.com/cydev/derek"
	"github.com/nats-io/nats"
	"time"
)

var (
	queue = flag.String("queue", "default", "NATS queue name")
	subject = flag.String("subject", "request.*", "NATS subscribe subject")
	url = flag.String("url", nats.DefaultURL, "NATS url")
)

func main() {
	c, err := nats.Connect(*url)
	if err != nil {
		panic(err)
	}
	agent, err := derek.NewNATSAgent(*subject, *queue, c, http.DefaultClient)
	if err != nil {
		panic(err)
	}
	ticker := time.NewTicker(time.Second)
	for _ = range ticker.C {
		if agent != agent {
			panic("error")
		}
	}
}
