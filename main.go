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
)

const (
	doublePressDelay = 300 * time.Millisecond
	stateDir         = "~/.local/state/hush"
	stateFile        = "muted"
)

func main() { mainthread.Init(fn) } // Required for MacOS

func fn() {
	hk := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl}, hotkey.KeyTab)
	err := hk.Register()
	if err != nil {
		log.Fatalf("hotkey: failed to register hotkey: %v", err)
		return
	}
	defer hk.Unregister()

	log.Printf("hotkey: %v is registered\n", hk)

	notifySketchyBar() // Send initial notifications for initial state

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
}

func handleKeyUp() {
	endKeyPress()
}

func startKeyPress() {
	toggleMicrophone()
	notifyMutedState()
}

func endKeyPress() {
	toggleMicrophone()
	notifyMutedState()
}

func handleDoublePress() {
	// If we don't do anything here then the function will switch from PTT to PTS
	fmt.Println("Double press detected - switching modes")
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

func notifyMutedState() {
	updateSketchyBar()
	updateStateFile()
}

func updateStateFile() {
	expandedStateDir, err := expandTilde(stateDir)
	if err != nil {
		log.Fatalf("failed to expand state dir: %v", err)
		return
	}

	err = os.MkdirAll(expandedStateDir, 0755)
	if err != nil {
		log.Fatalf("failed to create state dir: %v", err)
		return
	}

	fullPath := filepath.Join(expandedStateDir, stateFile)

	file, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("failed to open state file: %v", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("%t", isMuted()))
	if err != nil {
		log.Fatalf("failed to write state file: %v", err)
	}
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
		log.Fatalf("failed to trigger sketchybar event: %v", err)
		return err
	}
	log.Printf("Triggered sketchybar event: %s %s", eventName, arg)
	return nil
}
