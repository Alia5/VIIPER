import { ViiperClient, ViiperDevice, Mouse, Types } from "viiperclient";

const { MouseInput, Btn_ } = Mouse;

const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms));

// Minimal example: ensure a bus, create a mouse device, stream inputs, clean up on exit.
async function main() {
  if (process.argv.length < 3) {
    console.log("Usage: node virtual_mouse.js <api_addr>");
    console.log("Example: node virtual_mouse.js localhost:3242");
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
    const req: Types.DeviceCreateRequest = { type: "mouse" };
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

  // Send a short movement once every 3 seconds for easy local testing.
  // Followed by a short click and a single scroll notch.
  let dir = 1;
  const step = 50; // move diagonally by 50 px in X and Y
  let running = true;

  console.log("Every 3s: move diagonally by 50px (X and Y), then click and scroll. Press Ctrl+C to stop.");

  const interval = setInterval(async () => {
    if (!running) return;
    
    try {
      // Move diagonally: (+step,+step) then (-step,-step) next tick
      const dx = step * dir;
      const dy = step * dir;
      dir *= -1;

      // One-shot movement report (diagonal)
      const move = new MouseInput({ Buttons: 0, Dx: dx, Dy: dy, Wheel: 0, Pan: 0 });
      await dev.send(move);
      console.log(`→ Moved mouse dx=${dx} dy=${dy}`);

      // Zero state shortly after to keep movement one-shot (harmless safety)
      await sleep(30);
      const zero = new MouseInput({ Buttons: 0, Dx: 0, Dy: 0, Wheel: 0, Pan: 0 });
      await dev.send(zero);

      // Simulate a short left click: press then release
      await sleep(50);
      const press = new MouseInput({ Buttons: Btn_.Left, Dx: 0, Dy: 0, Wheel: 0, Pan: 0 });
      await dev.send(press);
      await sleep(60);
      const rel = new MouseInput({ Buttons: 0x00, Dx: 0, Dy: 0, Wheel: 0, Pan: 0 });
      await dev.send(rel);
      console.log("→ Clicked (left)");

      // Simulate a short scroll: one notch upwards
      await sleep(50);
      const scr = new MouseInput({ Buttons: 0, Dx: 0, Dy: 0, Wheel: 1, Pan: 0 });
      await dev.send(scr);
      await sleep(30);
      const scr0 = new MouseInput({ Buttons: 0, Dx: 0, Dy: 0, Wheel: 0, Pan: 0 });
      await dev.send(scr0);
      console.log("→ Scrolled (wheel=+1)");
    } catch (err) {
      console.error(`Write error: ${err}`);
      running = false;
      clearInterval(interval);
      await cleanup();
      process.exit(1);
    }
  }, 3000);

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
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
