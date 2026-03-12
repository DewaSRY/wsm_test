package domain

type APIResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type PaginatedResponse struct {
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

func SuccessResponse(message string, data interface{}) *APIResponse {
	return &APIResponse{
		Message: message,
		Data:    data,
	}
}

func ErrorResponse(message string, err string) *APIResponse {
	return &APIResponse{
		Message: message,
		Error:   err,
	}
}
