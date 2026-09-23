//go:build windows

package api

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"github.com/Alia5/VIIPER/usbip"
	"golang.org/x/sys/windows"
)

var (
	setupapi                             = windows.NewLazySystemDLL("setupapi.dll")
	procSetupDiGetClassDevsW             = setupapi.NewProc("SetupDiGetClassDevsW")
	procSetupDiEnumDeviceInterfaces      = setupapi.NewProc("SetupDiEnumDeviceInterfaces")
	procSetupDiGetDeviceInterfaceDetailW = setupapi.NewProc("SetupDiGetDeviceInterfaceDetailW")
	procSetupDiDestroyDeviceInfoList     = setupapi.NewProc("SetupDiDestroyDeviceInfoList")
)

const (
	DigcfPresent         = 0x00000002
	DigcfDeviceInterface = 0x00000010
)

type SpDeviceInterfaceData struct {
	CbSize             uint32
	InterfaceClassGUID windows.GUID
	Flags              uint32
	Reserved           uintptr
}

type SpDeviceInterfaceDetailData struct {
	CbSize     uint32
	DevicePath [1]uint16
}

// Device GUID from usbip-win2 driver
var deviceGUID = windows.GUID{
	Data1: 0xB4030C06,
	Data2: 0xDC5F,
	Data3: 0x4FCC,
	Data4: [8]byte{0x87, 0xEB, 0xE5, 0x51, 0x5A, 0x09, 0x35, 0xC0},
}

const (
	niMaxHost   = 1025
	niMaxServ   = 32
	niMaxSerial = 16
)

// PLUGIN_HARDWARE structure from usbip-win2
type attachIOCTL_97 struct { // nolint
	Size       uint32
	PortOutput int32
	BusID      [32]byte
	Service    [niMaxServ]byte
	Host       [niMaxHost]byte
}
type attachIOCTL_98Plus struct { // nolint
	Size       uint32
	PortOutput int32
	BusID      [32]byte
	Service    [niMaxServ]byte
	Host       [niMaxHost]byte
	Serial     [niMaxSerial]byte
	WskEvents  bool
	_          [3]byte
}

const (
	fileDeviceUnknown   = 0x00000022
	methodBuffered      = 0
	fileReadData        = 0x0001
	fileWriteData       = 0x0002
	ioctlPluginHardware = (fileDeviceUnknown << 16) | ((fileReadData | fileWriteData) << 14) | (0x800 << 2) | methodBuffered
)

func attachLocalhostClientImpl(ctx context.Context, deviceExportMeta *usbip.ExportMeta, usbipServerPort uint16, useNativeIOCTL bool, logger *slog.Logger) error {
	if useNativeIOCTL {
		err := attachViaIOCTL(ctx, deviceExportMeta, usbipServerPort, logger)
		if err != nil {
			slog.Error("Native IOCTL auto-attach failed, falling back to command execution", "error", err)
			slog.Info("Trying fallback via usbip executable")
		} else {
			return nil
		}
	}
	return attachViaCommand(ctx, deviceExportMeta, usbipServerPort, logger)
}

