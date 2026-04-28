using System.Runtime.InteropServices;

class Program
{
    static void RumbleCallback(nuint handle, byte leftMotor, byte rightMotor)
        => Console.WriteLine($"<- Rumble: Left={leftMotor}, Right={rightMotor}");

    static void LogCallback(VIIPERLogLevel level, string message)
    {
        string levelStr = level switch
        {
            VIIPERLogLevel.Debug => "DEBUG",
            VIIPERLogLevel.Info  => "INFO",
            VIIPERLogLevel.Warn  => "WARN",
            VIIPERLogLevel.Error => "ERROR",
            _                    => "UNKNOWN",
        };
        Console.WriteLine($"libVIIPER [{levelStr}] {message}");
    }

    static int Main()
    {
        Console.WriteLine("Hello, libVIIPER!");

        bool run = true;

        USBServerConfig conf = new() { 
            addr = "localhost:3245" 
        };

        VIIPERLogCallbackDelegate logCb = LogCallback;
        bool success = LibVIIPER.NewUSBServer(ref conf, out nuint serverHandle, logCb);
        if (!success) {
            Console.WriteLine("Failed to create USB server.");
            return 1;
        }
        Console.WriteLine($"Created USB server with handle: {serverHandle}");

        uint busID = 0;
        success = LibVIIPER.CreateUSBBus(serverHandle, ref busID);
        if (!success) {
            Console.WriteLine("Failed to create USB bus.");
            return 1;
        }
        Console.WriteLine($"Created USB bus with ID: {busID}");

        success = LibVIIPER.CreateXbox360Device(serverHandle, out nuint deviceHandle, busID, true, 0, 0, 0);
        if (!success) {
            Console.WriteLine("Failed to create Xbox 360 device.");
            return 1;
        }

        Console.CancelKeyPress += (_, e) => {
            e.Cancel = true;
            run = false;
        };
        AppDomain.CurrentDomain.ProcessExit += (_, _) => run = false;

        Xbox360RumbleCallbackDelegate rumbleCb = RumbleCallback;
        LibVIIPER.SetXbox360RumbleCallback(deviceHandle, rumbleCb);

        Thread.Sleep(5000);

        Xbox360DeviceState deviceState = new();
        ulong frame = 0;

        while (run) {
            frame++;
            deviceState.Buttons = (uint)((frame / 60) % 4) switch
            {
                0 => Xbox360Buttons.A,
                1 => Xbox360Buttons.B,
                2 => Xbox360Buttons.X,
                _ => Xbox360Buttons.Y,
            };
            deviceState.LT = (byte)((frame * 2) % 256);
            deviceState.RT = (byte)((frame * 3) % 256);
            deviceState.LX = (short)(20000.0 * Math.Sin(frame * 0.02));
            deviceState.LY = (short)(20000.0 * Math.Cos(frame * 0.02));

            LibVIIPER.SetXbox360DeviceState(deviceHandle, deviceState);

            if (frame % 60 == 0) {
                Console.WriteLine($"-> Sent input (frame {frame}): buttons=0x{deviceState.Buttons:X4}, LT={deviceState.LT}, RT={deviceState.RT}");
            }

            Thread.Sleep(16);
        }

        Console.WriteLine("\nShutting down...");
        LibVIIPER.CloseUSBServer(serverHandle);
        return 0;
    }
}

enum VIIPERLogLevel
{
    Debug = -4,
    Info  = 0,
    Warn  = 4,
    Error = 8,
}

static class Xbox360Buttons
{
    public const uint DPadUp    = 0x0001;
    public const uint DPadDown  = 0x0002;
    public const uint DPadLeft  = 0x0004;
    public const uint DPadRight = 0x0008;
    public const uint Start     = 0x0010;
    public const uint Back      = 0x0020;
    public const uint LThumb    = 0x0040;
    public const uint RThumb    = 0x0080;
    public const uint LShoulder = 0x0100;
    public const uint RShoulder = 0x0200;
    public const uint Guide     = 0x0400;
    public const uint A         = 0x1000;
    public const uint B         = 0x2000;
    public const uint X         = 0x4000;
    public const uint Y         = 0x8000;
}

[StructLayout(LayoutKind.Sequential)]
struct USBServerConfig
{
    [MarshalAs(UnmanagedType.LPStr)]
    public string? addr;
    public ulong connection_timeout_ms;
    public ulong device_handler_connect_timeout_ms;
    public uint  write_batch_flush_interval_ms;
}

[StructLayout(LayoutKind.Sequential)]
struct Xbox360DeviceState
{
    public uint  Buttons;
    public byte  LT;
    public byte  RT;
    public short LX;
    public short LY;
    public short RX;
    public short RY;
    public byte  Reserved0, Reserved1, Reserved2, Reserved3, Reserved4, Reserved5;
}

[UnmanagedFunctionPointer(CallingConvention.Cdecl)]
delegate void Xbox360RumbleCallbackDelegate(nuint handle, byte leftMotor, byte rightMotor);

[UnmanagedFunctionPointer(CallingConvention.Cdecl)]
delegate void VIIPERLogCallbackDelegate(VIIPERLogLevel level, [MarshalAs(UnmanagedType.LPStr)] string message);

static class LibVIIPER
{
    const string Lib = "libVIIPER";

    [DllImport(Lib, CallingConvention = CallingConvention.Cdecl)]
    [return: MarshalAs(UnmanagedType.I1)]
    public static extern bool NewUSBServer([In] ref USBServerConfig config, out nuint outHandle, VIIPERLogCallbackDelegate? logCallback);

    [DllImport(Lib, CallingConvention = CallingConvention.Cdecl)]
    [return: MarshalAs(UnmanagedType.I1)]
    public static extern bool CloseUSBServer(nuint handle);

    [DllImport(Lib, CallingConvention = CallingConvention.Cdecl)]
    [return: MarshalAs(UnmanagedType.I1)]
    public static extern bool CreateUSBBus(nuint handle, ref uint busID);

    [DllImport(Lib, CallingConvention = CallingConvention.Cdecl)]
    [return: MarshalAs(UnmanagedType.I1)]
    public static extern bool CreateXbox360Device(nuint serverHandle, out nuint outDeviceHandle, uint busID, [MarshalAs(UnmanagedType.I1)] bool autoAttachLocalhost, ushort idVendor, ushort idProduct, byte xinputSubType);

    [DllImport(Lib, CallingConvention = CallingConvention.Cdecl)]
    [return: MarshalAs(UnmanagedType.I1)]
    public static extern bool SetXbox360DeviceState(nuint deviceHandle, Xbox360DeviceState state);

    [DllImport(Lib, CallingConvention = CallingConvention.Cdecl)]
    [return: MarshalAs(UnmanagedType.I1)]
    public static extern bool SetXbox360RumbleCallback(nuint deviceHandle, Xbox360RumbleCallbackDelegate? callback);
}
