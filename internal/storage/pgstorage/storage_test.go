package pgstorage

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/nkiryanov/go-metrics/internal/db"
	"github.com/nkiryanov/go-metrics/internal/models"
)

func Test_PgStorage(t *testing.T) {
	container, err := postgres.Run(t.Context(),
		"postgres:17-alpine",
		postgres.WithDatabase("go-metrics-test"),
		postgres.WithUsername("go-metrics"),
		postgres.WithPassword("pwd"),
		postgres.BasicWaitStrategies(),
		testcontainers.CustomizeRequestOption(func(req *testcontainers.GenericContainerRequest) error {
			req.ExposedPorts = []string{"25432:5432"}
			return nil
		}),
	)
	defer testcontainers.CleanupContainer(t, container)
	require.NoError(t, err, "container with pg start failed")

	dbURI, err := container.ConnectionString(t.Context())
	require.NoError(t, err)
	t.Logf("Container with pg started, dbURI=%v", dbURI)

	// Migrate and request connection pool
	err = db.Migrate(dbURI)
	require.NoError(t, err)
	pool, err := db.Connect(t.Context(), dbURI)
	require.NoError(t, err)
	defer pool.Close()

	// Helper to run tests with its own PgStorage in transaction
	// When test end rollback
	withTx := func(pool *pgxpool.Pool, t *testing.T, testFunc func(*PgStorage)) {
		tx, err := pool.Begin(t.Context())
		require.NoError(t, err)

		defer func() {
			err := tx.Rollback(t.Context())
			require.NoError(t, err)
		}()

		txStorage := New(t.Context(), tx)
		testFunc(txStorage)
	}

	t.Run("ping ok", func(t *testing.T) {
		withTx(pool, t, func(s *PgStorage) {
			err := s.Ping(t.Context())

			require.NoError(t, err)
		})
	})

	t.Run("update gauge ok", func(t *testing.T) {
		withTx(pool, t, func(s *PgStorage) {
			gauge := models.Metric{Name: "cpu", Type: "gauge", Value: 2234.232}

			// Gauge update once ok
			updated, err := s.UpdateMetric(t.Context(), &gauge)
			require.NoError(t, err)
			require.Equal(t, gauge, updated)

			// Gauge metric update idempotent
			updated, err = s.UpdateMetric(t.Context(), &gauge)
			require.NoError(t, err)
			require.EqualValues(t, gauge.Value, updated.Value)
		})
	})

	t.Run("update counter ok", func(t *testing.T) {
		withTx(pool, t, func(s *PgStorage) {
			counter := models.Metric{Name: "task", Type: "counter", Delta: 2}

			// Counter update once ok
			got, err := s.UpdateMetric(t.Context(), &counter)

			require.NoError(t, err)
			require.Equal(t, counter, got)

			// Counter update actually update counter by value
			got, err = s.UpdateMetric(t.Context(), &counter)

			require.NoError(t, err)
			require.EqualValues(t, 4, got.Delta, "Counter update should increment delta")
		})
	})

	t.Run("update bulk ok", func(t *testing.T) {
		withTx(pool, t, func(s *PgStorage) {
			metrics := []models.Metric{
				{Name: "task", Type: "counter", Delta: 2},
				{Name: "task", Type: "counter", Delta: 1},
				{Name: "cpu", Type: "gauge", Value: 34.2},
				{Name: "cpu", Type: "gauge", Value: 34.2},
			}

			got, err := s.UpdateMetricBulk(t.Context(), metrics)

			require.NoError(t, err)
			require.Len(t, got, 4)
			assert.EqualValues(t, models.Metric{Name: "task", Type: "counter", Delta: 2}, got[0])
			assert.EqualValues(t, models.Metric{Name: "task", Type: "counter", Delta: 3}, got[1], "Existed counter should incremented")
			assert.EqualValues(t, models.Metric{Name: "cpu", Type: "gauge", Value: 34.2}, got[2])
			assert.EqualValues(t, models.Metric{Name: "cpu", Type: "gauge", Value: 34.2}, got[3], "Existed gauge should overridden")

		})
	})

	t.Run("get metric ok", func(t *testing.T) {
		withTx(pool, t, func(s *PgStorage) {
			counter := models.Metric{Name: "task", Type: "counter", Delta: 23}
			_, err := s.UpdateMetric(t.Context(), &counter)
			require.NoError(t, err)

			// Get known metric
			got, err := s.GetMetric(t.Context(), "counter", "task")

			assert.NoError(t, err)
			assert.EqualValues(t, counter, got, "Known metrics must captured without error")

			// Get unknown metric
			_, err = s.GetMetric(t.Context(), "gauge", "cpu")

			assert.Error(t, err, "Error on unknown metrics must return")
		})
	})

	t.Run("count metrics ok", func(t *testing.T) {
		withTx(pool, t, func(s *PgStorage) {
			_, err := s.UpdateMetricBulk(t.Context(), []models.Metric{
				{Name: "task", Type: "counter", Delta: 23},
				{Name: "task", Type: "counter", Delta: 1},
				{Name: "cpu", Type: "gauge", Value: 13.23},
			})
			require.NoError(t, err)

			got, err := s.CountMetric(t.Context())

			require.NoError(t, err)
			assert.EqualValues(t, 2, got, "Uniq metrics must be counted")
		})
	})

	t.Run("list metrics ok", func(t *testing.T) {
		withTx(pool, t, func(s *PgStorage) {
			_, err := s.UpdateMetricBulk(t.Context(), []models.Metric{
				{Name: "task", Type: "counter", Delta: 23},
				{Name: "task", Type: "counter", Delta: 1},
				{Name: "cpu", Type: "gauge", Value: 13.23},
				{Name: "amount", Type: "counter", Delta: 2},
			})
			require.NoError(t, err)

			got, err := s.ListMetric(t.Context())

			require.NoError(t, err)
			require.Len(t, got, 3, "Only uniq metrics must return")
			require.Equal(t, []string{"amount", "cpu", "task"}, []string{got[0].Name, got[1].Name, got[2].Name}, "Must ordered by name")
			assert.Equal(t, models.Metric{Name: "amount", Type: "counter", Delta: 2}, got[0])
			assert.Equal(t, models.Metric{Name: "cpu", Type: "gauge", Value: 13.23}, got[1])
			assert.Equal(t, models.Metric{Name: "task", Type: "counter", Delta: 24}, got[2])
		})
	})
}