func attachViaIOCTL(_ context.Context, deviceExportMeta *usbip.ExportMeta, usbipServerPort uint16, logger *slog.Logger) error {
	logger.Info("Auto-attaching localhost client via native IOCTL",
		"busID", deviceExportMeta.BusID,
		"deviceID", deviceExportMeta.DevID)

	if usbipServerPort == 0 {
		return fmt.Errorf("argumentValidation: invalid TCP port number (0)")
	}

	devicePath, err := getDeviceInterfacePath(&deviceGUID)
	if err != nil {
		return fmt.Errorf("discovery: %w", err)
	}

	logger.Debug("Found usbip-win2 device", "path", devicePath)

	var ioctlData attachIOCTL_98Plus
	ioctlData.Size = uint32(unsafe.Sizeof(ioctlData))

	busID := fmt.Sprintf("%d-%d", deviceExportMeta.BusID, deviceExportMeta.DevID)
	if len(busID) >= len(ioctlData.BusID) {
		return fmt.Errorf("argumentValidation: bus ID too long: %s", busID)
	}
	copy(ioctlData.BusID[:], busID)

	service := fmt.Sprintf("%d", usbipServerPort)
	if len(service) >= len(ioctlData.Service) {
		return fmt.Errorf("argumentValidation: service string too long: %s", service)
	}
	copy(ioctlData.Service[:], service)
	copy(ioctlData.Host[:], "localhost")

	devicePathUTF16, err := windows.UTF16PtrFromString(devicePath)
	if err != nil {
		return fmt.Errorf("open: failed to convert device path: %w", err)
	}

	handle, err := windows.CreateFile(
		devicePathUTF16,
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		return fmt.Errorf("open: failed to open usbip-win2 device: %w", err)
	}
	defer windows.CloseHandle(handle) // nolint

	logger.Debug("Opened device handle")

	var bytesReturned uint32
	err = windows.DeviceIoControl(
		handle,
		ioctlPluginHardware,
		(*byte)(unsafe.Pointer(&ioctlData)),
		uint32(unsafe.Sizeof(ioctlData)),
		(*byte)(unsafe.Pointer(&ioctlData)),
		uint32(unsafe.Sizeof(ioctlData)),
		&bytesReturned,
		nil,
	)
	if err != nil {
		if strings.Contains(err.Error(), "supplied user buffer is not valid") {
			logger.Info("User is running older version of USBIP-Win2: Falling back", "error", err)

			ioctlData97 := attachIOCTL_97{}
			ioctlData97.Size = uint32(unsafe.Sizeof(ioctlData97))

			busID := fmt.Sprintf("%d-%d", deviceExportMeta.BusID, deviceExportMeta.DevID)
			if len(busID) >= len(ioctlData97.BusID) {
				return fmt.Errorf("argumentValidation: bus ID too long: %s", busID)
			}
			copy(ioctlData97.BusID[:], busID)

			service := fmt.Sprintf("%d", usbipServerPort)
			if len(service) >= len(ioctlData97.Service) {
				return fmt.Errorf("argumentValidation: service string too long: %s", service)
			}
			copy(ioctlData97.Service[:], service)
			copy(ioctlData97.Host[:], "localhost")
			err = windows.DeviceIoControl(
				handle,
				ioctlPluginHardware,
				(*byte)(unsafe.Pointer(&ioctlData97)),
				uint32(unsafe.Sizeof(ioctlData97)),
				(*byte)(unsafe.Pointer(&ioctlData97)),
				uint32(unsafe.Sizeof(ioctlData97)),
				&bytesReturned,
				nil,
			)
			if err != nil {
				return fmt.Errorf("IOControl: DeviceIoControl failed with attachIOCTL_97: %w", err)
			}
			ioctlData.PortOutput = ioctlData97.PortOutput
		} else {
			return fmt.Errorf("IOControl: DeviceIoControl failed: %w", err)
		}
	}

	logger.Debug("IOCTL completed", "bytesReturned", bytesReturned, "portOutput", ioctlData.PortOutput)

	if ioctlData.PortOutput <= 0 {
		return fmt.Errorf("ResponseValidation: invalid USB port returned: %d", ioctlData.PortOutput)
	}

	logger.Info("Successfully attached device via IOCTL",
		"busID", deviceExportMeta.BusID,
		"deviceID", deviceExportMeta.DevID,
		"usbPort", ioctlData.PortOutput)

	return nil
}

func attachViaCommand(ctx context.Context, deviceExportMeta *usbip.ExportMeta, usbipServerPort uint16, logger *slog.Logger) error {
	logger.Info("Auto-attaching localhost client", "busID", deviceExportMeta.BusID, "deviceID", deviceExportMeta.DevID)

	cmd := exec.CommandContext(
		ctx,
		"usbip",
		"--tcp-port",
		strconv.FormatUint(uint64(usbipServerPort), 10),
		"attach",
		"-r", "localhost",
		"-b", fmt.Sprintf("%d-%d", deviceExportMeta.BusID, deviceExportMeta.DevID),
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("Failed to attach device",
			"error", err,
			"port", usbipServerPort,
			"output", string(output))
		return err
	}
	logger.Debug("usbip attach output", "output", string(output))

	return nil
}

