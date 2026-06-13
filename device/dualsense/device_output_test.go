package dualsense

import (
	"context"
	"testing"

	"github.com/Alia5/VIIPER/usbip"
)

func TestDualSenseOutputReportFromEndpoint(t *testing.T) {
	dev, err := New(nil)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	var got OutputState
	called := false
	dev.SetOutputCallback(func(out OutputState) {
		got = out
		called = true
	})

	report := make([]byte, OutputReportSize)
	report[0] = ReportIDOutput
	report[3] = 0x22
	report[4] = 0x88
	report[11] = 0x21
	report[12] = 0xFC
	report[13] = 0x03
	report[20] = 0x44
	report[22] = 0x25
	report[23] = 0x40
	report[24] = 0x05
	report[31] = 0x55

	dev.HandleTransfer(context.Background(), EndpointOut, usbip.DirOut, report)

	if !called {
		t.Fatal("expected output callback")
	}
	if got.RumbleSmall != 0x22 || got.RumbleLarge != 0x88 {
		t.Fatalf("unexpected rumble: small=%#x large=%#x", got.RumbleSmall, got.RumbleLarge)
	}
	if got.TriggerR2Mode != 0x21 || got.TriggerR2StartResistance != 0xFC ||
		got.TriggerR2EffectForce != 0x03 || got.TriggerR2Frequency != 0x44 {
		t.Fatalf("unexpected R2 trigger feedback: %#v", got)
	}
	if got.TriggerL2Mode != 0x25 || got.TriggerL2StartResistance != 0x40 ||
		got.TriggerL2EffectForce != 0x05 || got.TriggerL2Frequency != 0x55 {
		t.Fatalf("unexpected L2 trigger feedback: %#v", got)
	}
}

func TestDualSenseOutputSetReportWithoutReportId(t *testing.T) {
	dev, err := New(nil)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	var got OutputState
	called := false
	dev.SetOutputCallback(func(out OutputState) {
		got = out
		called = true
	})

	payload := make([]byte, OutputReportSize-1)
	payload[2] = 0x33
	payload[3] = 0x99

	_, handled := dev.HandleControl(hidClassOUT, hidSetReport,
		uint16(reportTypeOutput)<<8|uint16(ReportIDOutput),
		0, uint16(len(payload)), payload)

	if !handled {
		t.Fatal("expected SET_REPORT output to be handled")
	}
	if !called {
		t.Fatal("expected output callback")
	}
	if got.RumbleSmall != 0x33 || got.RumbleLarge != 0x99 {
		t.Fatalf("unexpected rumble: small=%#x large=%#x", got.RumbleSmall, got.RumbleLarge)
	}
}
