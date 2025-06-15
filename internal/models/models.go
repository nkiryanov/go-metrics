package models

import (
	"strconv"

	"encoding/json"
	"fmt"
)

const (
	CounterTypeName = "counter"
	GaugeTypeName   = "gauge"
)

type Metric struct {
	Type  string  `json:"type"`
	Name  string  `json:"id"`
	Delta int64   `json:"delta"`
	Value float64 `json:"value"`
}

func (m Metric) MarshalJSON() ([]byte, error) {
	switch m.Type {
	case CounterTypeName:
		return json.Marshal(struct {
			ID    string `json:"id"`
			Type  string `json:"type"`
			Delta int64  `json:"delta"`
		}{
			ID:    m.Name,
			Type:  m.Type,
			Delta: m.Delta,
		})
	case GaugeTypeName:
		return json.Marshal(struct {
			ID    string  `json:"id"`
			Type  string  `json:"type"`
			Value float64 `json:"value"`
		}{
			ID:    m.Name,
			Type:  m.Type,
			Value: m.Value,
		})
	default:
		return nil, fmt.Errorf("unsupported metric type: %s", m.Type)
	}
}

func (m *Metric) UnmarshalJSON(data []byte) error {
	type MetricAlias Metric
	aux := (*MetricAlias)(m)

	err := json.Unmarshal(data, &aux)
	if err != nil {
		return err
	}

	err = m.Validate()
	if err != nil {
		return err
	}

	return nil
}

func (m *Metric) String() string {
	switch m.Type {
	case CounterTypeName:
		return strconv.FormatInt(m.Delta, 10)
	case GaugeTypeName:
		return strconv.FormatFloat(m.Value, 'f', -1, 64)
	default:
		return ""
	}
}

// Validate Metic invariant
func (m *Metric) Validate() error {
	switch {
	case m.Type != CounterTypeName && m.Type != GaugeTypeName:
		return fmt.Errorf("unknown metric type: %s", m.Type)
	case m.Name == "":
		return fmt.Errorf("empty metric name")
	case m.Type == CounterTypeName && m.Value != 0:
		return fmt.Errorf("counters must have Value equal zero, got '%f'", m.Value)
	case m.Type == GaugeTypeName && m.Delta != 0:
		return fmt.Errorf("gauges must have Delta equal zero, got '%d'", m.Delta)
	}

	return nil
}