func getDeviceInterfacePath(guid *windows.GUID) (string, error) {
	r0, _, e1 := syscall.SyscallN(procSetupDiGetClassDevsW.Addr(),
		uintptr(unsafe.Pointer(guid)),
		0,
		0,
		uintptr(DigcfPresent|DigcfDeviceInterface))

	devInfo := windows.Handle(r0)
	if devInfo == windows.InvalidHandle {
		if e1 != 0 {
			return "", fmt.Errorf("discovery: SetupDiGetClassDevsW failed: %w", e1)
		}
		return "", fmt.Errorf("discovery: SetupDiGetClassDevsW failed with invalid handle")
	}
	defer func() {
		_, _, err := syscall.SyscallN(procSetupDiDestroyDeviceInfoList.Addr(), uintptr(devInfo))
		if err != 0 {
			slog.Error("SetupDiDestroyDeviceInfoList failed", "error", err)
		}
	}()

	var interfaceData SpDeviceInterfaceData
	interfaceData.CbSize = uint32(unsafe.Sizeof(interfaceData))

	r1, _, e2 := syscall.SyscallN(procSetupDiEnumDeviceInterfaces.Addr(),
		uintptr(devInfo),
		0,
		uintptr(unsafe.Pointer(guid)),
		0,
		uintptr(unsafe.Pointer(&interfaceData)))

	if r1 == 0 {
		if e2 != 0 {
			return "", fmt.Errorf("discovery: usbip-win2 driver not found: %w", e2)
		}
		return "", fmt.Errorf("discovery: usbip-win2 driver not found")
	}

	var requiredSize uint32
	r2, _, err := syscall.SyscallN(procSetupDiGetDeviceInterfaceDetailW.Addr(),
		uintptr(devInfo),
		uintptr(unsafe.Pointer(&interfaceData)),
		0,
		0,
		uintptr(unsafe.Pointer(&requiredSize)),
		0)
	if r2 == 0 && err != windows.ERROR_INSUFFICIENT_BUFFER {
		return "", fmt.Errorf("discovery: SetupDiGetDeviceInterfaceDetailW (size query) failed: %w", err)
	}
	if requiredSize == 0 {
		return "", fmt.Errorf("discovery: SetupDiGetDeviceInterfaceDetailW (size query) returned invalid required size")
	}

	detailData := make([]byte, requiredSize)
	detailHeader := (*SpDeviceInterfaceDetailData)(unsafe.Pointer(&detailData[0]))
	detailHeader.CbSize = uint32(unsafe.Sizeof(SpDeviceInterfaceDetailData{}))

	r3, _, e3 := syscall.SyscallN(procSetupDiGetDeviceInterfaceDetailW.Addr(),
		uintptr(devInfo),
		uintptr(unsafe.Pointer(&interfaceData)),
		uintptr(unsafe.Pointer(detailHeader)),
		uintptr(requiredSize),
		0,
		0)

	if r3 == 0 {
		if e3 != 0 {
			return "", fmt.Errorf("discovery: SetupDiGetDeviceInterfaceDetailW failed: %w", e3)
		}
		return "", fmt.Errorf("discovery: SetupDiGetDeviceInterfaceDetailW failed")
	}

	path := windows.UTF16PtrToString(&detailHeader.DevicePath[0])
	return path, nil
}

func CheckAutoAttachPrerequisites(useNativeIOCTL bool, logger *slog.Logger) bool {
	if useNativeIOCTL {
		_, err := getDeviceInterfacePath(&deviceGUID)
		if err != nil {
			logger.Warn("Native IOCTL auto-attach prerequisites not met", "error", err)
			logger.Warn("Native IOCTL auto-attach is unavailable until discovery succeeds")
			logger.Info("If usbip-win2 is not installed, download and install:")
			logger.Info("  https://github.com/vadimgrn/usbip-win2")
			logger.Info("  https://github.com/OSSign/vadimgrn--usbip-win2")
			return false
		}
		logger.Debug("usbip-win2 driver found")
		return true
	}

	if _, err := exec.LookPath("usbip.exe"); err != nil {
		logger.Warn("USB/IP tool 'usbip.exe' not found in PATH")
		logger.Warn("Auto-attach requires usbip-win2")
		logger.Info("Download and install usbip-win2:")
		logger.Info("  https://github.com/vadimgrn/usbip-win2")
		return false
	}

	logger.Debug("usbip.exe tool found in PATH")
	return true
}
