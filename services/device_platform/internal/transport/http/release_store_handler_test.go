package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	releaseStoredomain "github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/release_store/domain"
)

type fakeReleaseStoreAdminService struct {
	inventory       releaseStoredomain.Inventory
	uploaded        releaseStoredomain.File
	refreshResult   releaseStoredomain.IndexRefreshResult
	err             error
	uploadPath      string
	uploadOverwrite bool
	uploadBody      string
	deletedPath     string
	refreshInput    releaseStoredomain.IndexRefreshRequest
}

func (service *fakeReleaseStoreAdminService) Inventory(
	_ context.Context,
) (releaseStoredomain.Inventory, error) {
	return service.inventory, service.err
}

func (service *fakeReleaseStoreAdminService) IndexStatus(
	_ context.Context,
) (releaseStoredomain.IndexStatus, error) {
	return releaseStoredomain.IndexStatus{}, service.err
}

func (service *fakeReleaseStoreAdminService) RefreshAllIndexes(
	_ context.Context,
) (releaseStoredomain.IndexRefreshSummary, error) {
	return releaseStoredomain.IndexRefreshSummary{}, service.err
}

func (service *fakeReleaseStoreAdminService) UploadFile(
	_ context.Context,
	input releaseStoredomain.UploadInput,
	reader io.Reader,
) (releaseStoredomain.File, error) {
	payload, _ := io.ReadAll(reader)
	service.uploadPath = input.RelativePath
	service.uploadOverwrite = input.Overwrite
	service.uploadBody = string(payload)
	return service.uploaded, service.err
}

func (service *fakeReleaseStoreAdminService) DeleteFile(
	_ context.Context,
	relativePath string,
) error {
	service.deletedPath = relativePath
	return service.err
}

func (service *fakeReleaseStoreAdminService) RefreshIndex(
	_ context.Context,
	input releaseStoredomain.IndexRefreshRequest,
) (releaseStoredomain.IndexRefreshResult, error) {
	service.refreshInput = input
	return service.refreshResult, service.err
}

func TestListReleaseFilesReturnsInventory(t *testing.T) {
	service := &fakeReleaseStoreAdminService{
		inventory: releaseStoredomain.Inventory{
			RootDir:       "/srv/releases",
			PublicBaseURL: "https://download.example.test",
			Files: []releaseStoredomain.File{
				{
					RelativePath: "0.12.3/stable/android/apk/app.apk",
					FileName:     "app.apk",
					SizeBytes:    12,
					SHA256:       strings.Repeat("a", 64),
				},
			},
		},
	}
	handler := adminHandler{releaseStoreService: service}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/release-files", nil)
	response := httptest.NewRecorder()

	handler.listReleaseFiles(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	var envelope struct {
		Data releaseStoredomain.Inventory `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode inventory response: %v", err)
	}
	if len(envelope.Data.Files) != 1 ||
		envelope.Data.Files[0].RelativePath != "0.12.3/stable/android/apk/app.apk" {
		t.Fatalf("unexpected inventory response: %s", response.Body.String())
	}
}

func TestUploadReleaseFilePassesPathAndStream(t *testing.T) {
	service := &fakeReleaseStoreAdminService{
		uploaded: releaseStoredomain.File{
			RelativePath: "0.12.3/stable/any/notes/manual.txt",
			SHA256:       strings.Repeat("b", 64),
		},
	}
	handler := adminHandler{releaseStoreService: service}
	request := newUploadRequest(
		t,
		"0.12.3/stable/any/notes/manual.txt",
		"true",
		"manual upload",
	)
	response := httptest.NewRecorder()

	handler.uploadReleaseFile(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
	if service.uploadPath != "0.12.3/stable/any/notes/manual.txt" ||
		!service.uploadOverwrite ||
		service.uploadBody != "manual upload" {
		t.Fatalf(
			"unexpected upload call path=%q overwrite=%v body=%q",
			service.uploadPath,
			service.uploadOverwrite,
			service.uploadBody,
		)
	}
}

func TestUploadReleaseFileMapsServiceErrors(t *testing.T) {
	testCases := []struct {
		name       string
		serviceErr error
		status     int
		code       string
	}{
		{
			name:       "path traversal",
			serviceErr: releaseStoredomain.ErrPathInvalid,
			status:     http.StatusUnprocessableEntity,
			code:       "invalid_path",
		},
		{
			name:       "duplicate",
			serviceErr: releaseStoredomain.ErrFileAlreadyExists,
			status:     http.StatusConflict,
			code:       "file_exists",
		},
		{
			name:       "too large",
			serviceErr: releaseStoredomain.ErrFileTooLarge,
			status:     http.StatusRequestEntityTooLarge,
			code:       "file_too_large",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := &fakeReleaseStoreAdminService{err: testCase.serviceErr}
			handler := adminHandler{releaseStoreService: service}
			request := newUploadRequest(t, "manual.txt", "false", "payload")
			response := httptest.NewRecorder()

			handler.uploadReleaseFile(response, request)

			if response.Code != testCase.status {
				t.Fatalf("expected status %d, got %d", testCase.status, response.Code)
			}
			var envelope struct {
				Error struct {
					Code string `json:"code"`
				} `json:"error"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
				t.Fatalf("decode error response: %v", err)
			}
			if envelope.Error.Code != testCase.code {
				t.Fatalf("expected error code %q, got %q", testCase.code, envelope.Error.Code)
			}
		})
	}
}

