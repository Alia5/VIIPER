set windows-shell := ["powershell.exe", "-NoProfile", "-Command"]

binary_name := "viiper"
main_pkg := "./cmd/viiper"
src_dir := "."
dist_dir := "dist"
target_goos := env_var_or_default("GOOS", if os_family() == "windows" { "windows" } else { "linux" })
target_goarch := env_var_or_default("GOARCH", if os_family() == "windows" { "amd64" } else { "amd64" })
exe_ext := if target_goos == "windows" { ".exe" } else { "" }
mkdir_p := if os_family() == "windows" { "New-Item -ItemType Directory -Force" } else { "mkdir -p" }
rm_rf := if os_family() == "windows" { "Remove-Item -Recurse -Force -ErrorAction 0" } else { "rm -rf" }
rm_f := if os_family() == "windows" { "Remove-Item -Force -ErrorAction 0" } else { "rm -f" }

version := env_var_or_default("VERSION", `git describe --tags --match "v[0-9]*.[0-9]*.[0-9]*" --always`)
commit := `git rev-parse --short HEAD`
build_time := if os_family() == "windows" {
    `Get-Date -Format 'yyyy-MM-ddTHH:mm:ssZ'`
} else {
    `date -u '+%Y-%m-%dT%H:%M:%SZ'`
}
build_type := env_var_or_default("BUILD_TYPE", "Debug")
output_name := env_var_or_default("OUTPUT_NAME", binary_name + exe_ext)
build_path := join(dist_dir, output_name)
go_licenses_cmd := "go run github.com/google/go-licenses/v2@v2.0.1"
licenses_template := "scripts/licenses.tpl"
licenses_template_work := if os_family() == "windows" { join(env_var_or_default("TEMP", "."), "viiper-licenses.rendered.tpl") } else { "/tmp/viiper-licenses.rendered.tpl" }
licenses_ignore := "github.com/Alia5/VIIPER,github.com/alecthomas/kong-yaml"
licenses_dir := join(dist_dir, "libVIIPER")
licenses_out := join(dist_dir, "licenses.txt")
lib_licenses_out := join(licenses_dir, "licenses.txt")

ldflags_common := "-X main.Version=" + version + " -X main.Commit=" + commit + " -X main.Date=" + build_time + " -X github.com/Alia5/VIIPER/internal/codegen/common.Version=" + version
ldflags_release := "-s -w " + ldflags_common

default:
	just --list

help:
	just --list

tidy:
	go mod tidy

test:
	go test -count=1 -v ./...

test-coverage:
	go test -count=1 -v -coverpkg="./..." -coverprofile="coverage.txt" ./...

[windows]
generate-versioninfo:
	go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
	pwsh -NoProfile -NonInteractive -File scripts/inject-version.ps1 "{{ version }}" "versioninfo.json" "versioninfo.tmp.json"
	{{
		if target_goarch == "amd64" {
			"goversioninfo -64 -o cmd/viiper/resource.syso versioninfo.tmp.json"
		} else if target_goarch == "arm64" {
			"goversioninfo -arm -64 -o cmd/viiper/resource.syso versioninfo.tmp.json"
		} else {
			"goversioninfo -o cmd/viiper/resource.syso versioninfo.tmp.json"
		}
	}}

[unix]
generate-versioninfo:
	@:

clean-versioninfo:
	-{{ rm_f }} cmd/viiper/resource.syso
	-{{ rm_f }} lib/viiper/resource.syso
	-{{ rm_f }} versioninfo.tmp.json
	-{{ rm_f }} libviiper.versioninfo.tmp.json

[arg("type", pattern="Debug|Release")]
[windows]
build type=build_type: generate-versioninfo
	{{ mkdir_p }} {{ dist_dir }}
	$env:CGO_ENABLED='0'; go build -trimpath -ldflags "{{ if type == "Release" { ldflags_release } else { ldflags_common } }}" -o {{ build_path }} {{ main_pkg }}
	just licenses

[arg("type", pattern="Debug|Release")]
[unix]
build type=build_type:
	{{ mkdir_p }} {{ dist_dir }}
	CGO_ENABLED=0 go build -trimpath -ldflags "{{ if type == "Release" { ldflags_release } else { ldflags_common } }}" -o {{ build_path }} {{ main_pkg }}
	just licenses

