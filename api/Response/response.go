package response

import (
    "encoding/json"
)

type Response struct {
    Status  int               `json:"status"`
    Message string            `json:"message"`
    Data    map[string]string `json:"data,omitempty"`
}

// Helpers
func (*Response) MakeSuccessResponse(message string, data map[string]string) Response {
    return Response{Status: 200, Message: message, Data: data}
}

func (*Response) MakeErrorResponse(statusCode int, message string, errorType string) Response {
    var error = make(map[string]string)
    error["error"] = errorType
    return Response{Status: statusCode, Message: message, Data: error}
}

// 
func (r Response) Serialize() ([]byte, error) {
    return json.Marshal(r)
}
