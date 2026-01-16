# TypeScript Client Library Documentation

The VIIPER TypeScript client library provides a modern, type-safe Node.js client library for interacting with VIIPER servers and controlling virtual devices.

The TypeScript client library features:

- **Type-safe API**: Structured request/response types with proper TypeScript definitions
- **Event-driven**: EventEmitter-based output handling for device feedback (LEDs, rumble)
- **Zero external dependencies**: Uses only built-in Node.js libraries

!!! note "License"
    The TypeScript client library is licensed under the **MIT License**, providing maximum flexibility for integration into your projects.  
    The core VIIPER server remains under its original license.

## Installation

### 1. Using the Published Package

Install the client library from the public npm registry:

```bash
npm install viiperclient
```

The latest stable version is tagged as `latest`.

!!! "Pre-Releases"
    Pre-release / snapshot builds are **not** published to npm. They are only available as GitHub Release artifacts (e.g. `dev-latest`) or by building from source.

To use a snapshot artifact from GitHub:

1. Download `viiperclient-typescript-sdk-Snapshot.tgz` (or a versioned tarball) from the appropriate Release.
2. Install it directly:

```bash
npm install ./viiperclient-typescript-sdk-Snapshot.tgz
```

Package page: [npm: viiperclient](https://www.npmjs.com/package/viiperclient)

### 2. Local Project Reference (For Development Against Source)

If you are actively modifying VIIPER or the code generator, link directly:

```json
{
  "dependencies": {
    "viiperclient": "file:../../clients/typescript"
  }
}
```

Then build locally after regeneration:

```bash
cd clients/typescript
npm install
npm run build
```

## Example

```typescript
import { ViiperClient, Keyboard } from "viiperclient";

const { KeyboardInput, Key, Mod } = Keyboard;

// Create new Viiper client
const client = new ViiperClient("localhost", 3242);

// Find or create a bus
const busesResp = await client.buslist();
let busID: number;
if (busesResp.buses.length === 0) {
  const resp = await client.buscreate(); // Auto-assign ID
  // Or specify ID: await client.buscreate(5);
  busID = resp.busId;
} else {
  busID = busesResp.buses[0];
}

// Add device and connect
const deviceReq = { type: "keyboard" };
const { device, response } = await client.addDeviceAndConnect(busID, deviceReq);

console.log(`Connected to device ${response.busId}-${response.devId}`);

// Send keyboard input
const input = new KeyboardInput({
  Modifiers: Mod.LeftShift,
  Count: 1,
  Keys: [Key.H]
});
await device.send(input);

// Cleanup
await client.busdeviceremove(busID, response.devId);
```

## Device Control/Feedback

### Creating a Device + Control/Feedback Stream

The simplest way to add a device and connect:

```typescript
const deviceReq = { type: "xbox360" };
const { device, response } = await client.addDeviceAndConnect(busID, deviceReq);
```

Or manually add and connect:

```typescript
const deviceResp = await client.busdeviceadd(busId, { type: "keyboard" });
const device = await client.connectDevice(busId, deviceResp.devId);
```

Or connect to an existing device:

```typescript
const device = await client.connectDevice(busId, deviceId);
```

### Sending Input

Device input is sent using generated classes:

```typescript
import { Xbox360 } from "viiperclient";

const { Xbox360Input, Button } = Xbox360;

const input = new Xbox360Input({
  Buttons: Button.A,
  Lt: 255,
  Rt: 0,
  Lx: -32768,  // Left stick left
  Ly: 32767,   // Left stick up
  Rx: 0,
  Ry: 0
});
await device.send(input);
```

### Receiving Feedback

For devices that send feedback (rumble, LEDs), subscribe to the `output` event:

```typescript
import { Keyboard } from "viiperclient";

const { LED } = Keyboard;

device.on("output", (data: Buffer) => {
  if (data.length < 1) return;
  const leds = data.readUInt8(0);
  
  console.log(`LEDs: ` +
    `Num=${(leds & LED.NumLock) !== 0} ` +
    `Caps=${(leds & LED.CapsLock) !== 0} ` +
    `Scroll=${(leds & LED.ScrollLock) !== 0}`);
});
```

For Xbox360 rumble:

```typescript
device.on("output", (data: Buffer) => {
  if (data.length < 2) return;
  const leftMotor = data.readUInt8(0);
  const rightMotor = data.readUInt8(1);
  console.log(`Rumble: Left=${leftMotor} Right=${rightMotor}`);
});
```

### Closing a Device

```typescript
device.close();
```

The VIIPER server automatically removes the device when the stream is closed after a short timeout.

### Error Handling and Events

Device streams emit `error` and `end` events that should be handled:

```typescript
device.on("error", async (err: Error) => {
  console.error(`Stream error: ${err}`);
  // Handle error and cleanup
});

device.on("end", async () => {
  console.log("Stream ended by server");
  // Handle disconnection and cleanup
});
```

For long-running applications with intervals or timers, stop them before cleanup:

```typescript
let running = true;
const interval = setInterval(async () => {
  if (!running) return;
  
  try {
    await device.send(input);
  } catch (err) {
    console.error(`Send error: ${err}`);
    running = false;
    clearInterval(interval);
    // Cleanup...
  }
}, 16);

// Handle Ctrl+C gracefully
process.on("SIGINT", async () => {
  console.log("Stopping...");
  running = false;
  clearInterval(interval);
  device.close();
  await client.busdeviceremove(busId, deviceId);
  process.exit(0);
});
```

## Generated Constants and Maps

The TypeScript SDK automatically generates enums and helper maps for each device type.

### Keyboard Constants

**Key Enum:**

```typescript
import { Keyboard } from "viiperclient";

const { Key } = Keyboard;

const key = Key.A;               // 0x04
const f1 = Key.F1;               // 0x3A
const enter = Key.Enter;         // 0x28
```

**Modifier Flags:**

```typescript
import { Keyboard } from "viiperclient";

const { Mod } = Keyboard;

const mods = Mod.LeftShift | Mod.LeftCtrl;  // 0x03
```

**LED Flags:**

```typescript
import { Keyboard } from "viiperclient";

const numLock = (leds & LED.NumLock) !== 0;
const capsLock = (leds & LED.CapsLock) !== 0;
```

### Helper Maps

The client library generates useful lookup maps for working with keyboard input:

**CharToKey Map** - Convert ASCII characters to key codes:

```typescript
import { Keyboard } from "viiperclient";

const { CharToKeyGet } = Keyboard;

const key = CharToKeyGet('A'.codePointAt(0)!);
if (key !== undefined) {
  console.log(`'A' maps to ${key}`);  // Key.A
}
```

**KeyName Map** - Get human-readable key names:

```typescript
import { Keyboard } from "viiperclient";

const { KeyNameGet } = Keyboard;

const name = KeyNameGet(Key.F1);
if (name !== undefined) {
  console.log(`Key name: ${name}`);  // "F1"
}
```

**ShiftChars Map** - Check if a character requires shift:

```typescript
import { Keyboard } from "viiperclient";

const { ShiftCharsHas } = Keyboard;

const needsShift = ShiftCharsHas('A'.codePointAt(0)!);  // true for uppercase
```

### Error Handling

The server returns errors as JSON. The client throws exceptions:

```typescript
try {
  await client.buscreate("invalid-bus-id");
} catch (err) {
  console.error(`Request failed: ${err}`);
}
```

Device Control/Feedback stream errors are surfaced through the EventEmitter error event:

```typescript
device.on('error', (err) => {
  console.error(`Stream error: ${err}`);
});
```

### Resource Management

Always close devices when done:

```typescript
try {
  const device = await client.connectDevice(busId, deviceId);
  // ... use device ...
} finally {
  device.close();
}
```

## Examples

Full working examples are available in the repository:

- **Virtual Keyboard**: `examples/typescript/virtual_keyboard.ts`
  - Types "Hello!" every 5 seconds using generated maps
  - Displays LED feedback in console
  
- **Virtual Mouse**: `examples/typescript/virtual_mouse.ts`
  - Moves cursor diagonally
  - Demonstrates button clicks and scroll wheel

- **Virtual Xbox360 Controller**: `examples/typescript/virtual_x360_pad.ts`
  - Runs at 60fps with cycling buttons and animated triggers
  - Handles rumble feedback

### Running Examples

```bash
cd examples/typescript
npm install
npm run build

node dist/virtual_keyboard.js localhost:3242
```

## See Also

- [Generator Documentation](generator.md): How generated client libraries work
- [Go Client Documentation](go.md): Reference implementation patterns
- [C# Client Library Documentation](csharp.md): Alternative managed language client library
- [Rust Client Library Documentation](rust.md): Rust client library with sync/async support
- [C Client Library Documentation](c.md): Alternative client library for native integration
- [C++ Client Library Documentation](cpp.md): Header-only C++ client library
- [API Overview](../api/overview.md): Management API reference
- [Device Documentation](../devices/): Wire formats and device-specific details
