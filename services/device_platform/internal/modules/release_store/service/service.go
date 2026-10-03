// Package service owns safe reads and writes for the shared download volume.
//
// Files are addressed by a slash-separated relative path. Every path is
// validated lexically and then resolved through os.Root, so symlinks and
// traversal cannot escape the configured root. Writes use a sibling temporary
// file followed by rename, which prevents readers from seeing partial files.
package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/release_store/domain"
	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/platform/clock"
)

var (
	versionPattern  = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)
	segmentPattern  = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._+-]*$`)
	channelPattern  = regexp.MustCompile(`^[a-z][a-z0-9._-]*$`)
	platformPattern = regexp.MustCompile(`^[a-z][a-z0-9._-]*$`)
	kindPattern     = regexp.MustCompile(`^[a-z][a-z0-9._-]*$`)
)

// Service owns the shared download volume.
type Service struct {
	root          *os.Root
	rootDir       string
	publicBaseURL string
	clock         clock.Clock
}

type releaseDirectoryTarget struct {
	version string
	channel string
}

// Options contains download-store dependencies.
type Options struct {
	RootDir       string
	PublicBaseURL string
	Clock         clock.Clock
}

// New opens the download store root. The directory must already exist so a
// typo in configuration fails startup instead of creating a new empty tree.
func New(options Options) (*Service, error) {
	rootDir := strings.TrimSpace(options.RootDir)
	if rootDir == "" {
		return nil, errors.New("download store root directory is required")
	}
	absoluteRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("resolve download store root: %w", err)
	}
	info, err := os.Stat(absoluteRoot)
	if err != nil {
		return nil, fmt.Errorf("stat download store root: %w", err)
	}
	if !info.IsDir() {
		return nil, errors.New("download store root is not a directory")
	}
	root, err := os.OpenRoot(absoluteRoot)
	if err != nil {
		return nil, fmt.Errorf("open download store root: %w", err)
	}
	timeSource := options.Clock
	if timeSource == nil {
		timeSource = clock.SystemClock{}
	}
	return &Service{
		root:          root,
		rootDir:       absoluteRoot,
		publicBaseURL: strings.TrimRight(strings.TrimSpace(options.PublicBaseURL), "/"),
		clock:         timeSource,
	}, nil
}

// Close releases the root handle.
func (s *Service) Close() error {
	if s == nil || s.root == nil {
		return nil
	}
	return s.root.Close()
}

// Inventory returns every regular file in the store. Canonical release
// artifacts are annotated with their parsed release metadata.
func (s *Service) Inventory(ctx context.Context) (domain.Inventory, error) {
	if err := ctx.Err(); err != nil {
		return domain.Inventory{}, err
	}
	indexedArtifacts := s.indexedArtifactSet()
	files := make([]domain.File, 0)
	err := walkRoot(s.root, ".", func(relativePath string, info os.FileInfo) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		digest, err := s.fileSHA256(relativePath)
		if err != nil {
			return err
		}
		release, _ := ParseArtifactPath(relativePath)
		files = append(files, domain.File{
			RelativePath: relativePath,
			FileName:     path.Base(relativePath),
			Name:         path.Base(relativePath),
			Directory:    directoryName(relativePath),
			SizeBytes:    info.Size(),
			ModifiedAt:   info.ModTime().UTC(),
			SHA256:       digest,
			Release:      release,
			DownloadURL:  s.publicURL(relativePath),
			IsIndexed:    isArtifactIndexed(relativePath, indexedArtifacts),
		})
		return nil
	})
	if err != nil {
		return domain.Inventory{}, err
	}
	sort.SliceStable(files, func(left int, right int) bool {
		return files[left].RelativePath < files[right].RelativePath
	})
	return domain.Inventory{
		RootDir:       s.rootDir,
		PublicBaseURL: s.publicBaseURL,
		Files:         files,
		GeneratedAt:   s.clock.Now().UTC(),
	}, nil
}

// IndexStatus reports how much of the current release inventory is represented
// by the public manifest and index. A missing index remains a valid first
// deployment state, so it is reported as unavailable rather than failing the
// whole download panel.
func (s *Service) IndexStatus(ctx context.Context) (domain.IndexStatus, error) {
	if err := ctx.Err(); err != nil {
		return domain.IndexStatus{}, err
	}
	inventory, err := s.Inventory(ctx)
	if err != nil {
		return domain.IndexStatus{}, err
	}
	index, err := s.loadIndex()
	if errors.Is(err, os.ErrNotExist) {
		return domain.IndexStatus{
			IsAvailable:      false,
			PendingFileCount: countCanonicalArtifacts(inventory.Files),
			ErrorMessage:     "",
		}, nil
	}
	if err != nil {
		return domain.IndexStatus{
			IsAvailable:      false,
			PendingFileCount: countCanonicalArtifacts(inventory.Files),
			ErrorMessage:     "download index unavailable",
		}, nil
	}
	indexedArtifacts := s.indexedArtifactSet()
	pending := 0
	for _, file := range inventory.Files {
		if file.Release == nil {
			continue
		}
		if _, exists := indexedArtifacts[file.RelativePath]; !exists {
			pending++
		}
	}
	return domain.IndexStatus{
		IsAvailable:      true,
		RefreshedAt:      index.GeneratedAt,
		IndexedFileCount: len(indexedArtifacts),
		PendingFileCount: pending,
	}, nil
}

// UploadFile stores one manually uploaded file.
//
// The caller passes a single reader so the handler can enforce the HTTP body
// limit before the service performs the atomic rename.
func (s *Service) UploadFile(
	ctx context.Context,
	input domain.UploadInput,
	reader io.Reader,
) (domain.File, error) {
	if err := ctx.Err(); err != nil {
		return domain.File{}, err
	}
	relativePath, err := NormalizeFilePath(input.RelativePath)
	if err != nil {
		return domain.File{}, err
	}
	parent := path.Dir(relativePath)
	if parent != "." {
		if err := s.root.MkdirAll(parent, 0o755); err != nil {
			return domain.File{}, fmt.Errorf("create download directory: %w", err)
		}
	}
	if !input.Overwrite {
		if _, err := s.root.Stat(relativePath); err == nil {
			return domain.File{}, domain.ErrFileAlreadyExists
		} else if !errors.Is(err, os.ErrNotExist) {
			return domain.File{}, err
		}
	}
	temporaryName, temporary, err := s.createTemporary(parent)
	if err != nil {
		return domain.File{}, err
	}
	temporaryPath := path.Join(parent, temporaryName)
	completed := false
	defer func() {
		if !completed {
			_ = s.root.Remove(temporaryPath)
		}
	}()
	digest := sha256.New()
	written, err := io.Copy(io.MultiWriter(temporary, digest), io.LimitReader(reader, domain.MaxUploadBytes+1))
	if err != nil {
		_ = temporary.Close()
		return domain.File{}, fmt.Errorf("write download file: %w", err)
	}
	if written > domain.MaxUploadBytes {
		_ = temporary.Close()
		return domain.File{}, domain.ErrFileTooLarge
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return domain.File{}, fmt.Errorf("sync download file: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return domain.File{}, fmt.Errorf("close download file: %w", err)
	}
	if err := s.root.Rename(temporaryPath, relativePath); err != nil {
		return domain.File{}, fmt.Errorf("publish download file: %w", err)
	}
	completed = true
	info, err := s.root.Stat(relativePath)
	if err != nil {
		return domain.File{}, fmt.Errorf("stat uploaded download file: %w", err)
	}
	release, _ := ParseArtifactPath(relativePath)
	return domain.File{
		RelativePath: relativePath,
		FileName:     path.Base(relativePath),
		SizeBytes:    info.Size(),
		ModifiedAt:   info.ModTime().UTC(),
		SHA256:       hex.EncodeToString(digest.Sum(nil)),
		Release:      release,
		DownloadURL:  s.publicURL(relativePath),
	}, nil
}

// DeleteFile removes one file and prunes now-empty parent directories.
func (s *Service) DeleteFile(ctx context.Context, relativePath string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	normalized, err := NormalizeFilePath(relativePath)
	if err != nil {
		return err
	}
	info, err := s.root.Stat(normalized)
	if errors.Is(err, os.ErrNotExist) {
		return domain.ErrFileNotFound
	}
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return domain.ErrPathInvalid
	}
	if err := s.root.Remove(normalized); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return domain.ErrFileNotFound
		}
		return fmt.Errorf("delete download file: %w", err)
	}
	// Removing empty parents keeps the console inventory free of empty
	// directory noise. Root itself is never removed.
	parent := path.Dir(normalized)
	for parent != "." {
		if err := s.root.Remove(parent); err != nil {
			break
		}
		parent = path.Dir(parent)
	}
	return nil
}

// RefreshIndex regenerates one release manifest and the global index.
func (s *Service) RefreshIndex(
	ctx context.Context,
	input domain.IndexRefreshRequest,
) (domain.IndexRefreshResult, error) {
	if err := ctx.Err(); err != nil {
		return domain.IndexRefreshResult{}, err
	}
	version := strings.TrimSpace(input.Version)
	channel := strings.TrimSpace(input.Channel)
	if channel == "" {
		channel = domain.DefaultChannel
	}
	if !versionPattern.MatchString(version) {
		return domain.IndexRefreshResult{}, domain.ErrReleaseInvalid
	}
	if !channelPattern.MatchString(channel) {
		return domain.IndexRefreshResult{}, domain.ErrReleaseInvalid
	}
	releaseDirectory := path.Join(version, channel)
	if info, err := s.root.Stat(releaseDirectory); err != nil ||
		!info.IsDir() {
		return domain.IndexRefreshResult{}, domain.ErrReleaseNotCreated
	}
	manifest, err := s.collectManifest(releaseDirectory)
	if err != nil {
		return domain.IndexRefreshResult{}, err
	}
	manifest.Version = version
	manifest.Channel = channel
	if len(manifest.Artifacts) == 0 {
		return domain.IndexRefreshResult{}, domain.ErrReleaseInvalid
	}
	if err := s.writeJSONAtomic(path.Join(releaseDirectory, "manifest.json"), manifest); err != nil {
		return domain.IndexRefreshResult{}, err
	}
	index, err := s.loadIndex()
	if err != nil {
		return domain.IndexRefreshResult{}, err
	}
	generatedAt := s.clock.Now().UTC()
	publishedAt := strings.TrimSpace(input.PublishedAt)
	if publishedAt == "" {
		publishedAt = generatedAt.Format(time.RFC3339)
	}
	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = "如此萌屋平台 " + version
	}
	entry := domain.IndexRelease{
		ArtifactCount: len(manifest.Artifacts),
		Channel:       channel,
		ManifestURL:   s.publicURL(path.Join(releaseDirectory, "manifest.json")),
		PublishedAt:   publishedAt,
		Title:         title,
		Version:       version,
	}
	index.GeneratedAt = generatedAt
	index.SchemaVersion = 1
	index.Releases = replaceIndexRelease(index.Releases, entry)
	if err := s.writeJSONAtomic("index.json", index); err != nil {
		return domain.IndexRefreshResult{}, err
	}
	return domain.IndexRefreshResult{
		Version:       version,
		Channel:       channel,
		ArtifactCount: len(manifest.Artifacts),
		ManifestURL:   entry.ManifestURL,
		GeneratedAt:   generatedAt,
	}, nil
}

// RefreshAllIndexes scans every canonical release directory, regenerates each
// manifest, and atomically replaces the global index. This is the operation
// used by the console's "refresh index" action and by release synchronization.
func (s *Service) RefreshAllIndexes(
	ctx context.Context,
) (domain.IndexRefreshSummary, error) {
	if err := ctx.Err(); err != nil {
		return domain.IndexRefreshSummary{}, err
	}
	targets := make([]releaseDirectoryTarget, 0)
	err := walkRoot(s.root, ".", func(relativePath string, info os.FileInfo) error {
		if !info.Mode().IsRegular() || path.Base(relativePath) != "manifest.json" {
			return nil
		}
		segments := strings.Split(relativePath, "/")
		if len(segments) != 3 {
			return nil
		}
		version, channel := segments[0], segments[1]
		if !versionPattern.MatchString(version) ||
			!channelPattern.MatchString(channel) {
			return nil
		}
		targets = append(targets, releaseDirectoryTarget{version: version, channel: channel})
		return nil
	})
	if err != nil {
		return domain.IndexRefreshSummary{}, err
	}
	// A release directory may contain artifacts before its first manifest is
	// generated. Discover those directories too so manual uploads are included
	// without requiring an extra operator step.
	directoryTargets, err := s.releaseDirectories()
	if err != nil {
		return domain.IndexRefreshSummary{}, err
	}
	targets = mergeReleaseTargets(targets, directoryTargets)
	if len(targets) == 0 {
		return domain.IndexRefreshSummary{}, domain.ErrReleaseNotCreated
	}
	sort.SliceStable(targets, func(left int, right int) bool {
		if targets[left].version == targets[right].version {
			return targets[left].channel < targets[right].channel
		}
		return targets[left].version < targets[right].version
	})
	summary := domain.IndexRefreshSummary{
		Versions: make([]string, 0),
		Channels: make([]string, 0),
	}
	for _, target := range targets {
		result, err := s.RefreshIndex(ctx, domain.IndexRefreshRequest{
			Version: target.version,
			Channel: target.channel,
		})
		if errors.Is(err, domain.ErrReleaseInvalid) {
			continue
		}
		if err != nil {
			return domain.IndexRefreshSummary{}, err
		}
		summary.ReleaseCount++
		summary.ArtifactCount += result.ArtifactCount
		summary.Versions = append(summary.Versions, result.Version)
		summary.Channels = append(summary.Channels, result.Channel)
	}
	if summary.ReleaseCount == 0 {
		return domain.IndexRefreshSummary{}, domain.ErrReleaseNotCreated
	}
	status, err := s.IndexStatus(ctx)
	if err != nil {
		return domain.IndexRefreshSummary{}, err
	}
	summary.RefreshedAt = status.RefreshedAt
	summary.IndexedFileCount = status.IndexedFileCount
	summary.PendingFileCount = status.PendingFileCount
	return summary, nil
}

func (s *Service) releaseDirectories() ([]releaseDirectoryTarget, error) {
	targets := make([]releaseDirectoryTarget, 0)
	entries, err := s.readDirectory(".")
	if err != nil {
		return nil, err
	}
	for _, versionEntry := range entries {
		if !versionEntry.IsDir() || !versionPattern.MatchString(versionEntry.Name()) {
			continue
		}
		channelEntries, err := s.readDirectory(versionEntry.Name())
		if err != nil {
			return nil, err
		}
		for _, channelEntry := range channelEntries {
			if !channelEntry.IsDir() || !channelPattern.MatchString(channelEntry.Name()) {
				continue
			}
			targets = append(targets, releaseDirectoryTarget{
				version: versionEntry.Name(),
				channel: channelEntry.Name(),
			})
		}
	}
	return targets, nil
}

func (s *Service) readDirectory(directory string) ([]os.FileInfo, error) {
	handle, err := s.root.Open(directory)
	if err != nil {
		return nil, fmt.Errorf("open download directory: %w", err)
	}
	defer handle.Close()
	entries, err := handle.Readdir(-1)
	if err != nil {
		return nil, fmt.Errorf("read download directory: %w", err)
	}
	return entries, nil
}

func mergeReleaseTargets(
	left []releaseDirectoryTarget,
	right []releaseDirectoryTarget,
) []releaseDirectoryTarget {
	seen := make(map[string]struct{})
	result := make([]releaseDirectoryTarget, 0, len(left)+len(right))
	for _, target := range append(left, right...) {
		key := target.version + "/" + target.channel
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, target)
	}
	return result
}

func (s *Service) collectManifest(releaseDirectory string) (domain.Manifest, error) {
	manifest := domain.Manifest{Artifacts: make([]domain.ManifestArtifact, 0)}
	err := walkRoot(s.root, releaseDirectory, func(relativePath string, info os.FileInfo) error {
		if !info.Mode().IsRegular() || path.Base(relativePath) == "manifest.json" {
			return nil
		}
		release, ok := ParseArtifactPath(relativePath)
		if !ok || release.Version+"/"+release.Channel != releaseDirectory {
			return nil
		}
		digest, err := s.fileSHA256(relativePath)
		if err != nil {
			return err
		}
		relativeToRelease := strings.TrimPrefix(relativePath, releaseDirectory+"/")
		manifest.Artifacts = append(manifest.Artifacts, domain.ManifestArtifact{
			FileName:     path.Base(relativePath),
			Kind:         release.Kind,
			Platform:     release.Platform,
			RelativePath: relativeToRelease,
			SHA256:       digest,
			SizeBytes:    info.Size(),
			URL:          s.publicURL(relativePath),
		})
		return nil
	})
	if err != nil {
		return domain.Manifest{}, err
	}
	sort.SliceStable(manifest.Artifacts, func(left int, right int) bool {
		return manifest.Artifacts[left].RelativePath <
			manifest.Artifacts[right].RelativePath
	})
	return manifest, nil
}

// ParseArtifactPath recognizes the canonical public layout without touching
// the filesystem. The boolean is false for any non-release path.
func ParseArtifactPath(relativePath string) (*domain.FileRelease, bool) {
	normalized, err := NormalizeRelativePath(relativePath)
	if err != nil {
		return nil, false
	}
	segments := strings.Split(normalized, "/")
	if len(segments) != 5 {
		return nil, false
	}
	version := segments[0]
	channel := segments[1]
	platform := segments[2]
	kind := segments[3]
	fileName := segments[4]
	if !versionPattern.MatchString(version) ||
		!channelPattern.MatchString(channel) ||
		!platformPattern.MatchString(platform) ||
		!kindPattern.MatchString(kind) ||
		!segmentPattern.MatchString(fileName) {
		return nil, false
	}
	return &domain.FileRelease{
		Version:  version,
		Channel:  channel,
		Platform: platform,
		Kind:     kind,
	}, true
}

// NormalizeFilePath validates a writable relative path.
func NormalizeFilePath(value string) (string, error) {
	normalized, err := NormalizeRelativePath(value)
	if err != nil {
		return "", err
	}
	if path.Base(normalized) == "." || strings.HasSuffix(normalized, "/") {
		return "", domain.ErrPathInvalid
	}
	return normalized, nil
}

// NormalizeRelativePath rejects empty, absolute, traversal, backslash, and
// control-character paths before they reach os.Root.
func NormalizeRelativePath(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || strings.ContainsRune(value, '\\') ||
		strings.HasPrefix(value, "/") || path.IsAbs(value) {
		return "", domain.ErrPathInvalid
	}
	cleaned := path.Clean(value)
	if cleaned == "." || cleaned == ".." ||
		strings.HasPrefix(cleaned, "../") ||
		strings.ContainsRune(cleaned, '\x00') {
		return "", domain.ErrPathInvalid
	}
	for _, segment := range strings.Split(cleaned, "/") {
		if segment == "" || segment == "." || segment == ".." ||
			!segmentPattern.MatchString(segment) {
			return "", domain.ErrPathInvalid
		}
	}
	return cleaned, nil
}

func (s *Service) fileSHA256(relativePath string) (string, error) {
	file, err := s.root.Open(relativePath)
	if err != nil {
		return "", fmt.Errorf("open download file: %w", err)
	}
	defer file.Close()
	digest := sha256.New()
	if _, err := io.Copy(digest, file); err != nil {
		return "", fmt.Errorf("hash download file: %w", err)
	}
	return hex.EncodeToString(digest.Sum(nil)), nil
}

func (s *Service) createTemporary(parent string) (string, *os.File, error) {
	for attempt := 0; attempt < 16; attempt++ {
		name := fmt.Sprintf(".upload-%d-%d.tmp", os.Getpid(), s.clock.Now().UnixNano()+int64(attempt))
		fullPath := path.Join(parent, name)
		file, err := s.root.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err == nil {
			return name, file, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return "", nil, fmt.Errorf("create temporary download file: %w", err)
		}
	}
	return "", nil, errors.New("create temporary download file: exhausted attempts")
}

func (s *Service) loadIndex() (domain.ReleaseIndex, error) {
	payload, err := s.root.ReadFile("index.json")
	if errors.Is(err, os.ErrNotExist) {
		return domain.ReleaseIndex{
			Releases:      make([]domain.IndexRelease, 0),
			SchemaVersion: 1,
		}, nil
	}
	if err != nil {
		return domain.ReleaseIndex{}, fmt.Errorf("read download index: %w", err)
	}
	var index domain.ReleaseIndex
	if err := json.Unmarshal(payload, &index); err != nil {
		// A malformed existing index may be the CI-owned publication record.
		// Refuse to overwrite it with a partial replacement; an operator can
		// repair or remove it deliberately.
		return domain.ReleaseIndex{}, fmt.Errorf("decode download index: %w", err)
	}
	if index.Releases == nil {
		index.Releases = make([]domain.IndexRelease, 0)
	}
	if index.SchemaVersion == 0 {
		index.SchemaVersion = 1
	}
	return index, nil
}

func (s *Service) writeJSONAtomic(relativePath string, payload any) error {
	parent := path.Dir(relativePath)
	if parent != "." {
		if err := s.root.MkdirAll(parent, 0o755); err != nil {
			return fmt.Errorf("create download metadata directory: %w", err)
		}
	}
	temporaryName, temporary, err := s.createTemporary(parent)
	if err != nil {
		return err
	}
	temporaryPath := path.Join(parent, temporaryName)
	completed := false
	defer func() {
		if !completed {
			_ = s.root.Remove(temporaryPath)
		}
	}()
	encoder := json.NewEncoder(temporary)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(payload); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("encode download metadata: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("sync download metadata: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close download metadata: %w", err)
	}
	if err := s.root.Rename(temporaryPath, relativePath); err != nil {
		return fmt.Errorf("publish download metadata: %w", err)
	}
	completed = true
	return nil
}

func (s *Service) publicURL(relativePath string) string {
	if s.publicBaseURL == "" {
		return ""
	}
	return s.publicBaseURL + "/" + strings.TrimPrefix(relativePath, "/")
}

func walkRoot(
	root *os.Root,
	directory string,
	visit func(relativePath string, info os.FileInfo) error,
) error {
	directoryHandle, err := root.Open(directory)
	if err != nil {
		return fmt.Errorf("open download directory: %w", err)
	}
	defer directoryHandle.Close()
	entries, err := directoryHandle.Readdir(-1)
	if err != nil {
		return fmt.Errorf("read download directory: %w", err)
	}
	for _, entry := range entries {
		entryPath := path.Join(directory, entry.Name())
		info, err := root.Stat(entryPath)
		if err != nil {
			return fmt.Errorf("stat download entry: %w", err)
		}
		if info.IsDir() {
			if err := walkRoot(root, entryPath, visit); err != nil {
				return err
			}
			continue
		}
		if err := visit(entryPath, info); err != nil {
			return err
		}
	}
	return nil
}

func replaceIndexRelease(
	releases []domain.IndexRelease,
	entry domain.IndexRelease,
) []domain.IndexRelease {
	result := make([]domain.IndexRelease, 0, len(releases)+1)
	for _, release := range releases {
		if release.Version == entry.Version && release.Channel == entry.Channel {
			continue
		}
		result = append(result, release)
	}
	result = append(result, entry)
	sort.SliceStable(result, func(left int, right int) bool {
		if result[left].PublishedAt == result[right].PublishedAt {
			return result[left].Version > result[right].Version
		}
		return result[left].PublishedAt > result[right].PublishedAt
	})
	return result
}

func directoryName(relativePath string) string {
	directory := path.Dir(relativePath)
	if directory == "." {
		return ""
	}
	return directory
}

func isArtifactIndexed(relativePath string, indexed map[string]struct{}) bool {
	if len(indexed) == 0 {
		return false
	}
	_, exists := indexed[relativePath]
	return exists
}

func countCanonicalArtifacts(files []domain.File) int {
	count := 0
	for _, file := range files {
		if file.Release != nil {
			count++
		}
	}
	return count
}

func (s *Service) indexedArtifactSet() map[string]struct{} {
	index, err := s.loadIndex()
	if err != nil {
		return map[string]struct{}{}
	}
	indexed := make(map[string]struct{})
	for _, release := range index.Releases {
		manifestPath := path.Join(release.Version, release.Channel, "manifest.json")
		manifestPayload, readErr := s.root.ReadFile(manifestPath)
		if readErr != nil {
			continue
		}
		var manifest domain.Manifest
		if json.Unmarshal(manifestPayload, &manifest) != nil {
			continue
		}
		for _, artifact := range manifest.Artifacts {
			indexed[path.Join(release.Version, release.Channel, artifact.RelativePath)] =
				struct{}{}
		}
	}
	return indexed
}
