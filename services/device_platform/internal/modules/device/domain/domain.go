// Package domain contains device domain types.
package domain

// Capability describes an optional hardware capability.
type Capability string

const (
	CapabilityCamera   Capability = "camera"
	CapabilityDisplay  Capability = "display"
	CapabilityBattery  Capability = "battery"
	CapabilityCellular Capability = "cellular_4g"
	CapabilityMotion   Capability = "motion"
)

// Device describes a registered game console.
type Device struct {
	ID           string
	FamilyID     string
	Capabilities []Capability
}
