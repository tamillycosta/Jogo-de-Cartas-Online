package utils


import (
    "encoding/json"
   
	
)


// Serializa qualquer struct em JSON
func Encode(v interface{}) string {
    bytes, err := json.Marshal(v)
    if err != nil {
        return "erro ao serializar: " 
    }
    return string(bytes)
}


