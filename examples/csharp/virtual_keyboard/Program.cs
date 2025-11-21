using System.Buffers;
using System.Text;
using Viiper.Client;
using Viiper.Client.Devices.Keyboard;
using Viiper.Client.Types;

// Usage: dotnet run -- <host> (e.g., localhost)
if (args.Length < 1)
{
    Console.WriteLine("Usage: dotnet run -- <host> [port]");
    Console.WriteLine("Example: dotnet run -- localhost 3242");
    return;
}

var host = args[0];
var port = args.Length > 1 && int.TryParse(args[1], out var p) ? p : 3242;
var client = new ViiperClient(host, port);

// Find or create a bus
uint busId;
bool createdBus = false;
{
    var list = await client.BusListAsync();
    if (list.Buses.Length == 0)
    {
        Exception? lastErr = null;
        busId = 0;
        for (uint i = 1; i <= 100; i++)
        {
            try
            {
                var resp = await client.BusCreateAsync(i);
                busId = resp.BusID;
                createdBus = true;
                break;
            }
            catch (Exception ex)
            {
                lastErr = ex;
            }
        }
        if (busId == 0)
        {
            Console.WriteLine($"BusCreate failed: {lastErr}");
            return;
        }
        Console.WriteLine($"Created bus {busId}");
    }
    else
    {
        busId = list.Buses.Min();
        Console.WriteLine($"Using existing bus {busId}");
    }
}

Device deviceInfo;
ViiperDevice device;
try
{
    // 2) Add device and connect
    deviceInfo = await client.BusDeviceAddAsync(busId, new DeviceCreateRequest { Type = "keyboard" });
    device = await client.ConnectDeviceAsync(deviceInfo.BusID, deviceInfo.DevId);
    Console.WriteLine($"Created and connected to device {deviceInfo.DevId} on bus {deviceInfo.BusID}");
}
catch (Exception ex)
{
    if (createdBus)
    {
        try { await client.BusRemoveAsync(busId); } catch { }
    }
    Console.WriteLine($"AddDevice/connect error: {ex}");
    return;
}

// Cleanup on exit
AppDomain.CurrentDomain.ProcessExit += async (_, __) => await Cleanup();
Console.CancelKeyPress += async (_, e) => { e.Cancel = true; await Cleanup(); Environment.Exit(0); };

async Task Cleanup()
{
    try { await client.BusDeviceRemoveAsync(deviceInfo.BusID, deviceInfo.DevId); Console.WriteLine($"Removed device {deviceInfo.DevId}"); } catch { }
    if (createdBus)
    {
        try { await client.BusRemoveAsync(busId); Console.WriteLine($"Removed bus {busId}"); } catch { }
    }
}

// Subscribe to LED output (1 byte frames)
device.OnOutput += data =>
{
    if (data.Length < 1) return;
    byte leds = data[0];
    Console.WriteLine($"→ LEDs: Num={(leds & (byte)LED.NumLock) != 0} Caps={(leds & (byte)LED.CapsLock) != 0} Scroll={(leds & (byte)LED.ScrollLock) != 0} Compose={(leds & (byte)LED.Compose) != 0} Kana={(leds & (byte)LED.Kana) != 0}");
};

Console.WriteLine("Every 5s: type 'Hello!' + Enter. Press Ctrl+C to stop.");
var timer = new PeriodicTimer(TimeSpan.FromSeconds(5));

while (await timer.WaitForNextTickAsync())
{
    await TypeString(device, "Hello!");
    await PressAndRelease(device, Key.Enter);
    Console.WriteLine("→ Typed: Hello!");
}

static async Task TypeString(ViiperDevice dev, string text)
{
    foreach (var c in text)
    {
        var (mods, key) = MapChar(c);
        if (key != 0)
        {
            // Press
            await Send(dev, mods, new byte[] { (byte)key });
            await Task.Delay(100);
            // Release
            await Send(dev, 0, Array.Empty<byte>());
            await Task.Delay(100);
        }
    }
}

static async Task PressAndRelease(ViiperDevice dev, Key key)
{
    await Task.Delay(100);
    await Send(dev, 0, new byte[] { (byte)key });
    await Task.Delay(100);
    await Send(dev, 0, Array.Empty<byte>());
}

static async Task Send(ViiperDevice dev, byte modifiers, byte[] keys)
{
    var input = new KeyboardInput
    {
        Modifiers = modifiers,
        Count = (byte)keys.Length,
        Keys = keys
    };
    await dev.SendAsync(input);
}

static (byte mods, Key key) MapChar(char ch)
{
    // Use the generated CharToKey map
    if (!CharToKey.TryGetValue((byte)ch, out var key))
    {
        return (0, 0); // Character not found
    }
    
    // Check if shift is needed using the ShiftChars map
    byte mods = ShiftChars.ContainsKey((byte)ch) ? (byte)Mod.LeftShift : (byte)0;
    
    return (mods, key);
}
