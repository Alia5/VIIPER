using Viiper.Client;
using Viiper.Client.Devices.Dualshock4;
using Viiper.Client.Types;

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
        try
        {
            var r = await client.BusCreateAsync(null);
            busId = r.BusID;
            createdBus = true;
            Console.WriteLine($"Created bus {busId}");
        }
        catch (Exception ex)
        {
            Console.WriteLine($"BusCreate failed: {ex}");
            return;
        }
    }
    else { busId = list.Buses.Min(); Console.WriteLine($"Using existing bus {busId}"); }
}

// Add device and connect
Device resp; ViiperDevice device;
try
{
    resp = await client.BusDeviceAddAsync(busId, new DeviceCreateRequest { Type = "dualshock4" });
    device = await client.ConnectDeviceAsync(resp.BusID, resp.DevID);
    Console.WriteLine($"Created and connected to device {resp.DevID} on bus {resp.BusID}");
}
catch (Exception ex)
{
    if (createdBus) { try { await client.BusRemoveAsync(busId); } catch { } }
    Console.WriteLine($"AddDevice/connect error: {ex}");
    return;
}

AppDomain.CurrentDomain.ProcessExit += async (_, __) => await Cleanup();
Console.CancelKeyPress += async (_, e) => { e.Cancel = true; await Cleanup(); Environment.Exit(0); };

async Task Cleanup()
{
    try { await client.BusDeviceRemoveAsync(resp.BusID, resp.DevID); Console.WriteLine($"Removed device {resp.DevID}"); } catch { }
    if (createdBus) { try { await client.BusRemoveAsync(busId); Console.WriteLine($"Removed bus {busId}"); } catch { } }
}

// Read rumble/LED output using callback with stream
device.OnOutput = async stream =>
{
    var buf = new byte[Dualshock4.OutputSize];
    await stream.ReadAsync(buf, 0, buf.Length);
    byte updateFlags = buf[0];
    byte rumbleSmall = buf[1];
    byte rumbleLarge = buf[2];
    byte ledRed      = buf[3];
    byte ledGreen    = buf[4];
    byte ledBlue     = buf[5];
    byte flashOn     = buf[6];
    byte flashOff    = buf[7];
    Console.WriteLine($"← Output: Flags=0x{updateFlags:X2}, RumbleSmall={rumbleSmall}, RumbleLarge={rumbleLarge}, LED=#{ledRed:X2}{ledGreen:X2}{ledBlue:X2}, Flash={flashOn}/{flashOff}");
};

// Handle disconnect
device.OnDisconnect = () => Console.WriteLine("!!! Server disconnected");

// Send inputs at ~60 FPS
var sw = new PeriodicTimer(TimeSpan.FromMilliseconds(16));
ulong frame = 0;
while (await sw.WaitForNextTickAsync())
{
    frame++;
    ushort buttons = (ushort)(((frame / 60) % 4) switch
    {
        0 => (ulong)Button.Cross,
        1 => (ulong)Button.Circle,
        2 => (ulong)Button.Square,
        _ => (ulong)Button.Triangle,
    });
    var state = new Dualshock4Input
    {
        Buttons     = buttons,
        Dpad        = (byte)D.PadUSBNeutral,
        Sticklx     = (sbyte)((frame * 2) % 128),
        Stickly     = (sbyte)((frame * 3) % 128),
        Stickrx     = 0,
        Stickry     = 0,
        Triggerl2   = (byte)((frame * 2) % 256),
        Triggerr2   = (byte)((frame * 3) % 256),
        Touch1x     = 0,
        Touch1y     = 0,
        Touch1active = 0,
        Touch2x     = 0,
        Touch2y     = 0,
        Touch2active = 0,
        Gyrox       = 0,
        Gyroy       = 0,
        Gyroz       = 0,
        Accelx      = 0,
        Accely      = 0,
        Accelz      = (short)Default.AccelZRaw,
    };
    await device.SendAsync(state);
    if (frame % 60 == 0)
        Console.WriteLine($"→ Sent input (frame {frame}): buttons=0x{buttons:X4}, L2={state.Triggerl2}, R2={state.Triggerr2}");
}
