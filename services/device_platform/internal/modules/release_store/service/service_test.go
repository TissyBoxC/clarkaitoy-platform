package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/release_store/domain"
)

type fixedClock struct {
	now time.Time
}

func (clock fixedClock) Now() time.Time {
	return clock.now
}

func newTestService(t *testing.T) (*Service, string) {
	t.Helper()
	root := t.TempDir()
	service, err := New(Options{
		RootDir:       root,
		PublicBaseURL: "https://download.example.test",
		Clock: fixedClock{
			now: time.Date(2026, 10, 4, 8, 0, 0, 0, time.UTC),
		},
	})
	if err != nil {
		t.Fatalf("create release store service: %v", err)
	}
	t.Cleanup(func() {
		_ = service.Close()
	})
	return service, root
}

func TestInventoryListsNestedFilesAndAnnotatesReleaseArtifacts(t *testing.T) {
	service, root := newTestService(t)
	artifactPath := filepath.Join(
		root,
		"0.12.3",
		"stable",
		"android",
		"apk",
		"sprout-parent-app-v0.12.3.apk",
	)
	if err := os.MkdirAll(filepath.Dir(artifactPath), 0o755); err != nil {
		t.Fatalf("create artifact directory: %v", err)
	}
	if err := os.WriteFile(artifactPath, []byte("apk-bytes"), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "manual-image.png"), []byte("png"), 0o644); err != nil {
		t.Fatalf("write manual file: %v", err)
	}

	inventory, err := service.Inventory(context.Background())
	if err != nil {
		t.Fatalf("Inventory() returned unexpected error: %v", err)
	}
	if len(inventory.Files) != 2 {
		t.Fatalf("expected two files, got %d", len(inventory.Files))
	}
	var releaseFile *domain.File
	for index := range inventory.Files {
		if inventory.Files[index].Release != nil {
			releaseFile = &inventory.Files[index]
		}
	}
	if releaseFile == nil {
		t.Fatal("expected one canonical release artifact")
	}
	if releaseFile.Release.Version != "0.12.3" ||
		releaseFile.Release.Channel != "stable" ||
		releaseFile.Release.Platform != "android" ||
		releaseFile.Release.Kind != "apk" {
		t.Fatalf("unexpected release metadata: %+v", releaseFile.Release)
	}
	if releaseFile.SHA256 != "93e5c2dd0e6b0e5f2c9e6e8c17fd22b7d62075a6e2e81e6b0f8a9c6d5e4f3a2b" &&
		len(releaseFile.SHA256) != 64 {
		t.Fatalf("expected sha256 digest, got %q", releaseFile.SHA256)
	}
	if releaseFile.DownloadURL !=
		"https://download.example.test/0.12.3/stable/android/apk/sprout-parent-app-v0.12.3.apk" {
		t.Fatalf("unexpected download URL: %q", releaseFile.DownloadURL)
	}
}

func TestUploadFileWritesAtomicallyAndRejectsDuplicate(t *testing.T) {
	service, root := newTestService(t)
	uploaded, err := service.UploadFile(
		context.Background(),
		domain.UploadInput{
			RelativePath: "0.13.0/stable/any/notes/readme.txt",
		},
		strings.NewReader("release notes"),
	)
	if err != nil {
		t.Fatalf("UploadFile() returned unexpected error: %v", err)
	}
	if uploaded.RelativePath != "0.13.0/stable/any/notes/readme.txt" {
		t.Fatalf("unexpected relative path: %q", uploaded.RelativePath)
	}
	payload, err := os.ReadFile(filepath.Join(
		root,
		"0.13.0",
		"stable",
		"any",
		"notes",
		"readme.txt",
	))
	if err != nil {
		t.Fatalf("read uploaded file: %v", err)
	}
	if string(payload) != "release notes" {
		t.Fatalf("unexpected uploaded payload: %q", payload)
	}
	_, err = service.UploadFile(
		context.Background(),
		domain.UploadInput{
			RelativePath: "0.13.0/stable/any/notes/readme.txt",
		},
		strings.NewReader("duplicate"),
	)
	if !errors.Is(err, domain.ErrFileAlreadyExists) {
		t.Fatalf("expected duplicate upload rejection, got %v", err)
	}
}

func TestUploadFileRejectsPathTraversal(t *testing.T) {
	service, _ := newTestService(t)
	for _, relativePath := range []string{
		"../outside.txt",
		"/absolute.txt",
		"release/../../outside.txt",
		`release\..\outside.txt`,
		"",
	} {
		_, err := service.UploadFile(
			context.Background(),
			domain.UploadInput{RelativePath: relativePath},
			strings.NewReader("blocked"),
		)
		if !errors.Is(err, domain.ErrPathInvalid) {
			t.Fatalf("expected %q to be rejected, got %v", relativePath, err)
		}
	}
}

