package transport

import (
	"fmt"
	"strings"

	"github.com/chariotplatform/goapi/internal/shared/id"
)

// FieldType defines how a filterable field maps to SQL.
type FieldType int

const (
	FieldString  FieldType = iota // VARCHAR / TEXT columns
	FieldNumeric                  // DECIMAL / INT columns
	FieldUUID                     // UUID columns
	FieldJSONB                    // JSONB key access (attributes->>'key')
)

// AllowedField describes a single whitelisted filterable/sortable field.
type AllowedField struct {
	Column   string    // SQL column name (e.g. "price", "status")
	Type     FieldType // determines casting / operator validation
	JSONBCol string    // if Type == FieldJSONB, the parent JSONB column (e.g. "attributes")
}

// opSQL maps FilterOp values to SQL operators.
var opSQL = map[FilterOp]string{
	OpEq:       "=",
	OpNeq:      "!=",
	OpGt:       ">",
	OpGte:      ">=",
	OpLt:       "<",
	OpLte:      "<=",
	OpLike:     "ILIKE",
	OpContains: "@>",
}

// BuildWhereClause creates a parameterized WHERE clause from filters.
// startParam is the starting $N index (e.g. 1 if no prior params).
// Returns the clause (without "WHERE"), the parameter values, and any error.
func BuildWhereClause(filters []Filter, allowed map[string]AllowedField, startParam int) (string, []any, error) {
	if len(filters) == 0 {
		return "", nil, nil
	}

	clauses := make([]string, 0, len(filters))
	args := make([]any, 0, len(filters))
	paramIdx := startParam

	for _, f := range filters {
		field, ok := allowed[f.Field]
		if !ok {
			return "", nil, BadRequest(
				fmt.Sprintf("unknown filter field: %q", f.Field),
				"filter_validation", nil,
			)
		}

		sqlOp, ok := opSQL[f.Op]
		if !ok {
			return "", nil, BadRequest(
				fmt.Sprintf("unsupported operator: %q", f.Op),
				"filter_validation", nil,
			)
		}

		var clause string
		var arg any

		switch field.Type {
		case FieldJSONB:
			if f.Op == OpContains {
				clause = fmt.Sprintf("%s %s $%d::jsonb", field.JSONBCol, sqlOp, paramIdx)
				arg = f.Value
			} else {
				// Extract the JSONB key from the field name: "attributes.color" → key = "color"
				parts := strings.SplitN(f.Field, ".", 2)
				if len(parts) != 2 {
					return "", nil, BadRequest(
						fmt.Sprintf("JSONB filter must be in format column.key, got: %q", f.Field),
						"filter_validation", nil,
					)
				}
				key := parts[1]

				if f.Op == OpLike {
					clause = fmt.Sprintf("%s->>'%s' ILIKE $%d", field.JSONBCol, key, paramIdx)
					arg = "%" + f.Value + "%"
				} else {
					clause = fmt.Sprintf("%s->>'%s' %s $%d", field.JSONBCol, key, sqlOp, paramIdx)
					arg = f.Value
				}
			}

		case FieldNumeric:
			if f.Op == OpLike {
				return "", nil, BadRequest(
					fmt.Sprintf("operator 'like' is not supported on numeric field %q", f.Field),
					"filter_validation", nil,
				)
			}
			clause = fmt.Sprintf("%s %s $%d::numeric", field.Column, sqlOp, paramIdx)
			arg = f.Value

		case FieldUUID:
			if f.Op != OpEq && f.Op != OpNeq {
				return "", nil, BadRequest(
					fmt.Sprintf("UUID field %q only supports eq/neq operators", f.Field),
					"filter_validation", nil,
				)
			}

			clause = fmt.Sprintf("%s %s $%d::uuid", field.Column, sqlOp, paramIdx)
			parsedUUID, err := id.DecodeAnyPrefix(f.Value)
			if err != nil {
				return "", nil, BadRequest(
					fmt.Sprintf("invalid id value: %q", f.Value),
					"filter_validation", nil,
				)
			}
			arg = parsedUUID

		default: // FieldString
			if f.Op == OpLike {
				clause = fmt.Sprintf("%s ILIKE $%d", field.Column, paramIdx)
				arg = "%" + f.Value + "%"
			} else {
				clause = fmt.Sprintf("%s %s $%d", field.Column, sqlOp, paramIdx)
				arg = f.Value
			}
		}

		clauses = append(clauses, clause)
		args = append(args, arg)
		paramIdx++
	}

	return strings.Join(clauses, " AND "), args, nil
}

// BuildOrderClause returns a safe ORDER BY clause from a Sort directive.
// Returns empty string if sort is nil or field is not allowed.
func BuildOrderClause(sort *Sort, allowed map[string]AllowedField) string {
	if sort == nil {
		return ""
	}

	field, ok := allowed[sort.Field]
	if !ok {
		return ""
	}

	dir := "ASC"
	if sort.Dir == SortDesc {
		dir = "DESC"
	}

	// For JSONB fields, sort by the extracted key
	if field.Type == FieldJSONB {
		parts := strings.SplitN(sort.Field, ".", 2)
		if len(parts) == 2 {
			return fmt.Sprintf("%s->>'%s' %s", field.JSONBCol, parts[1], dir)
		}
	}

	return fmt.Sprintf("%s %s", field.Column, dir)
}
