package data

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

//go:embed nigeria.json
var nigeriaJSON []byte

// State is one Nigerian state (or the Federal Capital Territory)
// with its capital, geopolitical zone and Local Government Areas.
type State struct {
	Name    string   `json:"name"`
	Capital string   `json:"capital"`
	Zone    string   `json:"zone"`
	LGAs    []string `json:"lgas"`
}

// Nigeria is the root of the embedded administrative dataset.
type Nigeria struct {
	States []State `json:"states"`
}

var (
	nigeriaOnce sync.Once
	nigeriaData *Nigeria
	nigeriaErr  error
)

// NigeriaData returns the loaded Nigeria data. Loaded once, cached forever.
func NigeriaData() (*Nigeria, error) {
	nigeriaOnce.Do(func() {
		var n Nigeria
		if err := json.Unmarshal(nigeriaJSON, &n); err != nil {
			nigeriaErr = fmt.Errorf("parse nigeria.json: %w", err)
			return
		}
		if len(n.States) == 0 {
			nigeriaErr = fmt.Errorf("parse nigeria.json: no states found")
			return
		}
		nigeriaData = &n
	})
	return nigeriaData, nigeriaErr
}

// FindState returns the state with the given name (case-insensitive).
func FindState(name string) (*State, bool) {
	n, err := NigeriaData()
	if err != nil || n == nil {
		return nil, false
	}
	trimmed := strings.TrimSpace(name)
	for i := range n.States {
		if strings.EqualFold(n.States[i].Name, trimmed) {
			return &n.States[i], true
		}
	}
	return nil, false
}

// IsValidLGA returns true if the lga exists in the given state
// (case-insensitive).
func IsValidLGA(stateName, lga string) bool {
	st, ok := FindState(stateName)
	if !ok {
		return false
	}
	trimmed := strings.TrimSpace(lga)
	if trimmed == "" {
		return false
	}
	for _, candidate := range st.LGAs {
		if strings.EqualFold(candidate, trimmed) {
			return true
		}
	}
	return false
}
