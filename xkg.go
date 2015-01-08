package xkg

// #cgo pkg-config: x11 xext xi
//
// #include <stdio.h>
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
	"os"
	"reflect"
	"unsafe"
)

// Constants
const KeyClass = 0
const InvalidType C.int = -1

// Global Variables
var cKeyPressType C.int = InvalidType
var cKeyReleaseType C.int = InvalidType

// StartXGrabber starts the X Keyboard Grabber
//  keys - channel to send keycodes
func StartXGrabber(keys chan int) {
	var cDisplay *C.Display
	var cDevice *C.XDeviceInfo
	var cNumEvents C.int

	// Open X Display
	cDisplay = C.XOpenDisplay(nil)

	// Get Keyboard Device
	cDevice = findKeyboardDevice(cDisplay)

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

	// Grab X keyboard events
	grabXEvents(cDisplay, keys)
}

// grabXEvents captures the keyboard events of Display X
func grabXEvents(cDisplay *C.Display, keys chan int) {
	var cEvent C.XEvent
	var cKey *C.XDeviceKeyEvent

	for {
		C.XNextEvent(cDisplay, &cEvent)
		keyPressed := (C.isType(&cEvent, cKeyPressType) != 0)

		if keyPressed {
			// Convert C.XEvent into *C.XDeviceKeyEvent
			cKey = ((*C.XDeviceKeyEvent)(unsafe.Pointer(&cEvent)))

			// Send Keycode to channel
			keys <- int(cKey.keycode)
		} else {
			// unknown event
		}
	}
}

// findKeyboardDevice return the default keyboard device (AT Translated Set 2 keyboard)
func findKeyboardDevice(cDisplay *C.Display) *C.XDeviceInfo {
	var cDevices *C.XDeviceInfo
	var cFound *C.XDeviceInfo
	var cNumDevices C.int

	cDevices = C.XListInputDevices(cDisplay, &cNumDevices)
	devices := toXDeviceInfoSlice(cDevices, cNumDevices)

	for _, device := range devices {
		if C.strcmp(device.name, C.CString("AT Translated Set 2 keyboard")) == 0 {
			cFound = &device
			break
		}
	}

	return cFound
}

// registerEvents register KeyPress and KeyRelease event classes into keyboard device
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
		classes = toXInputClassInfoSlice(cDevice.classes, cDevice.num_classes)

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

// toXDeviceInfoSlice converts *C.XDeviceInfo into []C.XDeviceInfo
func toXDeviceInfoSlice(array *C.XDeviceInfo, length C.int) []C.XDeviceInfo {
	// Convert C arrays into Go slices
	// Ref: https://code.google.com/p/go-wiki/wiki/cgo
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(array)),
		Len:  int(length),
		Cap:  int(length),
	}

	return *(*[]C.XDeviceInfo)(unsafe.Pointer(&hdr))
}

// toXInputClassInfoSlice converts *C.XInputClassInfo into []C.XInputClassInfo
func toXInputClassInfoSlice(array *C.XInputClassInfo, length C.int) []C.XInputClassInfo {
	// Convert C arrays into Go slices
	// Ref: https://code.google.com/p/go-wiki/wiki/cgo
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(array)),
		Len:  int(length),
		Cap:  int(length),
	}

	return *(*[]C.XInputClassInfo)(unsafe.Pointer(&hdr))
}
