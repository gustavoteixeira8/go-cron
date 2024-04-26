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
	ID           string
	Retries      int
	RetriesAfter time.Duration
	Callback     Callback
	Close        chan bool
	Debug        bool
}

type Cron struct {
	config map[string]*CronConfig
	wg     sync.WaitGroup
}

/*
cron.AddFunc adiciona um cron para ser executado

---

Parameters:

  - cronValue:

    S M H D M Y

    S = Second (0-59)

    M = Minute (0-59)

    H = Hour (0-23)

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
func (c *Cron) AddFunc(cronValue string, config *CronConfig) {
	if c.config == nil {
		c.config = map[string]*CronConfig{}
	}

	cronValue = fmt.Sprintf("%s %s", cronValue, uuid.NewString())

	c.config[config.ID] = config

	c.wg.Add(1)
	go c.processCron(cronValue, config)

}

// Inicia todos os crons e trava o processo.
func (c *Cron) Close(id string) {

	config := c.config[id]

	config.Close <- true

	c.wg.Done()
}

// Inicia todos os crons e trava o processo.
func (c *Cron) Wait() {
	c.wg.Wait()
}

func (c *Cron) isToExecEverySecond(cronValue string) bool {
	cSplit := strings.Split(cronValue, " ")
	cSplit = cSplit[:len(cSplit)-1]
	isEverySecond := true
	for _, str := range cSplit {
		if str != "*" {
			isEverySecond = false
			break
		}
	}
	return isEverySecond
}

// Verifica se o cronValue corresponde a hora atual e executa a callback.
func (c *Cron) processCron(cronValue string, config *CronConfig) {
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

	ticker := time.NewTicker(time.Second)

	executedAlready := false
	isToExecEverySecond := c.isToExecEverySecond(cronValue)
	for {
		select {
		case <-ticker.C:
			{
				canRunCallback := false

				if !isToExecEverySecond {
					now := time.Now()
					timeValues := []int64{
						int64(now.Second()),
						int64(now.Minute()),
						int64(now.Hour()),
						int64(now.Day()),
						int64(now.Month()),
						int64(now.Year()),
					}
					for i, timeUnit := range cronSplitted {
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
				} else {
					canRunCallback = true
					executedAlready = false
				}

				if canRunCallback && !executedAlready {
					err := config.Callback()

					if err != nil {
						log.Printf("Error executing callback: %v\n", err)

						if config != nil {
							time.Sleep(config.RetriesAfter)

							for i := 0; i < config.Retries; i++ {
								log.Printf("Retrying to execute callback (%d)", i+1)

								err := config.Callback()
								if err == nil {
									break
								}

								log.Printf("Error executing callback: %v\n", err)
								time.Sleep(config.RetriesAfter)
							}
						}
					}

					executedAlready = true
				}

			}
		case <-config.Close:
			return
		}

	}
}

func New() *Cron {
	return &Cron{config: make(map[string]*CronConfig)}
}