[arg("type", pattern="Debug|Release")]
[windows]
build-libVIIPER type=build_type:
	{{ mkdir_p }} dist/libVIIPER
	go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
	pwsh -NoProfile -NonInteractive -File scripts/inject-version.ps1 "{{ version }}" "lib/viiper/versioninfo.json" "libviiper.versioninfo.tmp.json"
	goversioninfo -64 -o lib/viiper/resource.syso libviiper.versioninfo.tmp.json
	$env:CGO_ENABLED='1'; go build -buildmode=c-shared -trimpath {{ if type == "Release" { "-ldflags \"-s -w\"" } else { "" } }} -o dist/libVIIPER/libVIIPER.dll ./lib/viiper
	gendef - dist/libVIIPER/libVIIPER.dll | Set-Content -Encoding ascii dist/libVIIPER/libVIIPER.def
	go run ./lib/viiper/postbuild
	{{ rm_f }} libviiper.versioninfo.tmp.json
	just licenses-libVIIPER

[arg("type", pattern="Debug|Release")]
[unix]
build-libVIIPER type=build_type:
	{{ mkdir_p }} dist/libVIIPER
	CGO_ENABLED=1 go build -buildmode=c-shared -trimpath {{ if type == "Release" { "-ldflags \"-s -w\"" } else { "" } }} -o dist/libVIIPER/libVIIPER.so ./lib/viiper
	go run ./lib/viiper/postbuild
	just licenses-libVIIPER

clean: clean-versioninfo
	-{{ rm_rf }} {{ dist_dir }}
	-{{ rm_f }} coverage.out
	-{{ rm_f }} coverage.html
	go clean

fmt:
	go fmt ./...

lint:
	golangci-lint run ./...

[windows]
licenses:
	go install github.com/google/go-licenses/v2@latest  
	{{ mkdir_p }} {{ dist_dir }}; $template = (Get-Content {{ licenses_template }} -Raw).Replace('VERSION_PLACEHOLDER', '{{ version }}'); [System.IO.File]::WriteAllText("{{ licenses_template_work }}", $template, [System.Text.UTF8Encoding]::new($false)); {{ go_licenses_cmd }} report {{ main_pkg }} --ignore {{ licenses_ignore }} --template {{ licenses_template_work }} | Set-Content -Encoding utf8 {{ licenses_out }}; Remove-Item -Force {{ licenses_template_work }} -ErrorAction SilentlyContinue

[windows]
licenses-libVIIPER:
	go install github.com/google/go-licenses/v2@latest  
	{{ mkdir_p }} {{ licenses_dir }}; $template = (Get-Content {{ licenses_template }} -Raw).Replace('VERSION_PLACEHOLDER', '{{ version }}'); [System.IO.File]::WriteAllText("{{ licenses_template_work }}", $template, [System.Text.UTF8Encoding]::new($false)); {{ go_licenses_cmd }} report ./lib/viiper --ignore {{ licenses_ignore }} --template {{ licenses_template_work }} | Set-Content -Encoding utf8 {{ lib_licenses_out }}; Remove-Item -Force {{ licenses_template_work }} -ErrorAction SilentlyContinue

[unix]
licenses:
	go install github.com/google/go-licenses/v2@latest  
	{{ mkdir_p }} {{ dist_dir }} && sed "s/VERSION_PLACEHOLDER/{{ version }}/g" {{ licenses_template }} > {{ licenses_template_work }} && {{ go_licenses_cmd }} report {{ main_pkg }} --ignore {{ licenses_ignore }} --template {{ licenses_template_work }} > {{ licenses_out }} && rm -f {{ licenses_template_work }}

[unix]
licenses-libVIIPER:
	go install github.com/google/go-licenses/v2@latest  
	{{ mkdir_p }} {{ licenses_dir }} && sed "s/VERSION_PLACEHOLDER/{{ version }}/g" {{ licenses_template }} > {{ licenses_template_work }} && {{ go_licenses_cmd }} report ./lib/viiper --ignore {{ licenses_ignore }} --template {{ licenses_template_work }} > {{ lib_licenses_out }} && rm -f {{ licenses_template_work }}

run *args: build
	{{ if os_family() == "windows" { "$env:DEV='1'; & './" + build_path + "'" } else { "DEV=1 './" + build_path + "'" } }} {{ args }}

run-server *args: build
	{{ if os_family() == "windows" { "$env:DEV='1'; & './" + build_path + "' server" } else { "DEV=1 './" + build_path + "' server" } }} {{ args }}

version:
	@echo Version: {{ version }}
	@echo Commit:  {{ commit }}
	@echo Built:   {{ build_time }}
