//go:build windows

package api

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"
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
	DIGCF_PRESENT         = 0x00000002
	DIGCF_DEVICEINTERFACE = 0x00000010
)

type SP_DEVICE_INTERFACE_DATA struct {
	CbSize             uint32
	InterfaceClassGuid windows.GUID
	Flags              uint32
	Reserved           uintptr
}

type SP_DEVICE_INTERFACE_DETAIL_DATA struct {
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
	niMaxHost = 1025
	niMaxServ = 32
)

// PLUGIN_HARDWARE structure from usbip-win2
type attachIOCTL struct {
	Size       uint32
	PortOutput int32
	BusID      [32]byte
	Service    [niMaxServ]byte
	Host       [niMaxHost]byte
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
		return attachViaIOCTL(ctx, deviceExportMeta, usbipServerPort, logger)
	}
	return attachViaCommand(ctx, deviceExportMeta, usbipServerPort, logger)
}

func attachViaIOCTL(ctx context.Context, deviceExportMeta *usbip.ExportMeta, usbipServerPort uint16, logger *slog.Logger) error {
	logger.Info("Auto-attaching localhost client via native IOCTL",
		"busID", deviceExportMeta.BusId,
		"deviceID", deviceExportMeta.DevId)

	if usbipServerPort == 0 {
		return fmt.Errorf("ArgumentValidation: invalid TCP port number (0)")
	}

	devicePath, err := getDeviceInterfacePath(&deviceGUID)
	if err != nil {
		return fmt.Errorf("Discovery: %w", err)
	}

	logger.Debug("Found usbip-win2 device", "path", devicePath)

	var ioctlData attachIOCTL
	ioctlData.Size = uint32(unsafe.Sizeof(ioctlData))

	busID := fmt.Sprintf("%d-%d", deviceExportMeta.BusId, deviceExportMeta.DevId)
	if len(busID) >= len(ioctlData.BusID) {
		return fmt.Errorf("ArgumentValidation: bus ID too long: %s", busID)
	}
	copy(ioctlData.BusID[:], busID)

	service := fmt.Sprintf("%d", usbipServerPort)
	if len(service) >= len(ioctlData.Service) {
		return fmt.Errorf("ArgumentValidation: service string too long: %s", service)
	}
	copy(ioctlData.Service[:], service)
	copy(ioctlData.Host[:], "localhost")

	devicePathUTF16, err := windows.UTF16PtrFromString(devicePath)
	if err != nil {
		return fmt.Errorf("Open: failed to convert device path: %w", err)
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
		return fmt.Errorf("Open: failed to open usbip-win2 device: %w", err)
	}
	defer windows.CloseHandle(handle)

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
		return fmt.Errorf("IOControl: DeviceIoControl failed: %w", err)
	}

	logger.Debug("IOCTL completed", "bytesReturned", bytesReturned, "portOutput", ioctlData.PortOutput)

	if ioctlData.PortOutput <= 0 {
		return fmt.Errorf("ResponseValidation: invalid USB port returned: %d", ioctlData.PortOutput)
	}

	logger.Info("Successfully attached device via IOCTL",
		"busID", deviceExportMeta.BusId,
		"deviceID", deviceExportMeta.DevId,
		"usbPort", ioctlData.PortOutput)

	return nil
}

func attachViaCommand(ctx context.Context, deviceExportMeta *usbip.ExportMeta, usbipServerPort uint16, logger *slog.Logger) error {
	logger.Info("Auto-attaching localhost client", "busID", deviceExportMeta.BusId, "deviceID", deviceExportMeta.DevId)

	cmd := exec.CommandContext(
		ctx,
		"usbip",
		"--tcp-port",
		strconv.FormatUint(uint64(usbipServerPort), 10),
		"attach",
		"-r", "localhost",
		"-b", fmt.Sprintf("%d-%d", deviceExportMeta.BusId, deviceExportMeta.DevId),
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
		uintptr(DIGCF_PRESENT|DIGCF_DEVICEINTERFACE))

	devInfo := windows.Handle(r0)
	if devInfo == windows.InvalidHandle {
		if e1 != 0 {
			return "", fmt.Errorf("Discovery: SetupDiGetClassDevsW failed: %w", e1)
		}
		return "", fmt.Errorf("Discovery: SetupDiGetClassDevsW failed with invalid handle")
	}
	defer func() {
		syscall.SyscallN(procSetupDiDestroyDeviceInfoList.Addr(), uintptr(devInfo))
	}()

	var interfaceData SP_DEVICE_INTERFACE_DATA
	interfaceData.CbSize = uint32(unsafe.Sizeof(interfaceData))

	r1, _, e2 := syscall.SyscallN(procSetupDiEnumDeviceInterfaces.Addr(),
		uintptr(devInfo),
		0,
		uintptr(unsafe.Pointer(guid)),
		0,
		uintptr(unsafe.Pointer(&interfaceData)))

	if r1 == 0 {
		if e2 != 0 {
			return "", fmt.Errorf("Discovery: usbip-win2 driver not found: %w", e2)
		}
		return "", fmt.Errorf("Discovery: usbip-win2 driver not found")
	}

	var requiredSize uint32
	syscall.SyscallN(procSetupDiGetDeviceInterfaceDetailW.Addr(),
		uintptr(devInfo),
		uintptr(unsafe.Pointer(&interfaceData)),
		0,
		0,
		uintptr(unsafe.Pointer(&requiredSize)),
		0)

	detailData := make([]byte, requiredSize)
	detailHeader := (*SP_DEVICE_INTERFACE_DETAIL_DATA)(unsafe.Pointer(&detailData[0]))
	detailHeader.CbSize = uint32(unsafe.Sizeof(SP_DEVICE_INTERFACE_DETAIL_DATA{}))

	r2, _, e3 := syscall.SyscallN(procSetupDiGetDeviceInterfaceDetailW.Addr(),
		uintptr(devInfo),
		uintptr(unsafe.Pointer(&interfaceData)),
		uintptr(unsafe.Pointer(detailHeader)),
		uintptr(requiredSize),
		0,
		0)

	if r2 == 0 {
		if e3 != 0 {
			return "", fmt.Errorf("Discovery: SetupDiGetDeviceInterfaceDetailW failed: %w", e3)
		}
		return "", fmt.Errorf("Discovery: SetupDiGetDeviceInterfaceDetailW failed")
	}

	path := windows.UTF16PtrToString(&detailHeader.DevicePath[0])
	return path, nil
}

func CheckAutoAttachPrerequisites(useNativeIOCTL bool, logger *slog.Logger) bool {
	if useNativeIOCTL {
		_, err := getDeviceInterfacePath(&deviceGUID)
		if err != nil {
			logger.Warn("usbip-win2 driver not found or not installed")
			logger.Warn("Native IOCTL auto-attach requires the usbip-win2 driver")
			logger.Info("Download and install usbip-win2:")
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
