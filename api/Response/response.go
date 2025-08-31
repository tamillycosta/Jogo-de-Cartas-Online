package response



import (
    "encoding/json"
    "fmt"
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



// Desserializa JSON para Response
func DecodeResponse(data string) (Response, error) {
    var resp Response
    err := json.Unmarshal([]byte(data), &resp)
    if err != nil {
        return Response{}, fmt.Errorf("erro ao decodificar response: %w", err)
    }
    return resp, nil
}


