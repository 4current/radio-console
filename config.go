package main

import (
	"encoding/json"
	"os"
)

type RadioConfig struct {
	RigID      string `json:"rig_id"`
	TCPHost    string `json:"tcp_host,omitempty"`
	TCPPort    string `json:"tcp_port,omitempty"`
	SerialPort string `json:"serial_port,omitempty"`
	BaudRate   int    `json:"baud_rate,omitempty"`
	RigctlFreq string `json:"rigctl_freq,omitempty"`
}

type Config struct {
	Radios []RadioConfig `json:"radios"`
}

func LoadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := &Config{}
	err = decoder.Decode(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
