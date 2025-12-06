package csharp

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Alia5/VIIPER/internal/codegen/meta"
)

const projectTemplate = `<Project Sdk="Microsoft.NET.Sdk">

  <PropertyGroup>
    <TargetFramework>net8.0</TargetFramework>
    <RootNamespace>Viiper.Client</RootNamespace>
    <Nullable>enable</Nullable>
    <LangVersion>latest</LangVersion>
    <ImplicitUsings>enable</ImplicitUsings>
    
    <PackageId>Viiper.Client</PackageId>
    <Version>{{.Version}}</Version>
    <Authors>Peter Repukat</Authors>
    <Description>VIIPER Client Library for C#</Description>
    <PackageLicenseExpression>MIT</PackageLicenseExpression>
    <PackageProjectUrl>https://github.com/Alia5/VIIPER</PackageProjectUrl>
    <RepositoryUrl>https://github.com/Alia5/VIIPER</RepositoryUrl>
    <RepositoryType>git</RepositoryType>
    <PackageTags>viiper;usbip;virtual-device;input-emulation;hid</PackageTags>
    <PackageReadmeFile>README.md</PackageReadmeFile>
  </PropertyGroup>

  <ItemGroup>
    <None Include="../README.md" Pack="true" PackagePath="/"/>
  </ItemGroup>

</Project>
`

func generateProject(logger *slog.Logger, projectDir string, md *meta.Metadata, version string) error {
	logger.Debug("Generating Viiper.Client.csproj")

	tmpl, err := template.New("csproj").Parse(projectTemplate)
	if err != nil {
		return fmt.Errorf("parse csproj template: %w", err)
	}

	f, err := os.Create(filepath.Join(projectDir, "Viiper.Client.csproj"))
	if err != nil {
		return fmt.Errorf("create csproj file: %w", err)
	}
	defer f.Close()

	data := struct {
		Version string
	}{
		Version: version,
	}

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("execute csproj template: %w", err)
	}

	logger.Info("Generated Viiper.Client.csproj", "version", version)
	return nil
}
