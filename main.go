package main

import (
	"fmt"
	"os/exec"

	"fyne.io/fyne/v2"
	"github.com/tarm/serial"

	// "fyne.io/fyne"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func sendSerialCommand(port string, baud int, command string) {
	c := &serial.Config{Name: port, Baud: baud}
	s, err := serial.OpenPort(c)
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

func main() {
	a := app.New()
	w := a.NewWindow("Radio Console")
	w.Resize(fyne.NewSize(400, 300))

	// Frequency Entry
	freqEntry := widget.NewEntry()
	freqEntry.SetPlaceHolder("Enter Frequency (Hz)")

	// Button to Send Frequency to rigctl
	setFreqButton := widget.NewButton("Set Frequency", func() {
		freq := freqEntry.Text
		cmd := exec.Command("rigctl", "-m", "1", "F", freq) // Adjust model (-m) for your radio
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("rigctl output:", string(output))
		}
	})

	// UI Layout
	w.SetContent(container.NewVBox(
		widget.NewLabel("Radio Console"),
		freqEntry,
		setFreqButton,
	))

	w.ShowAndRun()
}
