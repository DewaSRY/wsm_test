package marketplace

import (
	"backend/internal/domain"
	"fmt"
)

type AuthorizeResponse struct {
	Message string `json:"message"`
	Data    struct {
		Code   string `json:"code"`
		ShopID string `json:"shop_id"`
		State  string `json:"state"`
	} `json:"data"`
}

type TokenAPIResponse struct {
	Message string               `json:"message"`
	Data    domain.TokenResponse `json:"data"`
}

type APIErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

type APIError struct {
	StatusCode int
	Message    string
	ErrorMsg   string
}

func (e *APIError) Error() string {
	if e.ErrorMsg != "" {
		return fmt.Sprintf("API error %d: %s - %s", e.StatusCode, e.Message, e.ErrorMsg)
	}
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}
