import { ViiperClient, ViiperDevice, Xbox360, Types } from "viiperclient";

const { Xbox360Input, Button } = Xbox360;

const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms));

// Minimal example: ensure a bus, create an xbox360 device, stream inputs, read rumble, clean up on exit.
async function main() {
  if (process.argv.length < 3) {
    console.log("Usage: node virtual_x360_pad.js <api_addr>");
    console.log("Example: node virtual_x360_pad.js localhost:3242");
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
    let createErr: any;
    busID = 0;
    for (let tryBus = 1; tryBus <= 100; tryBus++) {
      try {
        const r = await client.buscreate(tryBus);
        busID = r.busId;
        createdBus = true;
        break;
      } catch (err) {
        createErr = err;
      }
    }
    if (busID === 0) {
      console.error(`BusCreate failed: ${createErr}`);
      process.exit(1);
    }
    console.log(`Created bus ${busID}`);
  } else {
    busID = Math.min(...busesResp.buses);
    console.log(`Using existing bus ${busID}`);
  }

  // Add device and connect to stream in one call
  let dev: ViiperDevice;
  let deviceDevId: string;
  try {
    const req: Types.DeviceCreateRequest = { type: "xbox360" };
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

  // Start event-driven rumble reading (2 bytes per rumble state)
  dev.on("output", (buf: Buffer) => {
    if (buf.length >= 2) {
      const leftMotor = buf.readUInt8(0);
      const rightMotor = buf.readUInt8(1);
      console.log(`← Rumble: Left=${leftMotor}, Right=${rightMotor}`);
    }
  });

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

  // Send controller inputs at 60fps (16ms intervals)
  let frame = 0;
  let running = true;
  const interval = setInterval(async () => {
    if (!running) return;
    
    try {
      frame++;
      let buttons = 0;
      switch (Math.floor((frame / 60) % 4)) {
        case 0:
          buttons = Button.A;
          break;
        case 1:
          buttons = Button.B;
          break;
        case 2:
          buttons = Button.X;
          break;
        default:
          buttons = Button.Y;
          break;
      }
      
      const state = new Xbox360Input({
        Buttons: buttons,
        Lt: (frame * 2) % 256,
        Rt: (frame * 3) % 256,
        Lx: Math.floor(20000.0 * 0.7071),
        Ly: Math.floor(20000.0 * 0.7071),
        Rx: 0,
        Ry: 0,
      });
      
      await dev.send(state);
      
      if (frame % 60 === 0) {
        console.log(`→ Sent input (frame ${frame}): buttons=0x${state.Buttons.toString(16).padStart(4, "0")}, LT=${state.Lt}, RT=${state.Rt}`);
      }
    } catch (err) {
      console.error(`Write error: ${err}`);
      running = false;
      clearInterval(interval);
      await cleanup();
      process.exit(1);
    }
  }, 16);
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
