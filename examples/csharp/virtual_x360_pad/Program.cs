using Viiper.Client;
using Viiper.Client.Devices.Xbox360;
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
    resp = await client.BusDeviceAddAsync(busId, new DeviceCreateRequest { Type = "xbox360" });
    device = await client.ConnectDeviceAsync(resp.BusID, resp.DevId);
    Console.WriteLine($"Created and connected to device {resp.DevId} on bus {resp.BusID}");
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
    try { await client.BusDeviceRemoveAsync(resp.BusID, resp.DevId); Console.WriteLine($"Removed device {resp.DevId}"); } catch { }
    if (createdBus) { try { await client.BusRemoveAsync(busId); Console.WriteLine($"Removed bus {busId}"); } catch { } }
}

// Read rumble output (2 bytes) and log
device.OnOutput += data =>
{
    if (data.Length < 2) return;
    byte left = data[0]; byte right = data[1];
    Console.WriteLine($"← Rumble: Left={left}, Right={right}");
};

// Send inputs at ~60 FPS
var sw = new PeriodicTimer(TimeSpan.FromMilliseconds(16));
ulong frame = 0;
while (await sw.WaitForNextTickAsync())
{
    frame++;
    uint buttons = (uint)(((frame / 60) % 4) switch { 0 => (ulong)Button.A, 1 => (ulong)Button.B, 2 => (ulong)Button.X, _ => (ulong)Button.Y });
    var state = new Xbox360Input
    {
        Buttons = buttons,
        Lt = (byte)((frame * 2) % 256),
        Rt = (byte)((frame * 3) % 256),
        Lx = (short)20000,
        Ly = (short)20000,
        Rx = 0,
        Ry = 0,
    };
    await device.SendAsync(state);
    if (frame % 60 == 0)
        Console.WriteLine($"→ Sent input (frame {frame}): buttons=0x{buttons:X4}, LT={state.Lt}, RT={state.Rt}");
}
