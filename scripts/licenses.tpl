{{ $viiperVersion := "VERSION_PLACEHOLDER" }}

# github.com/Alia5/VIIPER

* Name: github.com/Alia5/VIIPER
* Version: {{ $viiperVersion }}
* License: [GPL-3.0](https://github.com/Alia5/VIIPER/blob/HEAD/LICENSE.txt)

VIIPER - Virtual Input over IP EmulatoR

Copyright (C) 2025-2026 Peter Repukat

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.

{{ range . }}
## {{ .Name }}

* Name: {{ .Name }}
* Version: {{ .Version }}
* License: [{{ .LicenseName }}]({{ .LicenseURL }})

{{ if and .LicensePath (ne .LicensePath "Unknown") }}
{{ .LicenseText }}
{{ end }}

{{ end }}