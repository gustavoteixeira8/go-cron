package main

import (
	"fmt"
	"time"

	cron "github.com/gustavoteixeira8/go-cron"
)

func main() {
	c := cron.New()

	config := &cron.CronConfig{
		Retries:      2,
		RetriesAfter: time.Second,
		Debug:        true,
		Callback: func() error {
			fmt.Printf("Olá mundo 1: %v\n", time.Now())
			return nil
		},
		Close: make(chan bool),
	}

	//          s  m  h  d  m  y
	c.AddFunc("* * * * * *", config)

	go func() {
		time.Sleep(time.Second * 5)
		config.Close <- true

		time.Sleep(time.Second)

		c.AddFunc("* * * * * *", config)
	}()

	c.Wait()
}
