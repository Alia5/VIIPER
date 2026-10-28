# HID proxy (copy of another device)

A device that enumerates with descriptors the client supplies and passes reports between the host and the client.
The client reads and writes the real device itself and decides what passes; VIIPER never talks to the real device.
Example use: the real device is hidden from an application (e.g. with HidHide) and the application sees this copy
instead, so the client decides what reaches it. [Arcadia](https://github.com/jc-sw-git/Arcadia) does this for the
Steam Controller puck, but nothing in the device is specific to one kind of hardware.

Use `hidproxy` as the device type. `deviceSpecific` holds the descriptors as the real device reports them, as hex:

```json
{
  "device": "<18-byte device descriptor>",
  "config": "<whole configuration descriptor>",
  "strings": { "1": "Maker", "2": "Product", "3": "<serial number>" },
  "reports": { "<HID interface number>": "<report descriptor>" },
  "speed": 2
}
```

USB 2.01 is served as 2.00 (Windows would otherwise ask for a BOS descriptor). The device's arguments in API
responses are empty, so the serial number isn't echoed into responses or logs.

## Streaming protocol

Frames both ways: kind (1 byte), interface number (1 byte), payload length (uint16, little-endian), payload
(at most 4106 bytes).

| Kind | Direction | Payload |
|---|---|---|
| `0x01` Input | client → VIIPER | an input report for the interface's IN endpoint (interrupt preferred), as sent over USB |
| `0x02` Control reply | client → VIIPER | sequence (uint16), status (0 done, else refused), the IN data stage |
| `0x81` Output | VIIPER → client | what the host sent to an OUT endpoint of the interface |
| `0x82` Control | VIIPER → client | sequence (uint16), the 8-byte setup packet, the OUT data stage; interface `0xFF` for requests to the device |

- Input isn't repeated: when no new report came within `bInterval`, the host gets none (the device implements
  `usb.NoInRepeatDevice`). When the host doesn't poll fast enough, the oldest of 64 queued reports goes.
- Control requests VIIPER answers itself: standard requests, a HID interface's idle rate and protocol
  (`SET_IDLE`, `GET_IDLE`, `SET_PROTOCOL`, `GET_PROTOCOL`), a CDC interface's line coding and control line state
  (`SET_LINE_CODING`, `GET_LINE_CODING`, `SET_CONTROL_LINE_STATE`). Every other request (HID `GET_REPORT`/`SET_REPORT`,
  other CDC requests such as `SEND_BREAK`, vendor and unknown class requests) goes to the client, which answers within
  1 s. VIIPER's server can't stall a request: no client within that time, or a refusal, answers with an empty,
  successful data stage.
- One client at a time; another stream to the same device is refused until the first one ends.
