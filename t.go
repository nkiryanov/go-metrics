package main

import (
	"context"
	"fmt"
	_ "github.com/jackc/pgx/v5"
	"github.com/nkiryanov/go-metrics/internal/storage/pgstorage"
	"log/slog"

	"github.com/nkiryanov/go-metrics/internal/models"
)

func main() {
	ctx := context.TODO()
	s := pgstorage.New(ctx, "postgres://go-metrics@localhost:15432/go-metrics")

	metrics := []models.Metric{
		{Type: "counter", Name: "some", Delta: 23},
		{Type: "counter", Name: "other", Delta: 12},
		{Type: "gauge", Name: "some", Value: 23.12},
	}

	for _, m := range metrics {
		if updated, err := s.UpdateMetric(ctx, &m); err != nil {
			slog.Error("cant update metric", "error", err.Error())
		} else {
			fmt.Printf("%#v\n", updated)
		}
	}

	count, err := s.Count(ctx)
	fmt.Printf("count = %d, err = %s", count, err)
}
