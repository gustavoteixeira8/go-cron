package main

import (
	"fmt"
	"time"

	cron "github.com/gustavoteixeira8/go-cron"
)

func main() {
	c := cron.New()
	//          s  m  h  d  m  y
	c.AddFunc("20 40 08 16 03 2023", nil, func() error {
		fmt.Printf("Olá mundo 1: %v\n", time.Now())
		return nil
	})

	c.AddFunc("* * * * * *", nil, func() error {
		fmt.Printf("Olá mundo 2: %v\n", time.Now())
		return nil
	})

	c.Start()
}
