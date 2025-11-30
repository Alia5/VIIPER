# C# SDK Documentation

The VIIPER C# SDK provides a modern, type-safe .NET client library for interacting with VIIPER servers and controlling virtual devices.

## Overview

The C# SDK features:

- **Async/await support**: Full async API with cancellation token support
- **Type-safe**: Generated classes with enums, structs, and helper maps
- **Event-driven**: `OnOutput` event for device feedback (LEDs, rumble)
- **Modern .NET**: Targets .NET 8.0 with nullable reference types
- **Zero external dependencies**: Uses only built-in .NET libraries

!!! note "License"
    The C# SDK is licensed under the **MIT License**, providing maximum flexibility for integration into your projects.  
    The core VIIPER server remains under its original license.

## Installation

### 1. Using the Published NuGet Package (Recommended)

Install the stable package:

```bash
dotnet add package Viiper.Client
```

Package page: [Viiper.Client on NuGet](https://www.nuget.org/packages/Viiper.Client/)

> Pre-release / snapshot builds are **not** published to NuGet. They are only available as GitHub Release artifacts (e.g. `dev-latest`) or by building from source.

To use a snapshot `.nupkg` from a GitHub Release:

```bash
# 1. Download viiper-csharp-sdk-nupkg-Release.nupkg (or Snapshot) to ./packages
mkdir -p packages
cp /path/to/downloaded/viiper-csharp-sdk-nupkg.nupkg packages/

# 2. Add a temporary local source and install
dotnet nuget add source ./packages --name viiper-local || true
dotnet add package Viiper.Client --source viiper-local
```

Or add directly in your `.csproj` (stable only):

```xml
<ItemGroup>
    <PackageReference Include="Viiper.Client" Version="*" />
</ItemGroup>
```

### 2. Project Reference (For Local Development Against Source)

Use this when modifying the generator or contributing new device types:

```xml
<ItemGroup>
    <ProjectReference Include="..\..\clients\csharp\Viiper.Client\Viiper.Client.csproj" />
</ItemGroup>
```

### 3. Generating from Source (Advanced / Contributors)

Only required when enhancing VIIPER itself:

```bash
go run ./cmd/viiper codegen --lang=csharp
cd clients/csharp
dotnet build -c Release Viiper.Client
```

## Quick Start

```csharp
using Viiper.Client;
using Viiper.Client.Devices.Keyboard;

// Create new Viiper client
var client = new ViiperClient("localhost", 3242);

// Find or create a bus
var buses = await client.BusListAsync();
uint busId;
if (buses.Buses.Length == 0)
{
    var resp = await client.BusCreateAsync(null); // null = auto-assign ID
    // Or specify ID: await client.BusCreateAsync(5);
    busId = resp.BusID;
}
else
{
    busId = buses.Buses[0];
}

// Add device and connect
var deviceReq = new DeviceCreateRequest { Type = "keyboard" };
var deviceResp = await client.BusDeviceAddAsync(busId, deviceReq);
var device = await client.ConnectDeviceAsync(busId, deviceResp.DevId);

Console.WriteLine($"Connected to device {deviceResp.BusID}-{deviceResp.DevId}");

// Send keyboard input
var input = new KeyboardInput
{
    Modifiers = (byte)Mod.LeftShift,
    Count = 1,
    Keys = new[] { (byte)Key.H }
};
await device.SendAsync(input);

// Cleanup
await client.BusDeviceRemoveAsync(busId, deviceResp.DevId);
```

## Device Stream API

### Creating a Device Stream

The simplest way to add a device and connect:

```csharp
var deviceReq = new DeviceCreateRequest { Type = "xbox360" };
var deviceResp = await client.BusDeviceAddAsync(busId, deviceReq);
var device = await client.ConnectDeviceAsync(busId, deviceResp.DevId);
```

With custom VID/PID:

```csharp
var deviceReq = new DeviceCreateRequest { 
    Type = "keyboard", 
    IdVendor = 0x1234, 
    IdProduct = 0x5678 
};
var deviceResp = await client.BusDeviceAddAsync(busId, deviceReq);
var device = await client.ConnectDeviceAsync(busId, deviceResp.DevId);
```

Or connect to an existing device:

```csharp
var device = await client.ConnectDeviceAsync(busId, deviceId);
```

### Sending Input

Device input is sent using generated structs with async methods:

```csharp
using Viiper.Client.Devices.Xbox360;

var input = new Xbox360Input
{
    Buttons = (uint)Button.A,
    LeftTrigger = 255,
    RightTrigger = 0,
    ThumbLX = -32768,  // Left stick left
    ThumbLY = 32767,   // Left stick up
    ThumbRX = 0,
    ThumbRY = 0
};
await device.SendAsync(input);
```

### Receiving Output (Events)

For devices that send feedback (rumble, LEDs), subscribe to the `OnOutput` event:

```csharp
using Viiper.Client.Devices.Keyboard;

device.OnOutput += data =>
{
    if (data.Length < 1) return;
    byte leds = data[0];
    
    Console.WriteLine($"LEDs: " +
        $"Num={(leds & (byte)LED.NumLock) != 0} " +
        $"Caps={(leds & (byte)LED.CapsLock) != 0} " +
        $"Scroll={(leds & (byte)LED.ScrollLock) != 0}");
};
```

For Xbox360 rumble:

```csharp
using Viiper.Client.Devices.Xbox360;

device.OnOutput += data =>
{
    if (data.Length < 2) return;
    byte leftMotor = data[0];
    byte rightMotor = data[1];
    Console.WriteLine($"Rumble: Left={leftMotor} Right={rightMotor}");
};
```

### Closing a Device

```csharp
device.Dispose();
// or
await using var device = await client.ConnectDeviceAsync(busId, deviceId);
```

## Generated Constants and Maps

The C# SDK automatically generates enums and helper maps for each device type.

### Keyboard Constants

**Key Enum:**

```csharp
using Viiper.Client.Devices.Keyboard;

var key = Key.A;               // 0x04
var f1 = Key.F1;               // 0x3A
var enter = Key.Enter;         // 0x28
```

**Modifier Flags:**

```csharp
var mods = (byte)(Mod.LeftShift | Mod.LeftCtrl);  // 0x03
```

**LED Flags:**

```csharp
bool numLock = (leds & (byte)LED.NumLock) != 0;
bool capsLock = (leds & (byte)LED.CapsLock) != 0;
```

### Helper Maps

The SDK generates useful lookup maps for working with keyboard input:

**CharToKey Map** - Convert ASCII characters to key codes:

```csharp
if (CharToKey.TryGetValue((byte)'A', out var key))
{
    Console.WriteLine($"'A' maps to {key}");  // Key.A
}
```

**KeyName Map** - Get human-readable key names:

```csharp
if (KeyName.TryGetValue((byte)Key.F1, out var name))
{
    Console.WriteLine($"Key name: {name}");  // "F1"
}
```

**ShiftChars Map** - Check if a character requires shift:

```csharp
bool needsShift = ShiftChars.ContainsKey((byte)'A');  // true for uppercase
```

## Practical Example: Typing Text

Using the generated maps to type a string:

```csharp
async Task TypeString(ViiperDevice device, string text)
{
    foreach (char c in text)
    {
        if (!CharToKey.TryGetValue((byte)c, out var key))
            continue;
        
        byte mods = ShiftChars.ContainsKey((byte)c) 
            ? (byte)Mod.LeftShift 
            : (byte)0;
        
        // Press
        await device.SendAsync(new KeyboardInput
        {
            Modifiers = mods,
            Count = 1,
            Keys = new[] { (byte)key }
        });
        await Task.Delay(50);
        
        // Release
        await device.SendAsync(new KeyboardInput
        {
            Modifiers = 0,
            Count = 0,
            Keys = Array.Empty<byte>()
        });
        await Task.Delay(50);
    }
}

// Usage
await TypeString(device, "Hello, World!");
```

## Device-Specific Wire Formats

### Keyboard Input

```csharp
public struct KeyboardInput
{
    public byte Modifiers;    // Modifier flags (Ctrl, Shift, Alt, GUI)
    public byte Count;        // Number of keys in Keys array
    public byte[] Keys;       // Key codes (max 6 for HID compliance)
}
```

**Wire format:** 1 byte modifiers + 1 byte count + N bytes keys (variable-length)

### Keyboard Output (LEDs)

```csharp
// Single byte with LED flags
byte leds = data[0];
bool numLock = (leds & (byte)LED.NumLock) != 0;
```

### Xbox360 Input

```csharp
public struct Xbox360Input
{
    public ushort Buttons;     // Button flags
    public byte LeftTrigger;   // 0-255
    public byte RightTrigger;  // 0-255
    public short ThumbLX;      // -32768 to 32767
    public short ThumbLY;      // -32768 to 32767
    public short ThumbRX;      // -32768 to 32767
    public short ThumbRY;      // -32768 to 32767
}
```

**Wire format:** Fixed 14 bytes, packed structure

### Xbox360 Output (Rumble)

```csharp
// Two bytes: left motor + right motor (0-255 each)
byte leftMotor = data[0];
byte rightMotor = data[1];
```

### Mouse Input

```csharp
public struct MouseInput
{
    public byte Buttons;  // Button flags
    public sbyte X;       // Relative X movement (-128 to 127)
    public sbyte Y;       // Relative Y movement (-128 to 127)
    public sbyte Wheel;   // Vertical scroll
    public sbyte Pan;     // Horizontal scroll
}
```

**Wire format:** Fixed 5 bytes, packed structure

## Configuration and Advanced Usage

### Custom Timeouts

```csharp
var client = new ViiperClient("localhost", 3242)
{
    Timeout = TimeSpan.FromSeconds(10)
};
```

Default timeout is 5 seconds.

### Cancellation Tokens

All async methods support cancellation:

```csharp
using var cts = new CancellationTokenSource(TimeSpan.FromSeconds(2));

try
{
    var buses = await client.BusListAsync(cts.Token);
}
catch (OperationCanceledException)
{
    Console.WriteLine("Request timed out");
}
```

### Error Handling

The server returns errors as JSON. The client throws exceptions:

```csharp
try
{
    await client.BusCreateAsync("invalid-bus-id");
}
catch (Exception ex)
{
    Console.WriteLine($"Request failed: {ex.Message}");
}
```

### Resource Management

`ViiperDevice` implements `IDisposable`:

```csharp
await using var device = await client.ConnectDeviceAsync(busId, deviceId);
// Device automatically closed when scope exits
```

Or manual cleanup:

```csharp
try
{
    var device = await client.ConnectDeviceAsync(busId, deviceId);
    // ... use device ...
}
finally
{
    device.Dispose();
}
```

## Examples

Full working examples are available in the repository:

- **Virtual Keyboard**: `examples/csharp/virtual_keyboard/Program.cs`
  - Types "Hello!" every 5 seconds using generated maps
  - Displays LED feedback in console
  
- **Virtual Mouse**: `examples/csharp/virtual_mouse/Program.cs`
  - Moves cursor in a circle pattern
  - Demonstrates button clicks and scroll wheel

- **Virtual Xbox360 Controller**: `examples/csharp/virtual_x360_pad/Program.cs`
  - Presses buttons and moves sticks
  - Handles rumble feedback

### Running Examples

```bash
cd examples/csharp/virtual_keyboard
dotnet run -- localhost
```

## Project Structure

Generated SDK layout:

```text
clients/csharp/Viiper.Client/
├── ViiperClient.cs              # Management API client
├── ViiperDevice.cs              # Device stream wrapper
├── Types/
│   ├── BusListResponse.cs       # API response types
│   ├── BusCreateResponse.cs
│   └── ...
└── Devices/
        ├── Keyboard/
        │   ├── KeyboardInput.cs     # Wire format struct
        │   └── KeyboardConstants.cs # Enums + maps
        ├── Mouse/
        │   ├── MouseInput.cs
        │   └── MouseConstants.cs
        └── Xbox360/
                ├── Xbox360Input.cs
                ├── Xbox360Output.cs
                └── Xbox360Constants.cs
```

## Troubleshooting

**Build Errors:**

Ensure you have .NET 8.0 SDK installed:

```bash
dotnet --version  # Should be 8.0 or higher
```

**Nullable Reference Warnings:**

The generated code uses nullable annotations. You may see warnings like CS8601/CS8625. These are safe to ignore or suppress in your project file:

```xml
<PropertyGroup>
    <NoWarn>$(NoWarn);CS8601;CS8625</NoWarn>
</PropertyGroup>
```

## See Also

- [Generator Documentation](generator.md): How generated SDKs work
- [Go SDK Documentation](go.md): Reference implementation patterns
- [Rust SDK Documentation](rust.md): Rust SDK with sync/async support
- [TypeScript SDK Documentation](typescript.md): Node.js SDK
- [C SDK Documentation](c.md): Alternative SDK for native integration
- [API Overview](../api/overview.md): Management API reference
- [Device Documentation](../devices/): Wire formats and device-specific details

---

For questions or contributions, see the main VIIPER repository.
