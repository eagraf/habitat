package api

import (
	"encoding/json"
	"net/http"
)

func WriteError(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	w.Write([]byte(err.Error() + "\n"))
}

func WriteResponse(w http.ResponseWriter, res interface{}) {
	body, err := json.Marshal(res)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err)
	}
	w.Write(body)
}
