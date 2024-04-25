package main

import (
	"errors"
	"fmt"
	"time"

	cron "github.com/gustavoteixeira8/go-cron"
)

func main() {
	c := cron.New()

	config := &cron.CronConfig{Retries: 2, RetriesAfter: time.Second, Debug: true}

	//          s  m  h  d  m  y
	c.AddFunc("10 * * * * *", config, func() error {
		if time.Now().Second() == 12 {
			fmt.Printf("Olá mundo 1: %v\n", time.Now())
			return nil
		}

		return errors.New("some cron error")
	})

	// c.AddFunc("* * * * * *", nil, func() error {
	// 	fmt.Printf("Olá mundo 2: %v\n", time.Now())
	// 	return nil
	// })

	c.Start()
}
