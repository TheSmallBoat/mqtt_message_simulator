package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type SimulatorTopicConf struct {
	Clientname      string
	Clientnameshort string
	Devicelocation  string
	Devicetype      string
	Devicenumber    uint
	Devicegroupbit  uint8
	Deviceidmaxlen  uint8
	Taskinterval    uint8
}
type SimulatorMessageConf struct {
	Messagenumber   uint
	Minimuminterval uint8
	Maximuminterval uint8
}

type Simulator struct {
	MqttPublishClient *MqttPublishClient
	Monitor           *Monitor

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
		DeviceIdNum:       deviceIdNum,
		MessageTotal:      cfg.SimulatorMessage.Messagenumber,
		Interval:          uint(intv),
		PublishFailNum:    0,
		PublishSuccedNum:  0,
	}, nil
}

func (sim *Simulator) publishInfo(payload string) {
	topic := sim.MqttPublishClient.PubTopic
	qos := byte(sim.MqttPublishClient.Qos)
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
		cid := sim.MqttPublishClient.Opts.ClientID
		msg := fmt.Sprintf("{\"device_no\":\"%d\",\"message_total\":%d,\"interval\":%d,\"fail_num\":\"%d\",\"succed_num\":\"%d\"}", sim.DeviceIdNum, sim.MessageTotal, sim.Interval, sim.PublishFailNum, sim.PublishSuccedNum)
		payload := fmt.Sprintf("{\"c_id\":\"%s\",\"c_time\":%d,\"m_id\":%d,\"m_body\":%s}", cid, time.Now().Unix(), mid, msg)
		sim.publishInfo(payload)
	}

	sim.MqttPublishClient.Disconnect()
}
