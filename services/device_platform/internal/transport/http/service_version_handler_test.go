package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	serviceversiondomain "github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/service_version/domain"
	serviceversionservice "github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/service_version/service"
)

type fakeServiceVersionAdminService struct {
	serviceVersionAdminService
	listReleasesCalls []listReleasesCall
}

type listReleasesCall struct {
	serviceID string
	page      int
	pageSize  int
}

func (service *fakeServiceVersionAdminService) ListReleases(
	_ context.Context,
	serviceID string,
	page int,
	pageSize int,
) (serviceversiondomain.ReleasePage, error) {
	service.listReleasesCalls = append(service.listReleasesCalls, listReleasesCall{
		serviceID: serviceID,
		page:      page,
		pageSize:  pageSize,
	})
	return serviceversiondomain.ReleasePage{
		Service:  serviceID,
		Releases: []serviceversiondomain.Release{},
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func TestListServiceReleasesDefaultsToCatalogueLimit(t *testing.T) {
	service := &fakeServiceVersionAdminService{}
	handler := adminHandler{serviceVersionService: service}
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/admin/service-versions/device_platform/releases",
		nil,
	)
	request.SetPathValue("service", "device_platform")
	response := httptest.NewRecorder()

	handler.listServiceReleases(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if len(service.listReleasesCalls) != 1 {
		t.Fatalf("expected one release call, got %d", len(service.listReleasesCalls))
	}
	call := service.listReleasesCalls[0]
	if call.serviceID != "device_platform" || call.page != 1 ||
		call.pageSize != serviceversionservice.ReleaseCatalogLimit {
		t.Fatalf("unexpected release call: %+v", call)
	}
}

func TestListServiceReleasesPassesExplicitPageSizeForServiceClamping(t *testing.T) {
	service := &fakeServiceVersionAdminService{}
	handler := adminHandler{serviceVersionService: service}
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/admin/service-versions/device_platform/releases?page_size=250",
		nil,
	)
	request.SetPathValue("service", "device_platform")
	response := httptest.NewRecorder()

	handler.listServiceReleases(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if len(service.listReleasesCalls) != 1 {
		t.Fatalf("expected one release call, got %d", len(service.listReleasesCalls))
	}
	if service.listReleasesCalls[0].pageSize != 250 {
		t.Fatalf(
			"expected explicit page size 250 to reach the service layer, got %d",
			service.listReleasesCalls[0].pageSize,
		)
	}
}
