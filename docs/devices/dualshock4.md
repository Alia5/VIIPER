# DualShock 4 Controller

The DualShock 4 virtual gamepad emulates a complete PlayStation 4 Controller (V1)
 connected via USB.  
It supports sticks, triggers, D-pad, face/shoulder buttons, PS button,
 touchpad click, IMU (gyro + accelerometer), and touchpad finger coordinates.

Use `dualshock4` as the device type when adding a device via the API or client libraries.

## Client Library Support

The wire protocol is abstracted by client libraries.  
The **Go client** includes built-in types (`/device/dualshock4`),
and **generated client libraries** provide equivalent structures
with proper packing.  

You don't need to manually construct packets, just use the provided types
and send/receive them via the device control and feedback stream.

See: [API Reference](../api/overview.md)

## (RAW) Streaming protocol

The device stream is a bidirectional, raw TCP connection with fixed-size packets.

### Input State

- 31-byte packets, little-endian layout:
    - Sticks: StickLX, StickLY, StickRX, StickRY: int8 each (4 bytes)  
      -128 to 127 per axis (-128=min, 0=center, 127=max)
    - Buttons: uint16 (2 bytes, bitfield)
    - DPad: uint8 (1 byte, bitfield)
    - Triggers: TriggerL2, TriggerR2: uint8, uint8 (2 bytes)  
      0-255 (0=not pressed, 255=fully pressed)
    - Touch1: Touch1X, Touch1Y: uint16 each, Touch1Active: bool (5 bytes)
    - Touch2: Touch2X, Touch2Y: uint16 each, Touch2Active: bool (5 bytes)
    - Gyroscope: GyroX, GyroY, GyroZ: int16 each (6 bytes, fixed-point °/s)
    - Accelerometer: AccelX, AccelY, AccelZ: int16 each (6 bytes, fixed-point m/s²)

See `/device/dualshock4/inputstate.go` for details.

### Feedback (Rumble & LED)

- 7-byte packets:
    - RumbleSmall: uint8, RumbleLarge: uint8 (2 bytes)  
      0-255 intensity values
    - LED Color: LedRed, LedGreen, LedBlue: uint8 each (3 bytes)  
      0-255 per channel
    - LED Flash: FlashOn, FlashOff: uint8 each (2 bytes)  
      Units of 2.5ms per value

See `/device/dualshock4/inputstate.go` for the `OutputState` wire definition.

## Reference

### Button Constants

| Button | Hex Value |
| -------- | ----------- |
| Square button | 0x0010 |
| Cross (X) button | 0x0020 |
| Circle button | 0x0040 |
| Triangle button | 0x0080 |
| L1 (Left bumper) | 0x0100 |
| R1 (Right bumper) | 0x0200 |
| L2 button | 0x0400 |
| R2 button | 0x0800 |
| Share button | 0x1000 |
| Options button | 0x2000 |
| L3 (Left stick button) | 0x4000 |
| R3 (Right stick button) | 0x8000 |
| PS button | 0x0001 |
| Touchpad click | 0x0002 |

### D-Pad Constants

| D-Pad Direction | Hex Value |
| --------------- | ----------- |
| Up | 0x01 |
| Down | 0x02 |
| Left | 0x04 |
| Right | 0x08 |

### Touchpad Coordinates

Touch coordinates are sent as `Touch{1,2}X: uint16` and `Touch{1,2}Y: uint16` plus an explicit boolean `Touch{1,2}Active`.

VIIPER clamps touch coordinates to the DS4 range:

- X: **0..1920**
- Y: **0..942**

These are the bounds used by VIIPER’s DS4 implementation; see `/device/dualshock4/const.go`.

### IMU (Gyro + Accelerometer)

#### Fixed-Point Physical Units

VIIPER uses **fixed-point physical units** for IMU values on the wire (still stored as `int16`), to avoid float serialization differences across client languages.

Constants (see `/device/dualshock4/const.go`):

- `GyroCountsPerDps = 16`
- `AccelCountsPerMS2 = 512`

#### Conversion Formulas

**Gyro (degrees/second):**

```text
raw_gyro = round(gyro_dps * GyroCountsPerDps)
gyro_dps = raw_gyro / GyroCountsPerDps
```

**Accelerometer (m/s²):**

```text
raw_accel = round(accel_ms2 * AccelCountsPerMS2)
accel_ms2 = raw_accel / AccelCountsPerMS2
```

#### Resolution and range

With the default scales:

- **Gyro** (`GyroCountsPerDps = 16`):
    - Resolution: `1/16 = 0.0625 °/s`
    - Approx max magnitude: `32767/16 ≈ 2048 °/s`
- **Accelerometer** (`AccelCountsPerMS2 = 512`):
    - Resolution: `1/512 ≈ 0.001953125 m/s²`
    - Approx max magnitude: `32767/512 ≈ 64 m/s²` (≈ 6.5 g)

Conversions saturate to the `int16` range if inputs exceed representable values.

#### Default (Neutral) Report Gravity

On device creation, VIIPER initializes the accelerometer to represent a controller lying flat on a table, with gravity "downwards":

- `g = 9.81 m/s²`
- Default accel is: `(0, 0, -g)`

In raw fixed-point units, this means:

- `AccelX = 0`
- `AccelY = 0`
- `AccelZ = round(-9.81 * 512) = -5023`

Helpers for converting between physical units and raw values are provided in `/device/dualshock4/helpers.go`.
