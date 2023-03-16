package cron

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const _FIXED_CRON_LEN = 6

type Callback func() error

type CronConfig struct {
	Retries      int
	RetriesAfter time.Duration
}

type Cron struct {
	funcs  map[string]Callback
	config map[string]*CronConfig
}

/*
cron.AddFunc adiciona um cron para ser executado

---

Parameters:

  - cronValue:

    S M H D M Y

    S = Second (0-59)

    M = Minute (0-59)

    H = Hour (0-59)

    D = Day (1-31)

    M = Month (1-12)

    Y = Year (ano atual ou maior)

    ---

  - config:

    Retries: número de vezes que a callback será executada em caso de erro

    RetriesAfter: tempo de espera entre uma execução e outra

    ---

  - callback: função que será executada no momento em que o cronValue for verdadeiro
*/
func (c *Cron) AddFunc(cronValue string, config *CronConfig, callback Callback) {
	if c.funcs == nil {
		c.funcs = map[string]Callback{}
	}
	if c.config == nil {
		c.config = map[string]*CronConfig{}
	}

	cronValue = fmt.Sprintf("%s %s", cronValue, uuid.NewString())

	c.config[cronValue] = config
	c.funcs[cronValue] = callback
}

// Inicia todos os crons e trava o processo.
func (c *Cron) Start() {
	var wg sync.WaitGroup

	for cron, cb := range c.funcs {
		wg.Add(1)
		config := c.config[cron]
		go c.processCron(cron, config, cb)
	}

	wg.Wait()
}

// Verifica se o cronValue corresponde a hora atual e executa a callback.
func (c *Cron) processCron(cronValue string, config *CronConfig, callback Callback) {
	defer func() {
		r := recover()
		if r != nil {
			log.Printf("Error recovering: %v\n", r)
		}
	}()
	cronSplitted := strings.Split(cronValue, " ")
	cronSplitted = cronSplitted[:len(cronSplitted)-1]

	if len(cronSplitted) != _FIXED_CRON_LEN {
		log.Fatalln("Cron value only supports Second, Minute, Hour, Day, Month, Year (* * * * * *)")
		return
	}

	executedAlready := false
	executeEverySecond := false

	for {
		now := time.Now()
		timeValues := []int64{
			int64(now.Second()),
			int64(now.Minute()),
			int64(now.Hour()),
			int64(now.Day()),
			int64(now.Month()),
			int64(now.Year()),
		}
		fmt.Println(now.Second())
		canRunCallback := false

		for i, timeUnit := range cronSplitted {
			if timeUnit == "*" {
				canRunCallback = true
				executeEverySecond = true
				continue
			}
			executeEverySecond = false
			timeUnitAsInt, err := strconv.ParseInt(timeUnit, 10, 64)
			if err != nil {
				continue
			}
			if timeUnitAsInt != timeValues[i] {
				canRunCallback = false
				executedAlready = false
				break
			}
			canRunCallback = true
		}

		if canRunCallback && (executeEverySecond || !executedAlready) {
			err := callback()

			if err != nil && config != nil {
				log.Printf("Error executing callback: %v\n", err)
				time.Sleep(config.RetriesAfter)
				for i := 0; i < config.Retries; i++ {
					log.Printf("Retrying to execute callback (%d)", i+1)
					err := callback()
					if err == nil {
						break
					}
					log.Printf("Error executing callback: %v\n", err)
					time.Sleep(config.RetriesAfter)
				}
			} else if err != nil {
				log.Printf("Error executing callback: %v\n", err)
			}
			executedAlready = true
		}
		time.Sleep(time.Second)
	}
}

func New() *Cron {
	return &Cron{funcs: map[string]Callback{}}
}
