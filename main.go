package main

import (
	"fmt"
	"log"
	"net"
	"os/exec"

	"fyne.io/fyne/v2"
	"github.com/tarm/serial"

	// "fyne.io/fyne"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Function to send a command via TCP to the Kenwood TS-890S
func sendTCPCommand(host string, port string, command string) {
	conn, err := net.Dial("tcp", host+":"+port)
	if err != nil {
		fmt.Println("Error connecting to radio:", err)
		return
	}
	defer conn.Close()

	_, err = conn.Write([]byte(command + "\n"))
	if err != nil {
		fmt.Println("Error sending command:", err)
		return
	}

	// Read response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	fmt.Println("Radio Response:", string(buf[:n]))
}

func sendSerialCommand(port string, baud int, command string) {
	c := &serial.Config{Name: port, Baud: baud}
	s, err := serialPortOpener.OpenPort(c)
	if err != nil {
		fmt.Println("Error opening serial port:", err)
		return
	}
	defer s.Close()

	_, err = s.Write([]byte(command + "\n"))
	if err != nil {
		fmt.Println("Error writing to serial port:", err)
	}
}

// Function to send a command using rigctl (Hamlib)
func sendRigctlCommand(freq string) {
	cmd := exec.Command("rigctl", "-m", "1", "F", freq) // Adjust model (-m) for your radio
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("rigctl output:", string(output))
	}
}

func main() {
	// Load configuration
	config, err := LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	a := app.New()
	w := a.NewWindow("Radio Console")
	w.Resize(fyne.NewSize(400, 300))

	// Frequency Entry
	freqEntry := widget.NewEntry()
	freqEntry.SetPlaceHolder("Enter Frequency (Hz)")

	// Radio selection
	radioSelect := widget.NewSelect([]string{}, func(value string) {
		// Update UI based on selected radio
	})
	for _, radio := range config.Radios {
		radioSelect.Options = append(radioSelect.Options, radio.RigID)
	}

	// Button for TCP Control (TS-890S)
	tcpButton := widget.NewButton("Set Frequency (TCP)", func() {
		selectedRadio := getSelectedRadio(config, radioSelect.Selected)
		if selectedRadio != nil {
			freq := freqEntry.Text
			sendTCPCommand(selectedRadio.TCPHost, selectedRadio.TCPPort, "FA"+freq+";") // Example Kenwood CAT command
		}
	})

	// Button for Serial Control
	serialButton := widget.NewButton("Set Frequency (Serial)", func() {
		selectedRadio := getSelectedRadio(config, radioSelect.Selected)
		if selectedRadio != nil {
			freq := freqEntry.Text
			sendSerialCommand(selectedRadio.SerialPort, selectedRadio.BaudRate, "FA"+freq+";")
		}
	})

	// Button for rigctl Control
	rigctlButton := widget.NewButton("Set Frequency (rigctl)", func() {
		selectedRadio := getSelectedRadio(config, radioSelect.Selected)
		if selectedRadio != nil {
			freq := freqEntry.Text
			sendRigctlCommand(freq)
		}
	})

	// UI Layout
	w.SetContent(container.NewVBox(
		widget.NewLabel("Ham Radio Control"),
		radioSelect,
		freqEntry,
		tcpButton,
		serialButton,
		rigctlButton,
	))

	w.ShowAndRun()
}

func getSelectedRadio(config *Config, rigID string) *RadioConfig {
	for _, radio := range config.Radios {
		if radio.RigID == rigID {
			return &radio
		}
	}
	return nil
}
