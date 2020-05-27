package main

import (
	"fmt"
	"time"
)

type Monitor struct {
	MqttPublishClient *MqttPublishClient
	MessageTaskNum    uint
	PublishInterval   uint
	PubFailNum        uint
	PubSucceedNum     uint

	SimPubChan chan bool
	SimFailed  uint
	SimSucceed uint
	SimMPS     uint

	SimDevChan  chan bool // Device Channel of the simulator.
	DevConnNum  uint      // Total number of the connected devices.
	DevConnPS   uint      // Device connections per second.
	DevDisConPS uint      // Device disconnections per second.
}

func NewMonitor(cfg *Config) (*Monitor, error) {
	mmc, err := NewMonitorMqttClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("monitor mqtt publish client init err: %s", err)
	}
	return &Monitor{
		MqttPublishClient: mmc,
		MessageTaskNum:    cfg.SimulatorMessage.Messagenumber * cfg.SimulatorTopic.Devicenumber,
		PublishInterval:   cfg.MonitorInfo.PublishInterval,
		PubFailNum:        0,
		PubSucceedNum:     0,
		SimPubChan:        make(chan bool, cfg.MonitorInfo.Buffersize),
		SimFailed:         0,
		SimSucceed:        0,
		SimMPS:            0,
		SimDevChan:        make(chan bool, cfg.MonitorInfo.Buffersize),
		DevConnNum:        0,
		DevConnPS:         0,
		DevDisConPS:       0,
	}, nil
}

func (mon *Monitor) publishInfo(payload string) {
	topic := mon.MqttPublishClient.PubTopic
	qos := byte(mon.MqttPublishClient.Qos)
	if token := mon.MqttPublishClient.MqttClient.Publish(topic, qos, false, payload); token.Wait() && token.Error() != nil {
		mon.PubFailNum++
	}
	mon.PubSucceedNum++
}

//func (mon *Monitor) Start(wg *sync.WaitGroup) {
func (mon *Monitor) Start() {
	defer mon.MqttPublishClient.Disconnect()

	ticker := time.NewTicker(time.Duration(mon.PublishInterval) * time.Second)
	begin := time.Now()

	for {
		select {
		case <-ticker.C:
			intv := mon.PublishInterval
			runtime := time.Now().Sub(begin).Seconds()
			messagenum := mon.SimSucceed + mon.SimFailed
			progressrate := 100 * messagenum / mon.MessageTaskNum
			avgmps := float64(messagenum) / runtime
			fmtf := "{\"MessagePerSec\":%d,\"MessageSucceed\":%d,\"MessageFailed\":%d,\"Runtime(s)\":%.1f,\"AvgPeriodMessagePerSec\":%.2f,\"TargetMessageSum\":%d,\"Progress(%%)\":%d,\"OnlineDevice\":%d,\"ConnectionPerSec\":%d,\"DisConnectionPerSec\":%d,\"MonitorPublishSucceed\":%d,\"MonitorPublishFailed\":%d}"
			payload := fmt.Sprintf(fmtf, mon.SimMPS/intv, mon.SimSucceed, mon.SimFailed, runtime, avgmps, mon.MessageTaskNum, progressrate, mon.DevConnNum, mon.DevConnPS/intv, mon.DevDisConPS/intv, mon.PubSucceedNum, mon.PubFailNum)
			mon.publishInfo(payload)
			mon.SimMPS = 0
			mon.DevConnPS = 0
			mon.DevDisConPS = 0
		case flagPub := <-mon.SimPubChan:
			mon.SimMPS++
			if !flagPub {
				mon.SimFailed++
			} else {
				mon.SimSucceed++
			}
		case flagDev := <-mon.SimDevChan:
			if flagDev {
				mon.DevConnNum++
				mon.DevConnPS++
			} else {
				mon.DevConnNum--
				mon.DevDisConPS++
			}
		}
	}
}
