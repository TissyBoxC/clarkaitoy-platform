// Package domain contains device provisioning and binding types.
package domain

import (
	"errors"
	"time"
)

var (
	ErrTokenNotFound   = errors.New("binding token not found")
	ErrTokenExpired    = errors.New("binding token expired")
	ErrTokenConsumed   = errors.New("binding token already used")
	ErrDeviceNotFound  = errors.New("device not found")
	ErrDeviceBound     = errors.New("device already bound")
	ErrInvalidDeviceID = errors.New("invalid device id")
)

// BindingToken is a short-lived, single-use device provisioning token.
type BindingToken struct {
	ID         string
	DeviceID   string
	TokenHash  string
	ExpiresAt  time.Time
	ConsumedAt *time.Time
	CreatedAt  time.Time
}

// Binding is the durable relationship between a parent and a device.
type Binding struct {
	ID              string
	ParentAccountID string
	DeviceID        string
	DeviceName      string
	HardwareModel   string
	FirmwareVersion string
	Capabilities    []string
	BoundAt         time.Time
	UpdatedAt       time.Time
}
