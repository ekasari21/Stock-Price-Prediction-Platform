package model

type ApiResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func SuccessResponse(data interface{}, message ...string) ApiResponse {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	return ApiResponse{
		Success: true,
		Message: msg,
		Data:    data,
	}
}

func ErrorResponse(err string) ApiResponse {
	return ApiResponse{
		Success: false,
		Error:   err,
	}
}
