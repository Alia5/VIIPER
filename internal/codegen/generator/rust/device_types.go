package rust

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Alia5/VIIPER/internal/codegen/common"
	"github.com/Alia5/VIIPER/internal/codegen/meta"
	"github.com/Alia5/VIIPER/internal/codegen/scanner"
)

const deviceInputTemplate = `{{.Header}}
use crate::wire::DeviceInput;

#[derive(Debug, Clone, Default)]
pub struct {{.StructName}} {
{{range .Fields}}    pub {{.RustName}}: {{.RustType}},
{{end}}}

impl DeviceInput for {{.StructName}} {
    fn to_bytes(&self) -> Vec<u8> {
        let mut buf = Vec::new();
{{range .Fields}}{{if .IsArray}}
        for item in &self.{{.RustName}} {
            buf.extend_from_slice(&item.to_le_bytes());
        }
{{else}}        buf.extend_from_slice(&self.{{.RustName}}.to_le_bytes());
{{end}}{{end}}        buf
    }
}
`

const deviceOutputTemplate = `{{.Header}}
use crate::wire::DeviceOutput;

#[derive(Debug, Clone, Default)]
pub struct {{.StructName}} {
{{range .Fields}}    pub {{.RustName}}: {{.RustType}},
{{end}}}

impl DeviceOutput for {{.StructName}} {
    fn from_bytes(buf: &[u8]) -> Result<Self, crate::error::ViiperError> {
        let mut offset = 0;
{{range .Fields}}{{if .IsArray}}        // Read array (size determined by remaining bytes)
        let elem_size = std::mem::size_of::<{{.ElementType}}>();
        let mut {{.RustName}} = Vec::new();
        while offset + elem_size <= buf.len() {
            let bytes = &buf[offset..offset + elem_size];
            let value = {{.ElementType}}::from_le_bytes(bytes.try_into().unwrap());
            {{.RustName}}.push(value);
            offset += elem_size;
        }
{{else}}        if offset + std::mem::size_of::<{{.RustType}}>() > buf.len() {
            return Err(crate::error::ViiperError::UnexpectedResponse(
                "buffer too short".into()
            ));
        }
        let {{.RustName}} = {{.RustType}}::from_le_bytes(
            buf[offset..offset + std::mem::size_of::<{{.RustType}}>()]
                .try_into()
                .unwrap()
        );
        offset += std::mem::size_of::<{{.RustType}}>();
{{end}}{{end}}        let _ = offset; // Suppress unused warning for last field
        Ok(Self {
{{range .Fields}}            {{.RustName}},
{{end}}        })
    }
}
`

type rustWireField struct {
	Name        string
	RustName    string
	RustType    string
	ElementType string
	IsArray     bool
	CountName   string
}

type deviceTypeData struct {
	Header     string
	DeviceName string
	StructName string
	Fields     []rustWireField
}

func generateDeviceTypes(logger *slog.Logger, deviceDir string, deviceName string, md *meta.Metadata) error {
	logger.Debug("Generating Rust device types", "device", deviceName)

	if md.WireTags == nil {
		return nil
	}

	pascalDevice := common.ToPascalCase(deviceName)

	c2sTag := md.WireTags.GetTag(deviceName, "c2s")
	if c2sTag != nil {
		path := filepath.Join(deviceDir, "input.rs")
		if err := generateDeviceWireStruct(path, pascalDevice, "Input", c2sTag, deviceInputTemplate); err != nil {
			return err
		}
	}

	s2cTag := md.WireTags.GetTag(deviceName, "s2c")
	if s2cTag != nil {
		path := filepath.Join(deviceDir, "output.rs")
		if err := generateDeviceWireStruct(path, pascalDevice, "Output", s2cTag, deviceOutputTemplate); err != nil {
			return err
		}
	}

	return nil
}

func generateDeviceWireStruct(outputPath, deviceName, className string, tag *scanner.WireTag, tmplStr string) error {
	var fields []rustWireField

	for _, field := range tag.Fields {
		rustName := common.ToSnakeCase(field.Name)
		wireType := field.Type
		baseType := wireType

		isArray := isWireArray(field.Spec)
		var countName string
		var elemType string

		if isArray {
			countName = extractArrayCount(field.Spec)
			idx := strings.Index(wireType, "*")
			if idx > 0 {
				baseType = wireType[:idx]
			}
			elemType = wireTypeToRust(baseType)
		}

		rustType := wireTypeToRust(baseType)
		if isArray {
			rustType = fmt.Sprintf("Vec<%s>", elemType)
		}

		fields = append(fields, rustWireField{
			Name:        field.Name,
			RustName:    rustName,
			RustType:    rustType,
			ElementType: elemType,
			IsArray:     isArray,
			CountName:   countName,
		})
	}

	data := deviceTypeData{
		Header:     writeFileHeaderRust(),
		DeviceName: deviceName,
		StructName: deviceName + className,
		Fields:     fields,
	}

	funcMap := template.FuncMap{}
	tmpl, err := template.New("devicewire").Funcs(funcMap).Parse(tmplStr)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	return nil
}
