package framework

import "strings"

// InstallerType defines the type of installer for services
type InstallerType struct {
	Name string
}

const (
	cliInstallerKey = "cli"
	crInstallerKey  = "cr"
)

var (
	// CLIInstallerType defines the CLI installer
	CLIInstallerType InstallerType = InstallerType{Name: cliInstallerKey}
	// CRInstallerType defines the CR installer
	CRInstallerType InstallerType = InstallerType{Name: crInstallerKey}
)

// ParseInstallerType returns the correct installer type, based on the given string
func ParseInstallerType(typeStr string) InstallerType {
	switch t := strings.ToLower(typeStr); t {
	case cliInstallerKey:
		return CLIInstallerType
	case crInstallerKey:
		return CRInstallerType
	default:
		return InstallerType{Name: t}
	}
}
