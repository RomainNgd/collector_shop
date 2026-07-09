package controllers

import (
	"errors"
	"net/http"
	"poc-gin/services"
	"testing"
)

func TestProfileHandlerGetProfileReturnsStats(t *testing.T) {
	handler := NewProfileHandler(&mockProfileService{
		getStatsFn: func(userID uint) (*services.ProfileStats, error) {
			if userID != 1 {
				t.Fatalf("expected user id 1, got %d", userID)
			}
			return &services.ProfileStats{
				Email:          "collector@example.com",
				ProductsBought: 3,
				ListingsPosted: 2,
				ProductsSold:   4,
			}, nil
		},
	})

	recorder := performAuthenticatedJSONRequest(handler.GetProfile, http.MethodGet, "/profile", nil)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	resp := decodeAPIResponse(recorder, t)
	if !resp.Success {
		t.Fatalf("expected success response, got %#v", resp)
	}
	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected object data, got %#v", resp.Data)
	}
	if data["email"] != "collector@example.com" {
		t.Fatalf("expected email in payload, got %#v", data)
	}
	if data["products_bought"] != float64(3) || data["listings_posted"] != float64(2) || data["products_sold"] != float64(4) {
		t.Fatalf("expected stats in payload, got %#v", data)
	}
}

func TestProfileHandlerGetProfileRequiresAuthContext(t *testing.T) {
	handler := NewProfileHandler(&mockProfileService{})

	recorder := performJSONRequest(handler.GetProfile, http.MethodGet, "/profile", nil)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", recorder.Code)
	}
}

func TestProfileHandlerGetProfileMapsServiceErrors(t *testing.T) {
	cases := []struct {
		name           string
		serviceErr     error
		expectedStatus int
		expectedCode   string
	}{
		{name: "user not found", serviceErr: services.ErrUserNotFound, expectedStatus: http.StatusNotFound, expectedCode: "USER_NOT_FOUND"},
		{name: "internal error", serviceErr: errors.New("boom"), expectedStatus: http.StatusInternalServerError, expectedCode: "INTERNAL_ERROR"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			handler := NewProfileHandler(&mockProfileService{
				getStatsFn: func(uint) (*services.ProfileStats, error) {
					return nil, testCase.serviceErr
				},
			})

			recorder := performAuthenticatedJSONRequest(handler.GetProfile, http.MethodGet, "/profile", nil)

			if recorder.Code != testCase.expectedStatus {
				t.Fatalf("expected status %d, got %d", testCase.expectedStatus, recorder.Code)
			}
			resp := decodeAPIResponse(recorder, t)
			if resp.Error == nil || resp.Error.Code != testCase.expectedCode {
				t.Fatalf("expected error code %s, got %#v", testCase.expectedCode, resp.Error)
			}
		})
	}
}
