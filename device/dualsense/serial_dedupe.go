package dualsense

import "log/slog"

var (
	serials = map[string]struct{}{}
	macs    = map[string]struct{}{}
)

func (d *DualSense) ReleaseSerialAndMAC() {
	d.mtx.Lock()
	serial := d.metaState.SerialNumber
	mac := d.metaState.MACAddress
	d.mtx.Unlock()
	delete(serials, serial)
	delete(macs, mac)
	slog.Debug("DualSense disconnected, serial/mac released", "serial", serial, "mac", mac)
}
