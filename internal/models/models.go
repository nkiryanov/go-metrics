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
	Type string   `json:"type"`
	Name string   `json:"id"`
	Delta int64   `json:"delta"`
	Value float64 `json:"value"`
}

func (m Metric) MarshalJSON() ([]byte, error) {
	switch m.Type {
	case CounterTypeName:
		return json.Marshal(struct {
			Type string `json:"type"`
			ID string   `json:"id"`
			Delta int64 `json:"delta"`
		}{
			Type: m.Type,
			ID: m.Name,
			Delta: m.Delta,
		})
	case GaugeTypeName:
		return json.Marshal(struct {
			Type string `json:"type"`
			ID string `json:"id"`
			Value float64 `json:"value"`
		}{
			Type: m.Type,
			ID: m.Name,
			Value: m.Value,
		})
	default:
		return nil, fmt.Errorf("unsupported metric type: %s", m.Type)
	}
}

func (m *Metric) UnmarshalJSON(data []byte) error {
	temp := struct {
		Type string `json:"type"`
		Name string `json:"id"`
		Delta *int64 `json:"delta,omitempty"`
		Value *float64 `json:"value,omitempty"`
	}{}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	switch temp.Type {
	case CounterTypeName:
		if temp.Delta != nil {
			m.Type = temp.Type
			m.Name = temp.Name
			m.Delta = *temp.Delta
		} else {
			return fmt.Errorf("missing 'delta' for 'counter' type")
		}
	case GaugeTypeName:
		if temp.Value != nil {
			m.Type = temp.Type
			m.Name = temp.Name
			m.Value = *temp.Value
		} else {
			return fmt.Errorf("missing 'value' for 'gauge' type")
		}
	}
	
	return fmt.Errorf("unsupported type: '%s'", temp.Type)
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
