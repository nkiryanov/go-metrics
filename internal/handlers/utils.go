package handlers

import (
	"log/slog"
	"net/http"
)

func writeOrServerError(w http.ResponseWriter, value string) {
	_, err := w.Write([]byte(value))
	if err != nil {
		slog.Error("write response error", "error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
