package main

import (
	"fmt"
)

type Config struct {
	General            GeneralConf            `gcfg:"general"`
	SimulatorMqtt      SimulatorMqttConf      `gcfg:"simulator-mqtt"`
	SimulatorTopic     SimulatorTopicConf     `gcfg:"simulator-topic"`
	SimulatorMessage   SimulatorMessageConf   `gcfg:"simulator-message"`
	SimulatorJsonStyle SimulatorJsonStyleConf `gcfg:"simulator-json"`
	MonitorMqtt        MonitorMqttConf        `gcfg:"monitor-mqtt"`
	MonitorInfo        MonitorInfoConf        `gcfg:"monitor-info"`
}
type GeneralConf struct {
	Debug         bool
	Sleepinterval uint
}

type SimulatorMqttConf struct {
	Scheme       string
	Hostname     string
	Port         uint
	Cleansession bool
	Qos          uint8
	Pingtimeout  uint8
	Keepalive    uint16
	Username     string
	Password     string
	Topicroot    string
}

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

type SimulatorJsonStyleConf struct {
	Enableclientid   bool
	Enablemessageid  bool
	Enabeldeviceno   bool
	Enabeunixtime    bool
	Enablestringtime bool
	Jsonformat       string
}

type MonitorMqttConf struct {
	Scheme       string
	Hostname     string
	Port         uint
	Cleansession bool
	Qos          uint8
	Pingtimeout  uint8
	Keepalive    uint16
	Username     string
	Password     string
	Topicroot    string
}

type MonitorInfoConf struct {
	Buffersize      uint
	PublishInterval uint
}

func (cfg *Config) GetConfInfo() string {
	info := fmt.Sprintf("Configuration information ... \n [general] => %+v\n [simulator-mqtt] => %+v\n [simulator-topic] => %+v\n [simulator-message] => %+v\n [simulator-json] => %+v\n [monitor-mqtt] => %+v\n [monitor-info] => %+v\n ", cfg.General, cfg.SimulatorMqtt, cfg.SimulatorTopic, cfg.SimulatorMessage, cfg.SimulatorJsonStyle, cfg.MonitorMqtt, cfg.MonitorInfo)
	return info
}
