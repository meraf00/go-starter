package transport

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildWhereClause_Contains(t *testing.T) {
	allowed := map[string]AllowedField{
		"attributes":      {Column: "v.attributes", Type: FieldJSONB, JSONBCol: "attributes"},
		"attributes.size": {Column: "v.attributes", Type: FieldJSONB, JSONBCol: "attributes"},
	}

	tests := []struct {
		name       string
		filters    []Filter
		wantClause string
		wantArgs   []any
		wantErr    bool
	}{
		{
			name: "JSONB Contains",
			filters: []Filter{
				{Field: "attributes", Op: OpContains, Value: `{"color":"Red"}`},
			},
			wantClause: "attributes @> $1::jsonb",
			wantArgs:   []any{`{"color":"Red"}`},
			wantErr:    false,
		},
		{
			name: "JSONB Contains and Eq",
			filters: []Filter{
				{Field: "attributes", Op: OpContains, Value: `{"color":"Red"}`},
				{Field: "attributes.size", Op: OpEq, Value: "Large"},
			},
			wantClause: "attributes @> $1::jsonb AND attributes->>'size' = $2",
			wantArgs:   []any{`{"color":"Red"}`, "Large"},
			wantErr:    false,
		},
		{
			name: "JSONB missing dot error",
			filters: []Filter{
				{Field: "attributes", Op: OpEq, Value: "Red"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clause, args, err := BuildWhereClause(tt.filters, allowed, 1)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantClause, clause)
			assert.Equal(t, tt.wantArgs, args)
		})
	}
}
