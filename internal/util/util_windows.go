//go:build windows

package util

import (
	"log/slog"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	kernel32             = windows.NewLazySystemDLL("kernel32.dll")
	user32               = windows.NewLazySystemDLL("user32.dll")
	procGetConsoleWindow = kernel32.NewProc("GetConsoleWindow")
	procGetConsoleProcs  = kernel32.NewProc("GetConsoleProcessList")
	procShowWindow       = user32.NewProc("ShowWindow")
	procFreeConsole      = kernel32.NewProc("FreeConsole")
)

func IsRunFromGUI() bool {
	hwnd, _, _ := procGetConsoleWindow.Call()
	hasConsole := hwnd != 0

	if !hasConsole {
		return true
	}

	parentPID := getParentProcessID()
	parentAttached := isPIDAttachedToCurrentConsole(parentPID)

	slog.Debug("Console launch detection", "hasConsole", hasConsole, "parentPID", parentPID, "parentAttached", parentAttached)

	return !parentAttached
}

func HideConsoleWindow() {
	hwnd, _, _ := procGetConsoleWindow.Call()
	if hwnd == 0 {
		slog.Debug("HideConsoleWindow: no console window found")
		return
	}

	_, _, _ = procShowWindow.Call(hwnd, windows.SW_HIDE)
	_, _, _ = procFreeConsole.Call()
}

func getParentProcessID() uint32 {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0
	}
	defer windows.CloseHandle(snapshot) // nolint:errcheck

	var pe windows.ProcessEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))

	currentPID := uint32(os.Getpid())

	if err := windows.Process32First(snapshot, &pe); err != nil {
		return 0
	}

	for {
		if pe.ProcessID == currentPID {
			return pe.ParentProcessID
		}
		if err := windows.Process32Next(snapshot, &pe); err != nil {
			return 0
		}
	}
}

func isPIDAttachedToCurrentConsole(pid uint32) bool {
	if pid == 0 {
		return false
	}

	buf := make([]uint32, 8)

	for {
		count, _, _ := procGetConsoleProcs.Call(
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(len(buf)),
		)

		if count == 0 {
			return false
		}

		needed := int(count)
		if needed > len(buf) {
			buf = make([]uint32, needed)
			continue
		}

		for _, consolePID := range buf[:needed] {
			if consolePID == pid {
				return true
			}
		}

		return false
	}
}
