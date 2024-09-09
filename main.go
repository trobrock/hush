package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"golang.design/x/hotkey"
	"golang.design/x/hotkey/mainthread"
)

var (
	lastPressTime time.Time
	mode          string
	isHeld        bool
)

const (
	doublePressDelay = 300 * time.Millisecond
	ModePTT          = "PTT"
	ModePTS          = "PTS"
)

func main() { mainthread.Init(fn) } // Required for MacOS

func fn() {
	setInitialMode()

	hk := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl}, hotkey.KeyTab)
	err := hk.Register()
	if err != nil {
		log.Fatalf("hotkey: failed to register hotkey: %v", err)
		return
	}
	defer hk.Unregister()

	log.Printf("hotkey: %v is registered\n", hk)

	// Channel to receive OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// Main loop
	for {
		select {
		case <-sigChan:
			log.Println("Received interrupt signal. Exiting...")
			return
		case <-hk.Keydown():
			handleKeyDown()
		case <-hk.Keyup():
			handleKeyUp()
		}
	}
}

func handleKeyDown() {
	now := time.Now()

	if time.Since(lastPressTime) < doublePressDelay {
		handleDoublePress()
	} else {
		startKeyPress()
	}
	lastPressTime = now
	setIsHeld(true)
}

func handleKeyUp() {
	endKeyPress()
	setIsHeld(false)
}

func startKeyPress() {
	toggleMicrophone()
}

func endKeyPress() {
	toggleMicrophone()
}

func handleDoublePress() {
	// If we don't do anything here then the function will switch from PTT to PTS
	fmt.Println("Double press detected - executing action")

	// toggle the saved mode
	if mode == ModePTT {
		setMode(ModePTS)
	} else {
		setMode(ModePTT)
	}
}

func setMode(newMode string) {
	mode = newMode
	updateSketchyBar()
}

func setIsHeld(newIsHeld bool) {
	isHeld = newIsHeld
	updateSketchyBar()
}

func setInitialMode() {
	status, err := getMicrophoneStatus()
	if err != nil {
		log.Fatalf("failed to get microphone status: %v", err)
	}
	if status == "muted" {
		setMode(ModePTT)
	} else {
		setMode(ModePTS)
	}
}

func isMuted() bool {
	status, err := getMicrophoneStatus()
	if err != nil {
		log.Fatalf("failed to get microphone status: %v", err)
		return false
	}

	if status == "muted" {
		return true
	}
	return false
}

func expandTilde(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}

func updateSketchyBar() {
	triggerSketchyBarEvent("microphone_status_change", fmt.Sprintf("MUTED=%t", isMuted()))
}

func triggerSketchyBarEvent(eventName string, arg string) error {
	cmd := exec.Command("sketchybar", "--trigger", eventName, arg)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
