package errors

type ErrorResponse struct {
	StatusCode int    `json:"status_code"`
	ErrorCode  string `json:"error_code"`
	Message    string `json:"message"`
}

func ToHTTPResponse(err error) *ErrorResponse {
	if customErr, ok := err.(*CustomError); ok {
		return &ErrorResponse{
			StatusCode: customErr.HTTPStatus,
			ErrorCode:  customErr.ErrorCode,
			Message:    customErr.Message,
		}
	}

	return &ErrorResponse{
		StatusCode: 500,
		ErrorCode:  CodeInternalServerError,
		Message:    err.Error(),
	}
}