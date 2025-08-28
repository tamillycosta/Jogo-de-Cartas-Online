package request


import(
     "encoding/json"
)

type Request struct {
	User string 			`json:"user,omitempty"`
    Method string            `json:"method"`
    Params map[string]string `json:"params"`
}


// desserialização
func Deserialize(data []byte) (Request, error) {
    var req Request
    err := json.Unmarshal(data, &req)
    return req, err
}
