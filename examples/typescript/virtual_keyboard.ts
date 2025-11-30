import { ViiperClient, ViiperDevice, Keyboard, Types } from "viiperclient";

const { KeyboardInput, Key, Mod, CharToKeyGet, ShiftCharsHas } = Keyboard;

const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms));

async function main() {
  if (process.argv.length < 3) {
    console.log("Usage: node virtual_keyboard.js <api_addr>");
    console.log("Example: node virtual_keyboard.js localhost:3242");
    process.exit(1);
  }

  const addr = process.argv[2];
  const [host, portStr] = addr.split(':');
  const port = portStr ? parseInt(portStr, 10) : 3242;
  const client = new ViiperClient(host, port);

  // Find or create a bus
  const busesResp = await client.buslist();
  let busID: number;
  let createdBus = false;
  
  if (busesResp.buses.length === 0) {
    try {
      const r = await client.buscreate();
      busID = r.busId;
      createdBus = true;
      console.log(`Created bus ${busID}`);
    } catch (err) {
      console.error(`BusCreate failed: ${err}`);
      process.exit(1);
    }
  } else {
    busID = Math.min(...busesResp.buses);
    console.log(`Using existing bus ${busID}`);
  }

  // Add device and connect to stream in one call
  let dev: ViiperDevice;
  let deviceDevId: string;
  try {
    const req: Types.DeviceCreateRequest = { type: "keyboard" };
    const { device, response: addResp } = await client.addDeviceAndConnect(busID, req);
    dev = device;
    deviceDevId = addResp.devId;
    console.log(`Created and connected to device ${deviceDevId} on bus ${busID}`);
  } catch (err) {
    console.error(`AddDeviceAndConnect error: ${err}`);
    if (createdBus) {
      await client.busremove(busID).catch(() => {});
    }
    process.exit(1);
  }

  // Cleanup function
  const cleanup = async () => {
    try {
      dev.close();
      await client.busdeviceremove(busID, deviceDevId);
      console.log(`Removed device ${deviceDevId}`);
    } catch (err) {
      console.error(`DeviceRemove error: ${err}`);
    }
    if (createdBus) {
      try {
        await client.busremove(busID);
        console.log(`Removed bus ${busID}`);
      } catch (err) {
        console.error(`BusRemove error: ${err}`);
      }
    }
  };

  // Handle LED outputs (1 byte per LED state change)
  dev.on("output", (buf: Buffer) => {
    if (buf.length >= 1) {
      const leds = buf.readUInt8(0);
      const numLock = (leds & 0x01) !== 0;
      const capsLock = (leds & 0x02) !== 0;
      const scrollLock = (leds & 0x04) !== 0;
      const compose = (leds & 0x08) !== 0;
      const kana = (leds & 0x10) !== 0;
      console.log(`→ LEDs: Num=${numLock} Caps=${capsLock} Scroll=${scrollLock} Compose=${compose} Kana=${kana}`);
    }
  });

  // Helper to type a string character by character
  const typeString = async (text: string) => {
    for (const ch of text) {
      const codePoint = ch.codePointAt(0)!;
      const key = CharToKeyGet(codePoint);
      if (key === undefined) continue;

      let mods = 0;
      if (ShiftCharsHas(codePoint)) {
        mods |= Mod.LeftShift;
      }

      // Key down
      const down = new KeyboardInput({ Modifiers: mods, Count: 1, Keys: [key] });
      await dev.send(down);
      await sleep(100);

      // Key up
      const up = new KeyboardInput({ Modifiers: 0, Count: 0, Keys: [] });
      await dev.send(up);
      await sleep(100);
    }
  };

  // Helper to press and release a key
  const pressKey = async (key: number) => {
    const press = new KeyboardInput({ Modifiers: 0, Count: 1, Keys: [key] });
    await dev.send(press);
    await sleep(100);
    const release = new KeyboardInput({ Modifiers: 0, Count: 0, Keys: [] });
    await dev.send(release);
  };

  dev.on("error", async (err: Error) => {
    console.error(`Stream error: ${err}`);
    running = false;
    clearInterval(interval);
    await cleanup();
    process.exit(1);
  });

  dev.on("end", async () => {
    console.log("Stream ended by server");
    running = false;
    clearInterval(interval);
    await cleanup();
    process.exit(0);
  });

  // Handle signals for graceful shutdown
  process.on("SIGINT", async () => {
    console.log("Signal received, stopping…");
    running = false;
    clearInterval(interval);
    await cleanup();
    process.exit(0);
  });
  process.on("SIGTERM", async () => {
    console.log("Signal received, stopping…");
    running = false;
    clearInterval(interval);
    await cleanup();
    process.exit(0);
  });

  console.log("Every 5s: type 'Hello!' + Enter. Press Ctrl+C to stop.");

  // Type "Hello!" + Enter every 5 seconds
  let running = true;
  const interval = setInterval(async () => {
    if (!running) return;
    
    try {
      await typeString("Hello!");
      await sleep(100);
      await pressKey(Key.Enter);
      console.log("→ Typed: Hello!");
    } catch (err) {
      console.error(`Write error: ${err}`);
      running = false;
      clearInterval(interval);
      await cleanup();
      process.exit(1);
    }
  }, 5000);
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
