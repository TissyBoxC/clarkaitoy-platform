// Package service owns device provisioning token and binding workflows.
package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/device_binding/domain"
	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/device_binding/repository"
	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/platform/clock"
	"github.com/google/uuid"
)

var deviceIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{7,127}$`)

// Service creates, consumes, and revokes device provisioning state.
type Service struct {
	repository repository.Repository
	timeSource clock.Clock
	tokenTTL   time.Duration
}

// Options contains device binding service dependencies and policy.
type Options struct {
	Repository repository.Repository
	Clock      clock.Clock
	TokenTTL   time.Duration
}

// New creates the device binding service.
func New(options Options) (*Service, error) {
	if options.Repository == nil {
		return nil, errors.New("device binding repository is required")
	}
	if options.TokenTTL <= 0 {
		return nil, errors.New("device binding token TTL must be positive")
	}
	timeSource := options.Clock
	if timeSource == nil {
		timeSource = clock.SystemClock{}
	}
	return &Service{
		repository: options.Repository,
		timeSource: timeSource,
		tokenTTL:   options.TokenTTL,
	}, nil
}

// CreateToken creates a short-lived, single-use token for one device.
// Only the token hash is persisted; the plaintext is returned once.
func (s *Service) CreateToken(
	ctx context.Context,
	deviceID string,
) (string, *domain.BindingToken, error) {
	deviceID = strings.TrimSpace(deviceID)
	if !deviceIDPattern.MatchString(deviceID) {
		return "", nil, domain.ErrInvalidDeviceID
	}
	plainToken, err := randomToken()
	if err != nil {
		return "", nil, err
	}
	now := s.timeSource.Now().UTC()
	token := &domain.BindingToken{
		ID:        uuid.NewString(),
		DeviceID:  deviceID,
		TokenHash: hashToken(plainToken),
		ExpiresAt: now.Add(s.tokenTTL),
		CreatedAt: now,
	}
	if err := s.repository.CreateToken(ctx, token); err != nil {
		return "", nil, err
	}
	return plainToken, token, nil
}

// Bind consumes a provisioning token and binds the device to the parent.
func (s *Service) Bind(
	ctx context.Context,
	parentAccountID string,
	plainToken string,
	deviceName string,
	hardwareModel string,
	firmwareVersion string,
	capabilities []string,
) (*domain.Binding, error) {
	plainToken = strings.TrimSpace(plainToken)
	if plainToken == "" {
		return nil, domain.ErrTokenNotFound
	}
	token, err := s.repository.GetTokenByHash(ctx, hashToken(plainToken))
	if err != nil {
		return nil, err
	}
	now := s.timeSource.Now().UTC()
	if token.ConsumedAt != nil {
		return nil, domain.ErrTokenConsumed
	}
	if !token.ExpiresAt.After(now) {
		return nil, domain.ErrTokenExpired
	}
	deviceName = strings.TrimSpace(deviceName)
	hardwareModel = strings.TrimSpace(hardwareModel)
	firmwareVersion = strings.TrimSpace(firmwareVersion)
	if deviceName == "" {
		deviceName = "初芽"
	}
	if len([]rune(deviceName)) > 40 || len([]rune(hardwareModel)) > 64 ||
		len([]rune(firmwareVersion)) > 64 {
		return nil, domain.ErrInvalidDeviceID
	}

	if err := s.repository.ConsumeToken(ctx, token.ID); err != nil {
		return nil, err
	}
	binding := &domain.Binding{
		ID:              uuid.NewString(),
		ParentAccountID: parentAccountID,
		DeviceID:        token.DeviceID,
		DeviceName:      deviceName,
		HardwareModel:   hardwareModel,
		FirmwareVersion: firmwareVersion,
		Capabilities:    normalizeCapabilities(capabilities),
		BoundAt:         now,
		UpdatedAt:       now,
	}
	if err := s.repository.UpsertBinding(ctx, binding); err != nil {
		return nil, err
	}
	return binding, nil
}

// List returns the devices owned by one parent account.
func (s *Service) List(
	ctx context.Context,
	parentAccountID string,
) ([]domain.Binding, error) {
	return s.repository.ListByParentAccountID(ctx, parentAccountID)
}

// Delete removes a device binding owned by the parent account.
func (s *Service) Delete(
	ctx context.Context,
	parentAccountID string,
	deviceID string,
) error {
	deviceID = strings.TrimSpace(deviceID)
	if !deviceIDPattern.MatchString(deviceID) {
		return domain.ErrInvalidDeviceID
	}
	return s.repository.Delete(ctx, parentAccountID, deviceID)
}

func normalizeCapabilities(capabilities []string) []string {
	normalized := make([]string, 0, len(capabilities))
	seen := make(map[string]struct{}, len(capabilities))
	for _, capability := range capabilities {
		capability = strings.TrimSpace(capability)
		if capability == "" || len(capability) > 64 {
			continue
		}
		if _, exists := seen[capability]; exists {
			continue
		}
		seen[capability] = struct{}{}
		normalized = append(normalized, capability)
	}
	return normalized
}

func randomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func hashToken(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}
