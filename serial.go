package main

import (
	"github.com/tarm/serial"
)

type SerialPort interface {
	Write(p []byte) (n int, err error)
	Read(p []byte) (n int, err error)
	Close() error
}

type SerialPortOpener interface {
	OpenPort(c *serial.Config) (SerialPort, error)
}

type RealSerialPort struct {
	port *serial.Port
}

func (r *RealSerialPort) Write(p []byte) (n int, err error) {
	return r.port.Write(p)
}

func (r *RealSerialPort) Read(p []byte) (n int, err error) {
	return r.port.Read(p)
}

func (r *RealSerialPort) Close() error {
	return r.port.Close()
}

type RealSerialPortOpener struct{}

func (o *RealSerialPortOpener) OpenPort(c *serial.Config) (SerialPort, error) {
	port, err := serial.OpenPort(c)
	if err != nil {
		return nil, err
	}
	return &RealSerialPort{port: port}, nil
}

var serialPortOpener SerialPortOpener = &RealSerialPortOpener{}
