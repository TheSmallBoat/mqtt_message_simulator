package main

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
)

type Simulator struct {
	MqttPublishClient *MqttPublishClient
	Monitor           *Monitor

	JsonStyleConf *SimulatorJsonStyleConf

	DeviceIdNum uint16

	MessageTotal uint
	Interval     uint

	PublishFailNum   uint
	PublishSuccedNum uint
}

func NewSimulator(cfg *Config, deviceIdNum uint16, mon *Monitor) (*Simulator, error) {
	smc, err := NewSimulatorMqttClient(cfg, deviceIdNum, &mon.SimDevChan)
	if err != nil {
		return nil, fmt.Errorf("simulator mqtt publish client init err: %s", err)
	}

	intv := int(cfg.SimulatorMessage.Minimuminterval) + rand.Intn(int(cfg.SimulatorMessage.Maximuminterval-cfg.SimulatorMessage.Minimuminterval+1))
	return &Simulator{
		MqttPublishClient: smc,
		Monitor:           mon,
		JsonStyleConf:     &cfg.SimulatorJsonStyle,
		DeviceIdNum:       deviceIdNum,
		MessageTotal:      cfg.SimulatorMessage.Messagenumber,
		Interval:          uint(intv),
		PublishFailNum:    0,
		PublishSuccedNum:  0,
	}, nil
}

func (sim *Simulator) publishInfo(payload string) {
	topic := sim.MqttPublishClient.PubTopic
	qos := sim.MqttPublishClient.Qos
	if token := sim.MqttPublishClient.MqttClient.Publish(topic, qos, false, payload); token.Wait() && token.Error() != nil {
		sim.Monitor.SimPubChan <- false
		sim.PublishFailNum++
		//panic(token.Error())
	}
	sim.Monitor.SimPubChan <- true
	sim.PublishSuccedNum++

}

func (sim *Simulator) StartTask(wg *sync.WaitGroup) {
	defer wg.Done()

	for mid := 0; mid < int(sim.MessageTotal); mid++ {
		time.Sleep(time.Duration(sim.Interval) * time.Second)

		//msg := fmt.Sprintf("{\"device_no\":\"%d\",\"message_total\":%d,\"interval\":%d,\"fail_num\":\"%d\",\"succed_num\":\"%d\"}", sim.DeviceIdNum, sim.MessageTotal, sim.Interval, sim.PublishFailNum, sim.PublishSuccedNum)
		payload := sim.JsonStyleConf.Jsonformat
		if sim.JsonStyleConf.Enableclientid {
			cid := sim.MqttPublishClient.Opts.ClientID
			payload = strings.ReplaceAll(payload, "#CLIENT_ID#", cid)
		}
		if sim.JsonStyleConf.Enablemessageid {
			payload = strings.ReplaceAll(payload, "#MESSAGE_ID#", fmt.Sprintf("%d", mid))
		}
		if sim.JsonStyleConf.Enabeldeviceno {
			payload = strings.ReplaceAll(payload, "#DEVICE_NO#", fmt.Sprintf("%d", sim.DeviceIdNum))
		}
		if sim.JsonStyleConf.Enabeunixtime {
			payload = strings.ReplaceAll(payload, "#UNIX_TIME#", fmt.Sprintf("%d", time.Now().Unix()))
		}
		if sim.JsonStyleConf.Enablestringtime {
			payload = strings.ReplaceAll(payload, "#STRING_TIME#", time.Now().Format("2006-01-02T15:04:05Z"))
		}
		sim.publishInfo(payload)
	}

	sim.MqttPublishClient.Disconnect()
}
