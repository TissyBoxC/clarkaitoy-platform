package errorcode

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/TissyBoxC/sprout-platform/services/voice_gateway/internal/audio/codec"
	"github.com/TissyBoxC/sprout-platform/services/voice_gateway/internal/audio/frame"
)

func TestCatalogMatchesContractExample(t *testing.T) {
	examplePath := filepath.Join(
		"..", "..", "..", "..", "..",
		"packages", "contracts", "schemas", "audio_error.example.json",
	)
	content, err := os.ReadFile(examplePath)
	if err != nil {
		t.Fatalf("read contract example: %v", err)
	}

	var example struct {
		SchemaVersion  string `json:"schema_version"`
		CatalogVersion string `json:"catalog_version"`
		Codes          []struct {
			Code       string `json:"code"`
			Source     string `json:"source"`
			Retryable  bool   `json:"retryable"`
			Diagnostic string `json:"diagnostic"`
		} `json:"codes"`
	}
	if err := json.Unmarshal(content, &example); err != nil {
		t.Fatalf("parse contract example: %v", err)
	}
	if example.SchemaVersion != "1.0.0" {
		t.Fatalf("unexpected schema version: %s", example.SchemaVersion)
	}
	if example.CatalogVersion != CatalogVersion {
		t.Fatalf(
			"catalog version mismatch: contract=%s code=%s",
			example.CatalogVersion,
			CatalogVersion,
		)
	}

	details := Catalog()
	if len(details) != len(example.Codes) {
		t.Fatalf(
			"catalog size mismatch: contract=%d code=%d",
			len(example.Codes),
			len(details),
		)
	}
	for index, detail := range details {
		expected := example.Codes[index]
		if string(detail.Code) != expected.Code ||
			string(detail.Source) != expected.Source ||
			detail.Retryable != expected.Retryable ||
			detail.Diagnostic != expected.Diagnostic {
			t.Fatalf(
				"catalog entry %d mismatch: contract=%+v code=%+v",
				index,
				expected,
				detail,
			)
		}
	}
}

func TestDescribeRejectsUnknownCode(t *testing.T) {
	if _, ok := Describe("audio_codec_unknown"); ok {
		t.Fatal("expected an unknown code to be rejected")
	}
}

func TestFromCodecErrorMapsKnownFailures(t *testing.T) {
	cases := []struct {
		err  error
		want Code
	}{
		{err: codec.ErrInvalidPacketSize, want: CodePacketSizeInvalid},
		{err: codec.ErrInvalidPCMWindows, want: CodePCMWindowInvalid},
		{err: codec.ErrCodecUnavailable, want: CodeCodecUnavailable},
		{err: errors.New("opaque failure"), want: CodeCodecUnavailable},
	}
	for _, testCase := range cases {
		if got := FromCodecError(testCase.err); got != testCase.want {
			t.Fatalf("FromCodecError(%v) = %s, want %s", testCase.err, got, testCase.want)
		}
	}
}

func TestFromFrameErrorMapsContractViolations(t *testing.T) {
	cases := []struct {
		err  error
		want Code
	}{
		{err: frame.ErrPayloadTooLarge, want: CodePacketSizeInvalid},
		{err: frame.ErrInvalidPayload, want: CodePacketSizeInvalid},
		{err: frame.ErrInvalidProfile, want: CodePCMWindowInvalid},
		{err: frame.ErrMissingIdentifier, want: CodePacketSizeInvalid},
	}
	for _, testCase := range cases {
		if got := FromFrameError(testCase.err); got != testCase.want {
			t.Fatalf("FromFrameError(%v) = %s, want %s", testCase.err, got, testCase.want)
		}
	}
}

func TestIsRetryableTreatsUnknownCodeAsPermanent(t *testing.T) {
	if IsRetryable(CodeCodecUnavailable) != true {
		t.Fatal("expected codec unavailable to be retryable")
	}
	if IsRetryable(CodePacketSizeInvalid) != false {
		t.Fatal("expected invalid packet size to be permanent")
	}
	if IsRetryable("audio_codec_unknown") != false {
		t.Fatal("expected an unknown code to be permanent")
	}
}