func TestDeleteReleaseFilePassesPath(t *testing.T) {
	service := &fakeReleaseStoreAdminService{}
	handler := adminHandler{releaseStoreService: service}
	request := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/admin/storage/files/0.12.3/stable/any/notes/manual.txt",
		nil,
	)
	request.SetPathValue("path", "0.12.3/stable/any/notes/manual.txt")
	response := httptest.NewRecorder()

	handler.deleteReleaseFile(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if service.deletedPath != "0.12.3/stable/any/notes/manual.txt" {
		t.Fatalf("unexpected deleted path: %q", service.deletedPath)
	}
}

func TestStorageDeleteRouteCapturesNestedRelativePath(t *testing.T) {
	service := &fakeReleaseStoreAdminService{}
	request := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/admin/storage/files/0.12.3/stable/android/apk/app.apk",
		nil,
	)
	request.SetPathValue("path", "0.12.3/stable/android/apk/app.apk")
	handler := adminHandler{releaseStoreService: service}
	response := httptest.NewRecorder()

	handler.deleteReleaseFile(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if service.deletedPath != "0.12.3/stable/android/apk/app.apk" {
		t.Fatalf("unexpected deleted path: %q", service.deletedPath)
	}
}

func TestRefreshReleaseIndexPassesVersionMetadata(t *testing.T) {
	service := &fakeReleaseStoreAdminService{
		refreshResult: releaseStoredomain.IndexRefreshResult{
			Version:       "0.12.3",
			Channel:       "stable",
			ArtifactCount: 2,
		},
	}
	handler := adminHandler{releaseStoreService: service}
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/admin/release-index/refresh",
		strings.NewReader(
			`{"version":"0.12.3","channel":"stable","title":"如此萌屋平台 0.12.3",`+
				`"published_at":"2026-10-04T08:00:00Z"}`,
		),
	)
	response := httptest.NewRecorder()

	handler.refreshReleaseIndex(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if service.refreshInput.Version != "0.12.3" ||
		service.refreshInput.Channel != "stable" ||
		service.refreshInput.Title != "如此萌屋平台 0.12.3" ||
		service.refreshInput.PublishedAt != "2026-10-04T08:00:00Z" {
		t.Fatalf("unexpected refresh input: %+v", service.refreshInput)
	}
}

func newUploadRequest(
	t *testing.T,
	relativePath string,
	overwrite string,
	body string,
) *http.Request {
	t.Helper()
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)
	if err := writer.WriteField("path", relativePath); err != nil {
		t.Fatalf("write path field: %v", err)
	}
	if err := writer.WriteField("overwrite", overwrite); err != nil {
		t.Fatalf("write overwrite field: %v", err)
	}
	filePart, err := writer.CreateFormFile("file", "artifact.bin")
	if err != nil {
		t.Fatalf("create file field: %v", err)
	}
	if _, err := filePart.Write([]byte(body)); err != nil {
		t.Fatalf("write file field: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/admin/release-files",
		&requestBody,
	)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request
}
