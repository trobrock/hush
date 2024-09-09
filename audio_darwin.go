package main

/*
#cgo LDFLAGS: -framework CoreAudio

#include <CoreAudio/CoreAudio.h>
#include <stdio.h>

int toggleMute(AudioDeviceID deviceID) {
    UInt32 muted;
    UInt32 size = sizeof(muted);
    AudioObjectPropertyAddress propertyAddress = {
        kAudioDevicePropertyMute,
        kAudioDevicePropertyScopeInput,
        kAudioObjectPropertyElementMain
    };

    OSStatus status = AudioObjectGetPropertyData(deviceID, &propertyAddress, 0, NULL, &size, &muted);
    if (status != noErr) {
        printf("Error getting mute state: %d\n", status);
        return -1;
    }

    muted = !muted;
    status = AudioObjectSetPropertyData(deviceID, &propertyAddress, 0, NULL, size, &muted);
    if (status != noErr) {
        printf("Error setting mute state: %d\n", status);
        return -1;
    }

    printf("Mute state set to: %d\n", muted);
    return muted;
}

int getMuteState(AudioDeviceID deviceID) {
    UInt32 muted;
    UInt32 size = sizeof(muted);
    AudioObjectPropertyAddress propertyAddress = {
        kAudioDevicePropertyMute,
        kAudioDevicePropertyScopeInput,
        kAudioObjectPropertyElementMain
    };

    OSStatus status = AudioObjectGetPropertyData(deviceID, &propertyAddress, 0, NULL, &size, &muted);
    if (status != noErr) {
        printf("Error getting mute state: %d\n", status);
        return -1;
    }

    printf("Current mute state: %d\n", muted);
    return muted;
}

AudioDeviceID getDefaultInputDevice() {
    AudioDeviceID deviceID = kAudioObjectUnknown;
    UInt32 size = sizeof(deviceID);
    AudioObjectPropertyAddress propertyAddress = {
        kAudioHardwarePropertyDefaultInputDevice,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };

    OSStatus status = AudioObjectGetPropertyData(kAudioObjectSystemObject, &propertyAddress, 0, NULL, &size, &deviceID);
    if (status != noErr) {
        printf("Error getting default input device: %d\n", status);
        return kAudioObjectUnknown;
    }

    printf("Default input device ID: %d\n", deviceID);
    return deviceID;
}
*/
import "C"
import (
    "fmt"
    "errors"
    "time"
)

func toggleMicrophone() error {
    deviceID := C.getDefaultInputDevice()
    if deviceID == C.kAudioObjectUnknown {
        return errors.New("failed to get default input device")
    }

    result := C.toggleMute(deviceID)
    if result == -1 {
        return errors.New("failed to toggle mute state")
    }

    // Verify the mute state after toggling
    time.Sleep(100 * time.Millisecond) // Small delay to allow system to update
    verifyState := C.getMuteState(deviceID)
    
    if result == 0 {
        fmt.Println("Attempted to unmute")
    } else {
        fmt.Println("Attempted to mute")
    }

    if verifyState != result {
        return fmt.Errorf("mute state verification failed. Expected: %d, Got: %d", result, verifyState)
    }

    return nil
}

func getMicrophoneStatus() (string, error) {
    deviceID := C.getDefaultInputDevice()
    if deviceID == C.kAudioObjectUnknown {
        return "", errors.New("failed to get default input device")
    }

    state := C.getMuteState(deviceID)
    if state == -1 {
        return "", errors.New("failed to get mute state")
    }

    if state == 0 {
        return "unmuted", nil
    }
    return "muted", nil
}
