package dualsense

import (
	"encoding/binary"
	"math/bits"
)

// Trigger effects stolen from https://gist.github.com/Nielk1/6d54cc2c00d2201ccb8c2720ad7538db
/* MIT License
 *
 * Copyright (c) 2021-2022 John "Nielk1" Klein
 */

type TriggerEffect struct {
	Mode          uint8
	Params        [TriggerParamsSize]uint8
	ZoneStrengths [TriggerZoneCount]uint8
	StartZone     uint8
	EndZone       uint8
	StartPosition uint8
	EndPosition   uint8
	Strength      uint8
	SnapForce     uint8
	AmplitudeA    uint8
	AmplitudeB    uint8
	FirstFoot     uint8
	SecondFoot    uint8
	Frequency     uint8
	Period        uint8
}

func newTriggerEffect(mode uint8, params [TriggerParamsSize]uint8) TriggerEffect {
	effect := TriggerEffect{
		Mode:   mode,
		Params: params,
	}
	zones := binary.LittleEndian.Uint16(params[0:2])

	switch mode {
	case TriggerModeFeedback, TriggerModeVibration:
		packedStrengths := binary.LittleEndian.Uint32(params[2:6])
		for zone := range TriggerZoneCount {
			if zones&(1<<zone) != 0 {
				effect.ZoneStrengths[zone] = uint8(packedStrengths>>(triggerBitsPerStrength*zone))&triggerStrengthMask + 1
			}
		}
		if mode == TriggerModeVibration {
			effect.Frequency = params[triggerFrequencyParam]
		}
	case TriggerModeWeapon, TriggerModeBow, TriggerModeGalloping, TriggerModeMachine:
		if zones != 0 {
			effect.StartZone = uint8(bits.TrailingZeros16(zones))
			effect.EndZone = uint8(bits.Len16(zones) - 1)
		}
		first := params[2] & triggerStrengthMask
		second := params[2] >> triggerBitsPerStrength & triggerStrengthMask
		switch mode {
		case TriggerModeWeapon:
			effect.Strength = first + 1
		case TriggerModeBow:
			effect.Strength = first + 1
			effect.SnapForce = second + 1
		case TriggerModeGalloping:
			effect.SecondFoot = first
			effect.FirstFoot = second
			effect.Frequency = params[3]
		case TriggerModeMachine:
			effect.AmplitudeA = first
			effect.AmplitudeB = second
			effect.Frequency = params[3]
			effect.Period = params[4]
		}
	case TriggerModeSimpleFeedback, TriggerModeLimitedFeedback:
		effect.StartPosition = params[0]
		effect.Strength = params[1]
	case TriggerModeSimpleWeapon, TriggerModeLimitedWeapon:
		effect.StartPosition = params[0]
		effect.EndPosition = params[1]
		effect.Strength = params[2]
	case TriggerModeSimpleVibration:
		effect.Frequency = params[0]
		effect.Strength = params[1]
		effect.StartPosition = params[2]
	}
	return effect
}

type triggerState struct {
	effect      TriggerEffect
	weaponFired bool
}

func (t *triggerState) decode(position uint8) (effect, status uint8) {
	// TODO: position to zone mapping
	zone := uint8(int(position) * TriggerZoneCount / 256)

	switch t.effect.Mode {
	case TriggerModeFeedback:
		if t.effect.ZoneStrengths[zone] != 0 {
			return TriggerEffectFeedback, TriggerStatusFeedbackLoadApplied
		}
		return TriggerEffectFeedback, TriggerStatusFeedbackNoLoad
	case TriggerModeVibration:
		if t.effect.ZoneStrengths[zone] != 0 {
			return TriggerEffectVibration, TriggerStatusVibrationOn
		}
		return TriggerEffectVibration, TriggerStatusVibrationOff
	case TriggerModeWeapon:
		if zone < t.effect.StartZone {
			t.weaponFired = false
			return TriggerEffectWeapon, TriggerStatusWeaponReady
		}
		if zone > t.effect.EndZone {
			t.weaponFired = true
		}
		if t.weaponFired {
			return TriggerEffectWeapon, TriggerStatusWeaponFired
		}
		return TriggerEffectWeapon, TriggerStatusWeaponFiring
	}
	return TriggerEffectNone, 0
}
