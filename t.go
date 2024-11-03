package main

import (
	"context"
	"fmt"
	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nkiryanov/go-metrics/internal/storage/pgstorage"
	"log"
	"log/slog"

	"github.com/nkiryanov/go-metrics/internal/models"
)

func main() {
	dbpool, err := pgxpool.New(context.Background(), "postgres://go-metrics@localhost:15432/go-metrics")
	if err != nil {
		log.Fatalf("Can't connect to db %s", err)
	}

	metrics := []models.Metric{
		{Type: "counter", Name: "some", Delta: 23},
		{Type: "counter", Name: "other", Delta: 12},
		{Type: "gauge", Name: "some", Value: 23.12},
	}

	s := pgstorage.New(dbpool)
	ctx := context.TODO()

	for _, m := range metrics {
		if updated, err := s.UpdateMetric(ctx, &m); err != nil {
			slog.Error("cant update metric", "error", err.Error())
		} else {
			fmt.Printf("%#v\n", updated)
		}
	}
}
