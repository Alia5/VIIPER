package xbox360

// Button bitmasks for Xbox 360 controller (XInput compatible)
const (
	ButtonDPadUp    = 0x0001
	ButtonDPadDown  = 0x0002
	ButtonDPadLeft  = 0x0004
	ButtonDPadRight = 0x0008
	ButtonStart     = 0x0010
	ButtonBack      = 0x0020
	ButtonLThumb    = 0x0040 // Left stick button
	ButtonRThumb    = 0x0080 // Right stick button
	ButtonLShoulder = 0x0100 // Left bumper (LB)
	ButtonRShoulder = 0x0200 // Right bumper (RB)
	ButtonGuide     = 0x0400 // Xbox/Guide button (center logo)
	ButtonA         = 0x1000
	ButtonB         = 0x2000
	ButtonX         = 0x4000
	ButtonY         = 0x8000
)

const {
	SubtypeGamepad                         = 1
    SubtypeWheel                           = 2
    SubtypeArcadeStick                     = 3
    SubtypeFlightStick                     = 4
    SubtypeDancePad                        = 5
    SubtypeGuitar                          = 6
    SubtypeGuitarAlternate                 = 7
    SubtypeDrums                           = 8
    SubtypeStageKit                        = 9
    SubtypeGuitarBass                      = 11
    SubtypeProKeys                         = 15
    SubtypeArcadePad                       = 19
    SubtypeTurntable                       = 23
    SubtypeProGuitar                       = 25
    SubtypeDisneyInfinityAndLegoDimensions = 33
    SubtypeSkylanders                      = 36
}