// Command device-platform-admin performs controlled administrator operations
// against the device platform database.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	authrepository "github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/auth/repository"
	authservice "github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/auth/service"
	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/platform/database"
	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/platform/security"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(arguments []string) error {
	if len(arguments) == 0 || strings.TrimSpace(arguments[0]) != "bootstrap" {
		return errors.New("usage: device-platform-admin bootstrap --email <email> --display-name <name> (password from DEVICE_PLATFORM_ADMIN_PASSWORD)")
	}
	email, displayName, err := parseBootstrapArguments(arguments[1:])
	if err != nil {
		return err
	}
	password := strings.TrimSpace(os.Getenv("DEVICE_PLATFORM_ADMIN_PASSWORD"))
	if password == "" {
		return errors.New("DEVICE_PLATFORM_ADMIN_PASSWORD is required")
	}
	dsn := strings.TrimSpace(os.Getenv("DEVICE_PLATFORM_DATABASE_DSN"))
	if dsn == "" {
		return errors.New("DEVICE_PLATFORM_DATABASE_DSN is required")
	}
	mfaSecret := strings.TrimSpace(os.Getenv("DEVICE_PLATFORM_MFA_CREDENTIAL_KEY"))
	if mfaSecret == "" {
		return errors.New("DEVICE_PLATFORM_MFA_CREDENTIAL_KEY is required")
	}
	if len(mfaSecret) < 32 {
		return errors.New("DEVICE_PLATFORM_MFA_CREDENTIAL_KEY must contain at least 32 characters")
	}
	accessTokenSecret := strings.TrimSpace(
		os.Getenv("DEVICE_PLATFORM_AUTH_ACCESS_TOKEN_SECRET"),
	)
	if len(accessTokenSecret) < 32 {
		return errors.New(
			"DEVICE_PLATFORM_AUTH_ACCESS_TOKEN_SECRET must contain at least 32 characters",
		)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	store, err := database.Open(ctx, dsn)
	if err != nil {
		return err
	}
	defer store.Close(context.Background())

	mfaCipher, err := authservice.NewAESGCMTOTPCipher(mfaSecret)
	if err != nil {
		return err
	}
	tokenIssuer, err := security.NewHMACTokenIssuer(accessTokenSecret)
	if err != nil {
		return err
	}
	service, err := authservice.New(authservice.Options{
		Repository:      authrepository.NewPostgresRepository(store.Pool()),
		TokenIssuer:     tokenIssuer,
		MFACipher:       mfaCipher,
		AccessTTL:       15 * time.Minute,
		RefreshTTL:      30 * 24 * time.Hour,
		MFAChallengeTTL: 5 * time.Minute,
	})
	if err != nil {
		return err
	}
	uri, err := service.BootstrapAdmin(ctx, email, password, displayName)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(os.Stdout, "管理员已创建。请立即将下面的验证器密钥录入认证器应用：")
	_, _ = fmt.Fprintln(os.Stdout, uri)
	_, _ = fmt.Fprintln(os.Stdout, "完成后请清除设备平台管理员密码环境变量。")
	return nil
}

func parseBootstrapArguments(arguments []string) (string, string, error) {
	email := ""
	displayName := ""
	for index := 0; index < len(arguments); index++ {
		switch arguments[index] {
		case "--email":
			if index+1 >= len(arguments) {
				return "", "", errors.New("--email requires a value")
			}
			email = strings.TrimSpace(arguments[index+1])
			index++
		case "--display-name":
			if index+1 >= len(arguments) {
				return "", "", errors.New("--display-name requires a value")
			}
			displayName = strings.TrimSpace(arguments[index+1])
			index++
		default:
			return "", "", fmt.Errorf("unknown argument %q", arguments[index])
		}
	}
	if email == "" || displayName == "" {
		return "", "", errors.New("--email and --display-name are required")
	}
	return email, displayName, nil
}
