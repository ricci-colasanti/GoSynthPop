package main

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// UIUpdate struct for messages
type UIUpdate struct {
	Text string
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Channel Toy Example")

	// Create our UI
	statusLabel := widget.NewLabel("Ready to start...")

	// Create channel for UI updates
	uiUpdates := make(chan UIUpdate, 10)

	// Start the UI update handler (runs forever)
	go func() {
		for update := range uiUpdates {
			// Use fyne.Do for thread-safe UI updates
			fyne.Do(func() {
				statusLabel.SetText(update.Text)
			})
		}
	}()

	// Toy worker function that uses the channel
	toyWorker := func(updates chan<- UIUpdate) {
		updates <- UIUpdate{Text: "🚀 Starting toy worker..."}

		for i := 1; i <= 5; i++ {
			time.Sleep(1 * time.Second)
			updates <- UIUpdate{Text: fmt.Sprintf("🔧 Working... step %d/5", i)}
		}

		updates <- UIUpdate{Text: "✅ Toy work complete!"}
	}

	// Button that starts the worker in a goroutine
	startButton := widget.NewButton("Start Toy Worker", func() {
		go toyWorker(uiUpdates)
	})

	content := container.NewVBox(statusLabel, startButton)
	myWindow.SetContent(content)
	myWindow.ShowAndRun()
	close(uiUpdates)
}
