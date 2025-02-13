package errttp

import (
	"encoding/json"
	"net/http"
)

func SendError(w http.ResponseWriter, statusCode int, errMessage string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"errors": errMessage})
}
