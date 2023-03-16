package main

import (
	"errors"
	"fmt"
	"runtime"
	"time"

	cron "github.com/gustavoteixeira8/go-cron"
)

func main() {
	c := cron.New()

	c.AddFunc("* * * * * *", nil, func() error {
		fmt.Printf("Olá mundo 1: %v\n", time.Now())
		return errors.New("some error")
	})

	c.AddFunc("* * * * * *", nil, func() error {
		fmt.Printf("Olá mundo 2: %v\n", time.Now())
		return nil
	})

	go c.Start()

	for {
		fmt.Println()
		fmt.Println(runtime.NumGoroutine())
		fmt.Println()
		time.Sleep(time.Second)
	}
}
