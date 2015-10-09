package main

import (
	"flag"
	"net/http"

	"github.com/cydev/derek"
	"github.com/nats-io/nats"
	"time"
	"fmt"
	"bytes"
)

var (
	subject = flag.String("subject", "request.1", "NATS subscribe subject")
	url = flag.String("url", nats.DefaultURL, "NATS url")
)

func main() {
	c, err := nats.Connect(*url)
	if err != nil {
		panic(err)
	}

	client := derek.NewNATSClient(*subject, c)
	ticker := time.NewTicker(time.Second)
	for _ = range ticker.C {
		req, err := http.NewRequest("GET", "http://ya.ru", nil)
		if err != nil {
			panic(err)
		}
		r, err := derek.NewRequest(req)
		if err != nil {
			panic(err)
		}
		start := time.Now()
		res, err := client.Do(r)
		duration := time.Now().Sub(start)
		if err != nil {
			fmt.Println("ERR", duration, err)
		} else {
			fmt.Println("DONE", duration, res.StatusCode, bytes.NewBuffer(res.Body))
		}
	}
}