func TestUploadFileRejectsOversizedStream(t *testing.T) {
	service, _ := newTestService(t)
	reader := bytes.NewReader(make([]byte, int(domain.MaxUploadBytes+1)))
	_, err := service.UploadFile(
		context.Background(),
		domain.UploadInput{RelativePath: "large.bin"},
		reader,
	)
	if !errors.Is(err, domain.ErrFileTooLarge) {
		t.Fatalf("expected oversized upload rejection, got %v", err)
	}
}

func TestDeleteFileRemovesFileAndEmptyParents(t *testing.T) {
	service, root := newTestService(t)
	relativePath := "0.13.0/stable/any/notes/readme.txt"
	if _, err := service.UploadFile(
		context.Background(),
		domain.UploadInput{RelativePath: relativePath},
		strings.NewReader("release notes"),
	); err != nil {
		t.Fatalf("seed upload: %v", err)
	}
	if err := service.DeleteFile(context.Background(), relativePath); err != nil {
		t.Fatalf("DeleteFile() returned unexpected error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "0.13.0")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected empty release directories to be removed, got %v", err)
	}
	if err := service.DeleteFile(context.Background(), relativePath); !errors.Is(
		err,
		domain.ErrFileNotFound,
	) {
		t.Fatalf("expected missing file error, got %v", err)
	}
}

func TestRefreshIndexMatchesPublicReleaseContract(t *testing.T) {
	service, root := newTestService(t)
	relativePath := "0.12.3/stable/android/apk/sprout-parent-app-v0.12.3.apk"
	if _, err := service.UploadFile(
		context.Background(),
		domain.UploadInput{RelativePath: relativePath},
		strings.NewReader("apk"),
	); err != nil {
		t.Fatalf("seed upload: %v", err)
	}
	result, err := service.RefreshIndex(context.Background(), domain.IndexRefreshRequest{
		Version:     "0.12.3",
		Channel:     "stable",
		Title:       "如此萌屋平台 0.12.3",
		PublishedAt: "2026-10-04T08:00:00Z",
	})
	if err != nil {
		t.Fatalf("RefreshIndex() returned unexpected error: %v", err)
	}
	if result.ArtifactCount != 1 {
		t.Fatalf("expected one artifact, got %d", result.ArtifactCount)
	}
	manifestPayload, err := os.ReadFile(filepath.Join(
		root,
		"0.12.3",
		"stable",
		"manifest.json",
	))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var manifest domain.Manifest
	if err := json.Unmarshal(manifestPayload, &manifest); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	if manifest.Version != "0.12.3" ||
		manifest.Channel != "stable" ||
		len(manifest.Artifacts) != 1 {
		t.Fatalf("unexpected manifest: %+v", manifest)
	}
	artifact := manifest.Artifacts[0]
	if artifact.FileName != "sprout-parent-app-v0.12.3.apk" ||
		artifact.Platform != "android" ||
		artifact.Kind != "apk" ||
		artifact.RelativePath != "android/apk/sprout-parent-app-v0.12.3.apk" ||
		artifact.SizeBytes != 3 ||
		artifact.URL != "https://download.example.test/"+relativePath {
		t.Fatalf("unexpected manifest artifact: %+v", artifact)
	}
	indexPayload, err := os.ReadFile(filepath.Join(root, "index.json"))
	if err != nil {
		t.Fatalf("read index: %v", err)
	}
	var index domain.ReleaseIndex
	if err := json.Unmarshal(indexPayload, &index); err != nil {
		t.Fatalf("decode index: %v", err)
	}
	if index.SchemaVersion != 1 || len(index.Releases) != 1 {
		t.Fatalf("unexpected index: %+v", index)
	}
	entry := index.Releases[0]
	if entry.Version != "0.12.3" ||
		entry.Channel != "stable" ||
		entry.ArtifactCount != 1 ||
		entry.ManifestURL != "https://download.example.test/0.12.3/stable/manifest.json" {
		t.Fatalf("unexpected index entry: %+v", entry)
	}
}

func TestRefreshIndexRejectsInvalidReleaseDirectory(t *testing.T) {
	service, _ := newTestService(t)
	_, err := service.RefreshIndex(context.Background(), domain.IndexRefreshRequest{
		Version: "0.12.3",
		Channel: "stable",
	})
	if !errors.Is(err, domain.ErrReleaseNotCreated) {
		t.Fatalf("expected missing release directory error, got %v", err)
	}
}

func TestParseArtifactPathRejectsNonCanonicalPaths(t *testing.T) {
	for _, value := range []string{
		"0.12.3/stable/android/apk",
		"0.12.3/stable/android/apk/extra/file.apk",
		"v0.12.3/stable/android/apk/file.apk",
		"0.12.3/stable/android/apk/../file.apk",
		"0.12.3/stable/android/apk/file apk",
	} {
		if release, ok := ParseArtifactPath(value); ok {
			t.Fatalf("expected %q to be rejected, got %+v", value, release)
		}
	}
}
