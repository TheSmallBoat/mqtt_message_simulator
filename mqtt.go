package main

import (
	"crypto/rand"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

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

type MqttPublishClient struct {
	MqttClient MQTT.Client
	Opts       *MQTT.ClientOptions
	PubTopic   string
	Qos        uint8
	DevChan    *chan bool
}

func newMqttOptions(cfg *Config, useMonitor bool) *MQTT.ClientOptions {
	if cfg.General.Debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	if useMonitor {
		var brokerUri = fmt.Sprintf("%s://%s:%d", cfg.MonitorMqtt.Scheme, cfg.MonitorMqtt.Hostname, cfg.MonitorMqtt.Port)
		log.Infof("Monitor Broker URI: %s", brokerUri)
		return initMqttOptions(brokerUri, cfg.MonitorMqtt.Username, cfg.MonitorMqtt.Password, cfg.MonitorMqtt.Cleansession, cfg.MonitorMqtt.Pingtimeout, cfg.MonitorMqtt.Keepalive)
	} else {
		var brokerUri = fmt.Sprintf("%s://%s:%d", cfg.SimulatorMqtt.Scheme, cfg.SimulatorMqtt.Hostname, cfg.SimulatorMqtt.Port)
		return initMqttOptions(brokerUri, cfg.SimulatorMqtt.Username, cfg.SimulatorMqtt.Password, cfg.SimulatorMqtt.Cleansession, cfg.SimulatorMqtt.Pingtimeout, cfg.SimulatorMqtt.Keepalive)
	}
}

func initMqttOptions(brokerUri string, username string, password string, cleansession bool, pingtimeout uint8, keepalive uint16) *MQTT.ClientOptions {
	opts := MQTT.NewClientOptions()

	opts.SetAutoReconnect(true)
	opts.SetCleanSession(cleansession)
	opts.SetPingTimeout(time.Duration(pingtimeout) * time.Second)
	opts.SetConnectTimeout(time.Duration(keepalive) * time.Second)

	opts.AddBroker(brokerUri)
	if username != "" {
		opts.SetUsername(username)
	}
	if password != "" {
		opts.SetPassword(password)
	}

	return opts
}

// getRandomClientId returns randomized ClientId.
func getRandomClientId(clientName string, maxClientIdLen uint8) string {
	const alphaNum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, maxClientIdLen)
	_, _ = rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphaNum[b%byte(len(alphaNum))]
	}
	return clientName + "-" + string(bytes)
}

// with Connects connect to the MQTT broker with Options.
func NewSimulatorMqttClient(cfg *Config, deviceIdNum uint16, dc *chan bool) (*MqttPublishClient, error) {
	clientId := getRandomClientId(cfg.SimulatorTopic.Clientnameshort, cfg.SimulatorTopic.Deviceidmaxlen)
	groupBit := cfg.SimulatorTopic.Devicegroupbit
	pubTopic := fmt.Sprintf("%s/%s/%s/%s/%d", cfg.SimulatorMqtt.Topicroot, cfg.SimulatorTopic.Clientname, cfg.SimulatorTopic.Devicelocation, cfg.SimulatorTopic.Devicetype, deviceIdNum>>groupBit)

	opts := newMqttOptions(cfg, false)
	opts.SetClientID(clientId)
	mpc := &MqttPublishClient{Opts: opts, PubTopic: pubTopic, Qos: cfg.MonitorMqtt.Qos, DevChan: dc}
	err := mpc.setMqttPublishClientHandler(mpc.SimulatorOnConnect, mpc.SimulatorConnectionLost)
	return mpc, err
}

func NewMonitorMqttClient(cfg *Config) (*MqttPublishClient, error) {
	clientId := "sim-mon-" + cfg.SimulatorTopic.Clientnameshort
	pubTopic := fmt.Sprintf("%s/%s", cfg.MonitorMqtt.Topicroot, cfg.SimulatorTopic.Clientname)

	opts := newMqttOptions(cfg, true)
	opts.SetClientID(clientId)
	mpc := &MqttPublishClient{Opts: opts, PubTopic: pubTopic, Qos: cfg.MonitorMqtt.Qos, DevChan: nil}
	err := mpc.setMqttPublishClientHandler(mpc.MonitorOnConnect, mpc.MonitorConnectionLost)

	log.Infof("monitor: %s, topic: %s, qos: %d ", clientId, mpc.PubTopic, mpc.Qos)
	return mpc, err
}

func (m *MqttPublishClient) setMqttPublishClientHandler(onConn MQTT.OnConnectHandler, conLostHandler MQTT.ConnectionLostHandler) error {
	m.Opts.SetOnConnectHandler(onConn)
	m.Opts.SetConnectionLostHandler(conLostHandler)

	client := MQTT.NewClient(m.Opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	m.MqttClient = client
	return nil
}

func (m *MqttPublishClient) MonitorOnConnect(client MQTT.Client) {
	log.Debugf("Monitor [%s] mqtt connected", m.Opts.ClientID)
}

func (m *MqttPublishClient) MonitorConnectionLost(client MQTT.Client, reason error) {
	log.Errorf("Monitor mqtt disconnected: %s", reason)
}

func (m *MqttPublishClient) SimulatorOnConnect(client MQTT.Client) {
	*m.DevChan <- true
	log.Debugf("Virtual device: [%s] connected to mqtt broker.", m.Opts.ClientID)
}

func (m *MqttPublishClient) SimulatorConnectionLost(client MQTT.Client, reason error) {
	*m.DevChan <- false
	log.Errorf("Virtual device: [%s] has lost its connection to the mqtt broker: %s", m.Opts.ClientID, reason)
}

func (m *MqttPublishClient) Disconnect() {
	if m.MqttClient.IsConnected() {
		m.MqttClient.Disconnect(20)
		*m.DevChan <- false
		log.Debugf("[%s] mqtt disconnected", m.Opts.ClientID)
	}
}
