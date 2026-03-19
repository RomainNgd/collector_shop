package controllers

import (
	"errors"
	"net/http"
	"poc-gin/models"
	"poc-gin/services"
	"testing"
)

func TestAuthHandlerRegister(t *testing.T) {
	t.Run("returns 400 for invalid payload", func(t *testing.T) {
		handler := NewAuthHandler(&mockAuthService{})

		recorder := performJSONRequest(handler.Register, http.MethodPost, "/auth/register", map[string]any{
			"email":    "invalid",
			"password": "short",
		})

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", recorder.Code)
		}
	})

	t.Run("returns 409 for duplicated email", func(t *testing.T) {
		handler := NewAuthHandler(&mockAuthService{
			registerFn: func(email, password string) (*models.User, error) {
				return nil, services.ErrEmailAlreadyUsed
			},
		})

		recorder := performJSONRequest(handler.Register, http.MethodPost, "/auth/register", map[string]any{
			"email":    "john@example.com",
			"password": "password123",
		})

		resp := decodeAPIResponse(recorder, t)
		if recorder.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", recorder.Code)
		}
		if resp.Error == nil || resp.Error.Code != "EMAIL_ALREADY_USED" {
			t.Fatalf("unexpected error response: %+v", resp.Error)
		}
	})

	t.Run("returns 500 on internal error", func(t *testing.T) {
		handler := NewAuthHandler(&mockAuthService{
			registerFn: func(email, password string) (*models.User, error) {
				return nil, errors.New("db failure")
			},
		})

		recorder := performJSONRequest(handler.Register, http.MethodPost, "/auth/register", map[string]any{
			"email":    "john@example.com",
			"password": "password123",
		})

		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", recorder.Code)
		}
	})

	t.Run("returns 201 on success", func(t *testing.T) {
		handler := NewAuthHandler(&mockAuthService{
			registerFn: func(email, password string) (*models.User, error) {
				return &models.User{Email: email, Role: "ROLE_USER"}, nil
			},
		})

		recorder := performJSONRequest(handler.Register, http.MethodPost, "/auth/register", map[string]any{
			"email":    "john@example.com",
			"password": "password123",
		})

		resp := decodeAPIResponse(recorder, t)
		if recorder.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", recorder.Code)
		}
		data, ok := resp.Data.(map[string]any)
		if !ok {
			t.Fatalf("expected object data, got %#v", resp.Data)
		}
		if data["email"] != "john@example.com" {
			t.Fatalf("unexpected email: %#v", data["email"])
		}
		if data["role"] != "ROLE_USER" {
			t.Fatalf("unexpected role: %#v", data["role"])
		}
	})
}

func TestAuthHandlerLogin(t *testing.T) {
	t.Run("returns 400 for invalid payload", func(t *testing.T) {
		handler := NewAuthHandler(&mockAuthService{})

		recorder := performJSONRequest(handler.Login, http.MethodPost, "/auth/login", map[string]any{
			"email": "invalid",
		})

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", recorder.Code)
		}
	})

	t.Run("returns 401 for invalid credentials", func(t *testing.T) {
		handler := NewAuthHandler(&mockAuthService{
			loginFn: func(email, password string) (string, error) {
				return "", services.ErrInvalidCredentials
			},
		})

		recorder := performJSONRequest(handler.Login, http.MethodPost, "/auth/login", map[string]any{
			"email":    "john@example.com",
			"password": "password123",
		})

		resp := decodeAPIResponse(recorder, t)
		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", recorder.Code)
		}
		if resp.Error == nil || resp.Error.Code != "INVALID_CREDENTIALS" {
			t.Fatalf("unexpected error response: %+v", resp.Error)
		}
	})

	t.Run("returns 500 on internal error", func(t *testing.T) {
		handler := NewAuthHandler(&mockAuthService{
			loginFn: func(email, password string) (string, error) {
				return "", errors.New("jwt failure")
			},
		})

		recorder := performJSONRequest(handler.Login, http.MethodPost, "/auth/login", map[string]any{
			"email":    "john@example.com",
			"password": "password123",
		})

		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", recorder.Code)
		}
	})

	t.Run("returns 200 with token on success", func(t *testing.T) {
		handler := NewAuthHandler(&mockAuthService{
			loginFn: func(email, password string) (string, error) {
				return "signed-token", nil
			},
		})

		recorder := performJSONRequest(handler.Login, http.MethodPost, "/auth/login", map[string]any{
			"email":    "john@example.com",
			"password": "password123",
		})

		resp := decodeAPIResponse(recorder, t)
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", recorder.Code)
		}
		data, ok := resp.Data.(map[string]any)
		if !ok {
			t.Fatalf("expected object data, got %#v", resp.Data)
		}
		if data["token"] != "signed-token" {
			t.Fatalf("unexpected token: %#v", data["token"])
		}
	})
}
