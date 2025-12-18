package proxy

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log/slog"

	"github.com/Alia5/VIIPER/usbip"
)

// Parser handles USB-IP packet parsing for structured logging.
type Parser struct {
	logger *slog.Logger
	buf    bytes.Buffer
}

func NewParser(logger *slog.Logger) *Parser {
	return &Parser{
		logger: logger,
	}
}

// Parse processes incoming data and logs USB-IP protocol information.
func (p *Parser) Parse(data []byte, clientToServer bool) {
	p.buf.Write(data)

	for p.buf.Len() >= 8 {
		peek := p.buf.Bytes()

		ver := binary.BigEndian.Uint16(peek[0:2])
		code := binary.BigEndian.Uint16(peek[2:4])

		if ver == usbip.Version {
			switch code {
			case usbip.OpReqDevlist:
				if p.buf.Len() >= 8 {
					p.logMgmtOp("OP_REQ_DEVLIST", clientToServer)
					p.buf.Next(8)
					continue
				}
				return

			case usbip.OpRepDevlist:
				if consumed := p.parseOpRepDevlist(peek, clientToServer); consumed > 0 {
					p.buf.Next(consumed)
					continue
				}
				return

			case usbip.OpReqImport:
				if p.buf.Len() >= 40 { // 8 byte header + 32 byte busid
					busid := peek[8:40]
					end := bytes.IndexByte(busid, 0)
					if end == -1 {
						end = len(busid)
					}
					p.logger.Info("USBIP packet",
						"dir", dirString(clientToServer),
						"op", "OP_REQ_IMPORT",
						"busid", string(busid[:end]))
					p.buf.Next(40)
					continue
				}
				return

			case usbip.OpRepImport:
				if consumed := p.parseOpRepImport(peek, clientToServer); consumed > 0 {
					p.buf.Next(consumed)
					continue
				}
				return
			}
		}

		if p.buf.Len() >= 48 { // 0x30 bytes
			cmd := binary.BigEndian.Uint32(peek[0:4])

			switch cmd {
			case usbip.CmdSubmitCode:
				p.parseCmdSubmit(peek, clientToServer)
				p.buf.Next(48)
				dir := binary.BigEndian.Uint32(peek[12:16])
				xferLen := binary.BigEndian.Uint32(peek[24:28])
				if dir == usbip.DirOut && xferLen > 0 && uint32(p.buf.Len()) >= xferLen {
					p.buf.Next(int(xferLen))
				}
				continue

			case usbip.RetSubmitCode:
				p.parseRetSubmit(peek, clientToServer)
				p.buf.Next(48)
				actualLen := binary.BigEndian.Uint32(peek[24:28])
				if actualLen > 0 && uint32(p.buf.Len()) >= actualLen {
					p.buf.Next(int(actualLen))
				}
				continue

			case usbip.CmdUnlinkCode:
				p.parseCmdUnlink(peek, clientToServer)
				p.buf.Next(48)
				continue

			case usbip.RetUnlinkCode:
				p.parseRetUnlink(peek, clientToServer)
				p.buf.Next(48)
				continue
			}
		}

		if p.buf.Len() > 64*1024 {
			p.logger.Warn("Parser buffer overflow, resetting")
			p.buf.Reset()
		}
		return
	}
}

func (p *Parser) parseCmdSubmit(data []byte, clientToServer bool) {
	seqnum := binary.BigEndian.Uint32(data[4:8])
	devid := binary.BigEndian.Uint32(data[8:12])
	dir := binary.BigEndian.Uint32(data[12:16])
	ep := binary.BigEndian.Uint32(data[16:20])
	xferLen := binary.BigEndian.Uint32(data[24:28])
	setup := data[40:48]

	args := []any{
		"dir", dirString(clientToServer),
		"op", "CMD_SUBMIT",
		"seq", seqnum,
		"devid", devid,
		"ep", ep,
		"urb_dir", urbDirString(dir),
		"len", xferLen,
	}

	if ep == 0 {
		args = append(args, "setup", fmt.Sprintf("[%02x %02x %02x %02x %02x %02x %02x %02x]",
			setup[0], setup[1], setup[2], setup[3], setup[4], setup[5], setup[6], setup[7]))
	}

	p.logger.Info("USBIP packet", args...)
}

func (p *Parser) parseRetSubmit(data []byte, clientToServer bool) {
	seqnum := binary.BigEndian.Uint32(data[4:8])
	status := int32(binary.BigEndian.Uint32(data[20:24]))
	actualLen := binary.BigEndian.Uint32(data[24:28])

	p.logger.Info("USBIP packet",
		"dir", dirString(clientToServer),
		"op", "RET_SUBMIT",
		"seq", seqnum,
		"status", status,
		"actual_len", actualLen)
}

func (p *Parser) parseCmdUnlink(data []byte, clientToServer bool) {
	seqnum := binary.BigEndian.Uint32(data[4:8])
	unlinkSeq := binary.BigEndian.Uint32(data[20:24])

	p.logger.Info("USBIP packet",
		"dir", dirString(clientToServer),
		"op", "CMD_UNLINK",
		"seq", seqnum,
		"unlink_seq", unlinkSeq)
}

func (p *Parser) parseRetUnlink(data []byte, clientToServer bool) {
	seqnum := binary.BigEndian.Uint32(data[4:8])
	status := int32(binary.BigEndian.Uint32(data[20:24]))

	p.logger.Info("USBIP packet",
		"dir", dirString(clientToServer),
		"op", "RET_UNLINK",
		"seq", seqnum,
		"status", status)
}

func (p *Parser) logMgmtOp(op string, clientToServer bool) {
	p.logger.Info("USBIP packet",
		"dir", dirString(clientToServer),
		"op", op)
}

