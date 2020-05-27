package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mattn/go-colorable"
	"github.com/urfave/cli"
)

func init() {
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{ForceColors: true, FullTimestamp: true})
	log.SetOutput(colorable.NewColorableStdout())
}

func processConfiguration(c *cli.Context) (*Config, error) {
	path := c.String("c")
	cfg, err := LoadConf(path)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	log.Info(fmt.Sprintf("General Configuration: %+v", cfg.General))
	log.Info(fmt.Sprintf("SimulatorMqtt Configuration: %+v", cfg.SimulatorMqtt))
	log.Info(fmt.Sprintf("SimulatorTopic Configuration: %+v", cfg.SimulatorTopic))
	log.Info(fmt.Sprintf("SimulatorMessage Configuration: %+v", cfg.SimulatorMessage))
	log.Info(fmt.Sprintf("MonitorMqtt Configuration: %+v", cfg.MonitorMqtt))
	log.Info(fmt.Sprintf("MonitorInfo Configuration: %+v", cfg.MonitorInfo))
	return cfg, nil
}

func showConfiguration(c *cli.Context) {
	_, _ = processConfiguration(c)
}

func runSimulator(cont *cli.Context) {
	cfg, _ := processConfiguration(cont)
	mon, err := NewMonitor(cfg)
	if err != nil {
		log.Fatal(err)
	}

	go mon.Start()
	time.Sleep(time.Duration(cfg.General.Sleepinterval) * time.Millisecond)

	wg := sync.WaitGroup{}
	for i := uint16(0); i < uint16(cfg.SimulatorTopic.Devicenumber); i++ {
		wg.Add(1)
		time.Sleep(time.Duration(cfg.SimulatorTopic.Taskinterval) * time.Millisecond)
		var sim, err = NewSimulator(cfg, i, mon)
		if err != nil {
			log.Fatal(err)
		}
		go sim.StartTask(&wg)
	}
	wg.Wait()
	time.Sleep(time.Duration(2*cfg.MonitorInfo.PublishInterval) * time.Second)
	log.Infof("There are [%d] virtual devices that have completed all message simulation tasks.", cfg.SimulatorTopic.Devicenumber)
}

func main() {
	app := cli.NewApp()
	app.Name = "message simulator"
	app.Version = "19.06.19"
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		{
			Name: "Abe Cai",
		},
	}
	app.Usage = "A command-line simulator for generating messages published to the MQTT broker."

	app.Commands = []cli.Command{
		{
			Name:   "run",
			Usage:  "generate the messages.",
			Action: runSimulator,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "config, c",
					Usage: "config file path",
					Value: "~/.message-simulator.ini",
				},
			},
		},
		{
			Name:   "show",
			Usage:  "print the configuration information.",
			Action: showConfiguration,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "config, c",
					Usage: "config file path",
					Value: "~/.message-simulator.ini",
				},
			},
		},
	}

	_ = app.Run(os.Args)
}
