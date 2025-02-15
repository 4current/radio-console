package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/tarm/serial"
)

var execCommand = exec.Command

var listener net.Listener

type MockSerialPort struct {
	mock.Mock
}

func (m *MockSerialPort) Write(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockSerialPort) Read(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockSerialPort) Close() error {
	return m.Called().Error(0)
}

type MockSerialPortOpener struct {
	mockPort *MockSerialPort
}

func (o *MockSerialPortOpener) OpenPort(c *serial.Config) (SerialPort, error) {
	return o.mockPort, nil
}

func TestMain(m *testing.M) {
	// Load configuration
	config, err := LoadConfig("config.json")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Setup: Start a test TCP server
	listener, err = net.Listen("tcp", fmt.Sprintf("%s:%s", config.Radios[0].TCPHost, config.Radios[0].TCPPort))
	if err != nil {
		fmt.Printf("Failed to start test server: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Teardown: Close the listener
	listener.Close()

	os.Exit(code)
}

func TestSendTCPCommand(t *testing.T) {
	// Load configuration
	config, err := LoadConfig("config.json")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	errChan := make(chan error, 1)

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			errChan <- err
			return
		}
		defer conn.Close()

		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			errChan <- err
			return
		}

		expectedCommand := "FA12345678;\n"
		if string(buf[:n]) != expectedCommand {
			errChan <- fmt.Errorf("expected command %q, got %q", expectedCommand, string(buf[:n]))
			return
		}

		response := "OK\n"
		conn.Write([]byte(response))
		errChan <- nil
	}()

	// Allow some time for the server to start
	time.Sleep(100 * time.Millisecond)

	// Test the sendTCPCommand function
	sendTCPCommand(config.Radios[0].TCPHost, config.Radios[0].TCPPort, "FA12345678;")

	select {
	case err := <-errChan:
		if err != nil {
			t.Fatalf("Error in goroutine: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Test timed out")
	}
}

func TestSendTCPCommandError(t *testing.T) {
	// Load configuration
	config, err := LoadConfig("config.json")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test error handling in sendTCPCommand
	sendTCPCommand("invalid_host", config.Radios[0].TCPPort, "FA12345678;")
}

func TestSendSerialCommand(t *testing.T) {
	// Load configuration
	config, err := LoadConfig("config.json")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Mock serial port
	mockPort := new(MockSerialPort)
	mockPort.On("Write", []byte("FA12345678;\n")).Return(len("FA12345678;\n"), nil)
	mockPort.On("Close").Return(nil)

	// Replace the serialPortOpener with a mock
	originalSerialPortOpener := serialPortOpener
	serialPortOpener = &MockSerialPortOpener{mockPort}
	defer func() { serialPortOpener = originalSerialPortOpener }()

	// Test the sendSerialCommand function
	sendSerialCommand(config.Radios[0].SerialPort, config.Radios[0].BaudRate, "FA12345678;")

	// Assert that the mock was called as expected
	mockPort.AssertExpectations(t)
}

func TestSendSerialCommandError(t *testing.T) {
	// Load configuration
	config, err := LoadConfig("config.json")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Mock serial port
	mockPort := new(MockSerialPort)
	mockPort.On("Write", []byte("FA12345678;\n")).Return(0, errors.New("write error"))
	mockPort.On("Close").Return(nil)

	// Replace the serialPortOpener with a mock
	originalSerialPortOpener := serialPortOpener
	serialPortOpener = &MockSerialPortOpener{mockPort}
	defer func() { serialPortOpener = originalSerialPortOpener }()

	// Test the sendSerialCommand function
	sendSerialCommand(config.Radios[0].SerialPort, config.Radios[0].BaudRate, "FA12345678;")

	// Assert that the mock was called as expected
	mockPort.AssertExpectations(t)
}

func TestSendRigctlCommand(t *testing.T) {
	// Load configuration
	config, err := LoadConfig("config.json")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Mock rigctl command
	freq := config.Radios[0].RigctlFreq

	// Create a temporary file to mock rigctl output
	tmpfile, err := os.CreateTemp("", "rigctl")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Write mock output to the temp file
	mockOutput := "rigctl output\n"
	if _, err := tmpfile.Write([]byte(mockOutput)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Mock exec.Command to use the temp file
	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("cat", tmpfile.Name())
	}

	// Test the sendRigctlCommand function
	sendRigctlCommand(freq)

	// Restore exec.Command
	execCommand = exec.Command
}

func TestSendRigctlCommandError(t *testing.T) {
	// Mock exec.Command to return an error
	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("false")
	}

	// Test the sendRigctlCommand function
	sendRigctlCommand("12345678")

	// Restore exec.Command
	execCommand = exec.Command
}
