package store

import (
	"context"
	"fmt"
	"time"

	"github.com/chariotplatform/goapi/config"
	"github.com/chariotplatform/goapi/logger"

	sql "github.com/chariotplatform/goapi/store/sqlc"

	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database interface {
	Ping(context.Context) error
	Queries() *sql.Queries
	Pool() *pgxpool.Pool
	WithTransaction(ctx context.Context, fn func(q *sql.Queries) error) error
}

func NewDatabase(config config.DatabaseConfig, log logger.Log) (Database, func(), error) {
	ctx := context.Background()

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		config.Host, config.Username, config.Password, config.Database, config.Port, config.SSLMode,
	)
	db, err := NewPG(ctx, dsn, log)

	if err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err = db.Ping(ctx)
	log.Info("Connected to database.")

	cleanup := func() {
		db.Close()
		log.Info("Database connection closed.")
	}

	return db, cleanup, err
}

type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

type Float interface {
	~float32 | ~float64
}

func ToPGText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

func ToPGNullText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func ToPGNullTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}

	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func FromPGNullTimestamptz(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func ToPGInt4(i int32) pgtype.Int4 {
	return pgtype.Int4{Int32: i, Valid: true}
}

func ToPGBool(b bool) pgtype.Bool {
	return pgtype.Bool{Bool: b, Valid: true}
}

func ToPGNullBool(b *bool) pgtype.Bool {
	if b == nil {
		return pgtype.Bool{Valid: false}
	}
	return pgtype.Bool{Valid: true, Bool: *b}
}

func ToPGNullInt4[T Integer](i *T) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{Valid: false}
	}
	num := int32(*i)
	return pgtype.Int4{Int32: num, Valid: true}
}

func ToPGNullInt8[T Integer](i *T) pgtype.Int8 {
	if i == nil {
		return pgtype.Int8{Valid: false}
	}
	num := int64(*i)
	return pgtype.Int8{Int64: num, Valid: true}
}

func ToPGNullFloat8[T Float](f *T) pgtype.Float8 {
	if f == nil {
		return pgtype.Float8{Valid: false}
	}
	return pgtype.Float8{Float64: float64(*f), Valid: true}
}

func ToPGUUID(s uuid.UUID) pgtype.UUID {
	if s == uuid.Nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: s, Valid: true}
}

func ToPGNullUUID(s *uuid.UUID) pgtype.UUID {
	if s == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *s, Valid: true}
}

func FromPGUUID(u pgtype.UUID) uuid.UUID {
	if !u.Valid {
		return uuid.Nil
	}
	return u.Bytes
}

func FromPGNullUUID(u pgtype.UUID) *uuid.UUID {
	if !u.Valid {
		return nil
	}
	res := uuid.UUID(u.Bytes)
	return &res
}

func FromPGNumericToFloat64(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0
	}

	f64, err := n.Float64Value()
	if err != nil {
		return 0
	}

	return f64.Float64
}

func FromPGNumeric(n pgtype.Numeric) float64 {
	return FromPGNumericToFloat64(n)
}

func ToPGNumeric(f float64) pgtype.Numeric {
	var n pgtype.Numeric
	err := n.Scan(fmt.Sprintf("%f", f))
	if err != nil {
		return pgtype.Numeric{Valid: false}
	}
	return n
}

func ToPGNullNumeric(f *float64) pgtype.Numeric {
	if f == nil {
		return pgtype.Numeric{Valid: false}
	}
	var n pgtype.Numeric
	err := n.Scan(fmt.Sprintf("%f", *f))
	if err != nil {
		return pgtype.Numeric{Valid: false}
	}
	return n
}

func ToJSONB(data any) []byte {
	if data == nil {
		return []byte("{}")
	}
	b, err := json.Marshal(data)
	if err != nil {
		return []byte("{}")
	}
	return b
}

func FromJSONB(data []byte) map[string]any {
	var result map[string]any
	if len(data) == 0 {
		return make(map[string]any)
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return make(map[string]any)
	}
	return result
}
