package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"

	"fyne.io/fyne/v2"
	"github.com/tarm/serial"

	// "fyne.io/fyne"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// Define SerialPort interface and serialPortOpener
type SerialPort interface {
	Write(p []byte) (n int, err error)
	Read(p []byte) (n int, err error)
	Close() error
}

var serialPortOpener SerialPortOpener = &defaultSerialPortOpener{}

type defaultSerialPortOpener struct{}

func (o *defaultSerialPortOpener) OpenPort(c *serial.Config) (SerialPort, error) {
	return serial.OpenPort(c)
}

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

	// Button for setting frequency
	setFrequencyButton := widget.NewButton("Set Frequency", func() {
		selectedRadio := getSelectedRadio(config, radioSelect.Selected)
		if selectedRadio != nil {
			freq := freqEntry.Text
			switch selectedRadio.ConnType {
			case "TCP":
				sendTCPCommand(selectedRadio.TCPHost, selectedRadio.TCPPort, "FA"+freq+";")
			case "Serial":
				sendSerialCommand(selectedRadio.SerialPort, selectedRadio.BaudRate, "FA"+freq+";")
			case "rigctl":
				sendRigctlCommand(freq)
			default:
				fmt.Println("Invalid connection type selected")
			}
		}
	})

	// Settings menu
	settingsButton := widget.NewButton("Settings", func() {
		showSettingsDialog(w, config)
	})

	// UI Layout
	w.SetContent(container.NewVBox(
		widget.NewLabel("Radio Console"),
		radioSelect,
		freqEntry,
		setFrequencyButton,
		settingsButton,
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

func showSettingsDialog(w fyne.Window, config *Config) {
	rigIDEntry := widget.NewEntry()
	connTypeEntry := widget.NewEntry()
	tcpHostEntry := widget.NewEntry()
	tcpPortEntry := widget.NewEntry()
	serialPortEntry := widget.NewEntry()
	baudRateEntry := widget.NewEntry()
	rigctlFreqEntry := widget.NewEntry()

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Rig ID", Widget: rigIDEntry},
			{Text: "Connection Type", Widget: connTypeEntry},
			{Text: "TCP Host", Widget: tcpHostEntry},
			{Text: "TCP Port", Widget: tcpPortEntry},
			{Text: "Serial Port", Widget: serialPortEntry},
			{Text: "Baud Rate", Widget: baudRateEntry},
			{Text: "Rigctl Frequency", Widget: rigctlFreqEntry},
		},
		OnSubmit: func() {
			newRadio := RadioConfig{
				RigID:      rigIDEntry.Text,
				ConnType:   connTypeEntry.Text,
				TCPHost:    tcpHostEntry.Text,
				TCPPort:    tcpPortEntry.Text,
				SerialPort: serialPortEntry.Text,
				BaudRate:   parseInt(baudRateEntry.Text),
				RigctlFreq: rigctlFreqEntry.Text,
			}
			config.Radios = append(config.Radios, newRadio)
			saveConfig(config)
			dialog.ShowInformation("Settings", "Settings saved successfully", w)
		},
	}

	dialog.ShowForm("Settings", "Save", "Cancel", form.Items, func(bool) { form.OnSubmit() }, w)
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func saveConfig(config *Config) {
	file, err := os.Create("config.json")
	if err != nil {
		log.Fatalf("Failed to save config: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(config)
	if err != nil {
		log.Fatalf("Failed to encode config: %v", err)
	}

	encoder.SetIndent("", "  ")
	err = encoder.Encode(config)
	if err != nil {
		log.Fatalf("Failed to encode config: %v", err)
	}
}
