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

	authdomain "github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/auth/domain"
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
	if len(arguments) == 0 {
		return errors.New(
			"usage: device-platform-admin <bootstrap|reset-password|create-parent>",
		)
	}
	command := strings.TrimSpace(arguments[0])
	if command != "bootstrap" &&
		command != "reset-password" &&
		command != "create-parent" {
		return errors.New(
			"usage: device-platform-admin <bootstrap|reset-password|create-parent>",
		)
	}

	accountArguments, err := parseAdminArguments(command, arguments[1:])
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
	if command == "reset-password" {
		if err := service.ResetAdminPassword(ctx, accountArguments.email, password); err != nil {
			return err
		}
		_, _ = fmt.Fprintln(os.Stdout, "管理员密码已重置，原有登录会话已失效。")
		return nil
	}

	if command == "create-parent" {
		account, summary, err := service.CreateParent(ctx, authdomain.RegisterInput{
			Phone:                  accountArguments.phone,
			Password:               password,
			GuardianFamilyName:     accountArguments.guardianFamilyName,
			ChildNickname:          accountArguments.childNickname,
			ChildBirthday:          accountArguments.childBirthday,
			GuardianConsentVersion: "2026-01",
		})
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(
			os.Stdout,
			"家长账号已创建：%s（%s）\n",
			account.DisplayName,
			account.Phone,
		)
		if summary == nil {
			_, _ = fmt.Fprintln(os.Stdout, "AI 账户暂未开通，登录后会自动重试。")
		} else {
			_, _ = fmt.Fprintf(os.Stdout, "AI 账户状态：%s\n", summary.Status)
		}
		return nil
	}

	uri, err := service.BootstrapAdmin(
		ctx,
		accountArguments.email,
		password,
		accountArguments.displayName,
	)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(os.Stdout, "管理员已创建。请立即将下面的验证器密钥录入认证器应用：")
	_, _ = fmt.Fprintln(os.Stdout, uri)
	_, _ = fmt.Fprintln(os.Stdout, "完成后请清除设备平台管理员密码环境变量。")
	return nil
}

type adminArguments struct {
	email              string
	displayName        string
	phone              string
	guardianFamilyName string
	childNickname      string
	childBirthday      string
}

func parseAdminArguments(
	command string,
	arguments []string,
) (adminArguments, error) {
	var parsed adminArguments
	email := ""
	displayName := ""
	for index := 0; index < len(arguments); index++ {
		switch arguments[index] {
		case "--email":
			if index+1 >= len(arguments) {
				return adminArguments{}, errors.New("--email requires a value")
			}
			email = strings.TrimSpace(arguments[index+1])
			index++
		case "--display-name":
			if command != "bootstrap" {
				return adminArguments{}, errors.New(
					"--display-name is only valid for bootstrap",
				)
			}
			if index+1 >= len(arguments) {
				return adminArguments{}, errors.New("--display-name requires a value")
			}
			displayName = strings.TrimSpace(arguments[index+1])
			index++
		case "--phone":
			if command != "create-parent" {
				return adminArguments{}, errors.New(
					"--phone is only valid for create-parent",
				)
			}
			if index+1 >= len(arguments) {
				return adminArguments{}, errors.New("--phone requires a value")
			}
			parsed.phone = strings.TrimSpace(arguments[index+1])
			index++
		case "--guardian-family-name":
			if command != "create-parent" {
				return adminArguments{}, errors.New(
					"--guardian-family-name is only valid for create-parent",
				)
			}
			if index+1 >= len(arguments) {
				return adminArguments{}, errors.New(
					"--guardian-family-name requires a value",
				)
			}
			parsed.guardianFamilyName = strings.TrimSpace(arguments[index+1])
			index++
		case "--child-nickname":
			if command != "create-parent" {
				return adminArguments{}, errors.New(
					"--child-nickname is only valid for create-parent",
				)
			}
			if index+1 >= len(arguments) {
				return adminArguments{}, errors.New(
					"--child-nickname requires a value",
				)
			}
			parsed.childNickname = strings.TrimSpace(arguments[index+1])
			index++
		case "--child-birthday":
			if command != "create-parent" {
				return adminArguments{}, errors.New(
					"--child-birthday is only valid for create-parent",
				)
			}
			if index+1 >= len(arguments) {
				return adminArguments{}, errors.New(
					"--child-birthday requires a value",
				)
			}
			parsed.childBirthday = strings.TrimSpace(arguments[index+1])
			index++
		default:
			return adminArguments{}, fmt.Errorf(
				"unknown argument %q",
				arguments[index],
			)
		}
	}
	if command == "create-parent" {
		if parsed.phone == "" {
			return adminArguments{}, errors.New("--phone is required")
		}
		return parsed, nil
	}
	if email == "" {
		return adminArguments{}, errors.New("--email is required")
	}
	if command == "bootstrap" && displayName == "" {
		return adminArguments{}, errors.New("--display-name is required for bootstrap")
	}
	parsed.email = email
	parsed.displayName = displayName
	return parsed, nil
}