func (p *Parser) parseOpRepDevlist(data []byte, clientToServer bool) int {
	if len(data) < 12 {
		return 0
	}

	nDevices := binary.BigEndian.Uint32(data[8:12])
	p.logger.Info("USBIP packet",
		"dir", dirString(clientToServer),
		"op", "OP_REP_DEVLIST",
		"nDevices", nDevices)

	offset := 12
	for i := uint32(0); i < nDevices; i++ {
		// Device entry: 312 bytes base (path[256] + busid[32] + busid(4) + devid(4) + speed(4) + ids(8) + class(6))
		// Plus 4 bytes per interface (class, subclass, protocol, padding)
		if len(data) < offset+312 {
			return 0 // Need more data
		}

		path := data[offset : offset+256]
		pathEnd := bytes.IndexByte(path, 0)
		if pathEnd == -1 {
			pathEnd = len(path)
		}
		pathStr := string(path[:pathEnd])

		busid := data[offset+256 : offset+288]
		busidEnd := bytes.IndexByte(busid, 0)
		if busidEnd == -1 {
			busidEnd = len(busid)
		}
		busidStr := string(busid[:busidEnd])

		busnum := binary.BigEndian.Uint32(data[offset+288 : offset+292])
		devnum := binary.BigEndian.Uint32(data[offset+292 : offset+296])
		speed := binary.BigEndian.Uint32(data[offset+296 : offset+300])
		idVendor := binary.BigEndian.Uint16(data[offset+300 : offset+302])
		idProduct := binary.BigEndian.Uint16(data[offset+302 : offset+304])
		bcdDevice := binary.BigEndian.Uint16(data[offset+304 : offset+306])
		bDeviceClass := data[offset+306]
		bDeviceSubClass := data[offset+307]
		bDeviceProtocol := data[offset+308]
		bConfigurationValue := data[offset+309]
		bNumConfigurations := data[offset+310]
		bNumInterfaces := data[offset+311]

		p.logger.Info("  Device",
			"path", pathStr,
			"busid", busidStr,
			"bus", busnum,
			"dev", devnum,
			"speed", speed,
			"vid", fmt.Sprintf("%04x", idVendor),
			"pid", fmt.Sprintf("%04x", idProduct),
			"bcd", fmt.Sprintf("%04x", bcdDevice),
			"class", fmt.Sprintf("%02x", bDeviceClass),
			"subclass", fmt.Sprintf("%02x", bDeviceSubClass),
			"protocol", fmt.Sprintf("%02x", bDeviceProtocol),
			"config", bConfigurationValue,
			"nConfigs", bNumConfigurations,
			"nInterfaces", bNumInterfaces)

		offset += 312

		// Parse interfaces (4 bytes each)
		for j := uint8(0); j < bNumInterfaces; j++ {
			if len(data) < offset+4 {
				return 0
			}
			ifClass := data[offset]
			ifSubClass := data[offset+1]
			ifProtocol := data[offset+2]
			p.logger.Info("    Interface",
				"num", j,
				"class", fmt.Sprintf("%02x", ifClass),
				"subclass", fmt.Sprintf("%02x", ifSubClass),
				"protocol", fmt.Sprintf("%02x", ifProtocol))
			offset += 4
		}
	}

	return offset
}

func (p *Parser) parseOpRepImport(data []byte, clientToServer bool) int {
	// OP_REP_IMPORT: 8 byte header + 312 byte device descriptor (same as devlist, but without interfaces)
	if len(data) < 320 {
		return 0
	}

	status := binary.BigEndian.Uint32(data[4:8])

	path := data[8 : 8+256]
	pathEnd := bytes.IndexByte(path, 0)
	if pathEnd == -1 {
		pathEnd = len(path)
	}
	pathStr := string(path[:pathEnd])

	busid := data[264 : 264+32]
	busidEnd := bytes.IndexByte(busid, 0)
	if busidEnd == -1 {
		busidEnd = len(busid)
	}
	busidStr := string(busid[:busidEnd])

	busnum := binary.BigEndian.Uint32(data[296:300])
	devnum := binary.BigEndian.Uint32(data[300:304])
	speed := binary.BigEndian.Uint32(data[304:308])
	idVendor := binary.BigEndian.Uint16(data[308:310])
	idProduct := binary.BigEndian.Uint16(data[310:312])
	bcdDevice := binary.BigEndian.Uint16(data[312:314])
	bDeviceClass := data[314]
	bDeviceSubClass := data[315]
	bDeviceProtocol := data[316]
	bConfigurationValue := data[317]
	bNumConfigurations := data[318]
	bNumInterfaces := data[319]

	p.logger.Info("USBIP packet",
		"dir", dirString(clientToServer),
		"op", "OP_REP_IMPORT",
		"status", status,
		"path", pathStr,
		"busid", busidStr,
		"bus", busnum,
		"dev", devnum,
		"speed", speed,
		"vid", fmt.Sprintf("%04x", idVendor),
		"pid", fmt.Sprintf("%04x", idProduct),
		"bcd", fmt.Sprintf("%04x", bcdDevice),
		"class", fmt.Sprintf("%02x", bDeviceClass),
		"subclass", fmt.Sprintf("%02x", bDeviceSubClass),
		"protocol", fmt.Sprintf("%02x", bDeviceProtocol),
		"config", bConfigurationValue,
		"nConfigs", bNumConfigurations,
		"nInterfaces", bNumInterfaces)

	return 320
}

func dirString(clientToServer bool) string {
	if clientToServer {
		return "C→S"
	}
	return "S→C"
}

func urbDirString(dir uint32) string {
	if dir == usbip.DirOut {
		return "OUT"
	}
	return "IN"
}
