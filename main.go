package main

// #cgo pkg-config: x11 xext xi
//
// #include <stdio.h>
// #include <stdlib.h>
// #include <ctype.h>
// #include <string.h>
// #include <X11/Xlib.h>
// #include <X11/extensions/XInput.h>
// #include <X11/extensions/XInput2.h>
// #include <X11/Xutil.h>
//
// int defaultScreen(Display *display) {
//     return DefaultScreen(display);
// }
//
// Window rootWindow(Display *display, int screen) {
//     return RootWindow(display, screen);
// }
//
// void deviceKeyPress(XDevice *device, int *type, XEventClass *event) {
//     DeviceKeyPress(device, *type, *event);
// }
//
// void deviceKeyRelease(XDevice *device, int *type, XEventClass *event) {
//     DeviceKeyRelease(device, *type, *event);
// }
//
// int isType(XEvent *event, int type) {
//     return	 event->type == type;
// }
import "C"

import (
	"fmt"
	"os"
	"reflect"
	"unsafe"
)

// Constants
const KeyClass = 0
const _IOLBF = 1
const InvalidType C.int = -1

// Global Variables
var cKeyPressType C.int = InvalidType
var cKeyReleaseType C.int = InvalidType

func getKeyboardId(cDisplay *C.Display) C.XID {
	var cDevices *C.XDeviceInfo
	var cNumDevices C.int
	var cId C.XID = 0

	cDevices = C.XListInputDevices(cDisplay, &cNumDevices)
	devices := XDeviceInfoToSlice(cDevices, cNumDevices)

	for _, device := range devices {
		if C.strcmp(device.name, C.CString("AT Translated Set 2 keyboard")) == 0 {
			cId = device.id
			break
		}
	}

	return cId
}

func findDevice(cDisplay *C.Display, cId C.XID) *C.XDeviceInfo {
	var cDevices *C.XDeviceInfo
	var cFound *C.XDeviceInfo
	var cNumDevices C.int

	cDevices = C.XListInputDevices(cDisplay, &cNumDevices)
	devices := XDeviceInfoToSlice(cDevices, cNumDevices)

	for _, device := range devices {
		if device.id == cId {
			cFound = &device
			break
		}
	}

	return cFound
}

func registerEvents(cDisplay *C.Display, cDeviceInfo *C.XDeviceInfo) C.int {
	var cNumEvents C.int
	var cEventList [2]C.XEventClass
	var cEventListPtr *C.XEventClass
	var cDevice *C.XDevice
	var cRootWin C.Window
	var cScreen C.int
	var classes []C.XInputClassInfo

	cScreen = C.defaultScreen(cDisplay)
	cRootWin = C.rootWindow(cDisplay, cScreen)
	cDevice = C.XOpenDevice(cDisplay, cDeviceInfo.id)

	if cDevice == nil {
		// unable to open device
		return cNumEvents
	}

	if cDevice.num_classes > 0 {
		classes = XInputClassInfoToSlice(cDevice.classes, cDevice.num_classes)

		for _, class := range classes {
			switch class.input_class {
			case KeyClass:
				C.deviceKeyPress(cDevice, &cKeyPressType, &cEventList[cNumEvents])
				cNumEvents += 1
				C.deviceKeyRelease(cDevice, &cKeyReleaseType, &cEventList[cNumEvents])
				cNumEvents += 1
			default:
				// unknown class
			}
		}

		// Convert [2]C.XEventClass into *C.XEventClass
		cEventListPtr = ((*C.XEventClass)(unsafe.Pointer(&cEventList)))

		if C.XSelectExtensionEvent(cDisplay, cRootWin, cEventListPtr, cNumEvents) != 0 {
			// error selecting extended events
			return 0
		}
	}

	return cNumEvents
}

func keyEvents(cDisplay *C.Display) {
	var cEvent C.XEvent
	var cKey *C.XDeviceKeyEvent

	C.setvbuf(C.stdout, nil, _IOLBF, 0)

	for {
		C.XNextEvent(cDisplay, &cEvent)

		keyPressed := (C.isType(&cEvent, cKeyPressType) != 0)
		keyReleased := (C.isType(&cEvent, cKeyReleaseType) != 0)

		if keyPressed || keyReleased {
			// Convert C.XEvent into *C.XDeviceKeyEvent
			cKey = ((*C.XDeviceKeyEvent)(unsafe.Pointer(&cEvent)))

			fmt.Print("press=", keyPressed, "  release=", keyReleased, "  keyCode=", cKey.keycode)

			if keyStr, exist := KeyMap[int(cKey.keycode)]; exist {
				fmt.Println("  keyStr=", keyStr)
			}

		} else {
			// unknown event
		}
	}
}

func main() {
	var cDisplay *C.Display
	var cId C.XID
	var cDevice *C.XDeviceInfo
	var cNumEvents C.int

	// Open X Display
	cDisplay = C.XOpenDisplay(nil)

	// Get Keyboard Id
	cId = getKeyboardId(cDisplay)

	// Get Keyboard Device
	cDevice = findDevice(cDisplay, cId)

	// Unable to find device
	if cDevice == nil {
		os.Exit(1)
	}

	// Register Events
	cNumEvents = registerEvents(cDisplay, cDevice)

	// No events registered
	if cNumEvents == 0 {
		os.Exit(1)
	}

	keyEvents(cDisplay)
}

// Convert C arrays into Go slices
// Ref: https://code.google.com/p/go-wiki/wiki/cgo
func XDeviceInfoToSlice(array *C.XDeviceInfo, length C.int) []C.XDeviceInfo {
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(array)),
		Len:  int(length),
		Cap:  int(length),
	}

	return *(*[]C.XDeviceInfo)(unsafe.Pointer(&hdr))
}

// Convert C arrays into Go slices
// Ref: https://code.google.com/p/go-wiki/wiki/cgo
func XInputClassInfoToSlice(array *C.XInputClassInfo, length C.int) []C.XInputClassInfo {
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(array)),
		Len:  int(length),
		Cap:  int(length),
	}

	return *(*[]C.XInputClassInfo)(unsafe.Pointer(&hdr))
}
