package flexibee

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Response is the top-level Flexibee API response wrapper.
type Response struct {
	Winstrom ResponseEnvelope `json:"winstrom"`
}

// ResponseEnvelope contains metadata and records from the API.
type ResponseEnvelope struct {
	Version  string
	RowCount *int
	Records  []map[string]any
}

// FlexibeeBool handles Flexibee's habit of returning booleans as strings ("true"/"false").
type FlexibeeBool bool

func (b *FlexibeeBool) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*b = FlexibeeBool(s == "true")
		return nil
	}
	var v bool
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*b = FlexibeeBool(v)
	return nil
}

// FlexibeeInt handles Flexibee returning integers as strings (e.g. "20").
type FlexibeeInt int

func (i *FlexibeeInt) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		n, err := strconv.Atoi(s)
		if err != nil {
			*i = 0
			return nil
		}
		*i = FlexibeeInt(n)
		return nil
	}
	var v int
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*i = FlexibeeInt(v)
	return nil
}

// Property describes a single field in a Flexibee evidence type.
type Property struct {
	Name      string       `json:"propertyName"`
	Type      string       `json:"type"`
	MaxLength FlexibeeInt  `json:"maxLength"`
	Mandatory FlexibeeBool `json:"mandatory"`
	ReadOnly  FlexibeeBool `json:"isReadOnly"`
}

// EvidenceInfo describes an available evidence type.
type EvidenceInfo struct {
	EvidencePath string `json:"evidencePath"`
	EvidenceName string `json:"evidenceName"`
}

// FetchOptions controls how evidence records are fetched.
type FetchOptions struct {
	Limit       int
	Start       int
	Detail      string // "full", "summary", "id", "custom:..."
	Filter      string // Flexibee filter expression
	AddRowCount bool
}

// parseResponse parses a raw JSON response, extracting records from the
// evidence-specific key in the winstrom envelope.
// Flexibee returns @rowCount as a string, so we parse everything manually.
func parseResponse(data []byte, evidence string) (*Response, error) {
	var raw struct {
		Winstrom map[string]json.RawMessage `json:"winstrom"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	resp := &Response{}

	if v, ok := raw.Winstrom["@version"]; ok {
		var version string
		if err := json.Unmarshal(v, &version); err == nil {
			resp.Winstrom.Version = version
		}
	}

	if v, ok := raw.Winstrom["@rowCount"]; ok {
		// Flexibee returns this as a string (e.g. "3") or number
		var s string
		if err := json.Unmarshal(v, &s); err == nil {
			if n, err := strconv.Atoi(s); err == nil {
				resp.Winstrom.RowCount = &n
			}
		} else {
			var n int
			if err := json.Unmarshal(v, &n); err == nil {
				resp.Winstrom.RowCount = &n
			}
		}
	}

	if recordsRaw, ok := raw.Winstrom[evidence]; ok {
		var records []map[string]any
		if err := json.Unmarshal(recordsRaw, &records); err != nil {
			return nil, fmt.Errorf("unmarshal records for %s: %w", evidence, err)
		}
		resp.Winstrom.Records = records
	}

	return resp, nil
}

// parseProperties parses a properties endpoint response.
func parseProperties(data []byte) ([]Property, error) {
	var wrapper struct {
		Properties struct {
			Property []Property `json:"property"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("unmarshal properties: %w", err)
	}
	return wrapper.Properties.Property, nil
}
