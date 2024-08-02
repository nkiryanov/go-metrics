package handlers

import (
	"net/http"
)

func UpdateMetricHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello"))  // nolint: errcheck
}
