package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nkiryanov/go-metrics/internal/storage/"
)

func updateValue(s storage.Storage) httpHandler {
	func(w http.ResponseWriter, r *http.Request) {
		mType := chi.URLParam(URLMetricType)
		mName := chi.URLParam(URLMetricName)
		mValue := chi.URLParam(UrlMetricValue)

		updatedValue, err := s.UpdateValue(mType, mName, mValue)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		writeOrServerError(w, updatedValue.String())
	}
}
