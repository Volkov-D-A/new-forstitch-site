package repository

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"new-forstitch-site/backend/internal/models"
)

func TestParsePostDate(t *testing.T) {
	date, err := parsePostDate("2026-06-14")
	if err != nil {
		t.Fatalf("parse valid date: %v", err)
	}
	if !date.Equal(time.Date(2026, 6, 14, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected parsed date: %v", date)
	}

	assertRepositoryErrorCode(t, mustParsePostDateError("14.06.2026"), "date_invalid")
}

func TestRequireAffected(t *testing.T) {
	if err := requireAffected(resultStub{affected: 1}, "missing", "not found"); err != nil {
		t.Fatalf("expected successful result: %v", err)
	}
	assertRepositoryErrorCode(t, requireAffected(resultStub{}, "missing", "not found"), "missing")

	rowsErr := errors.New("rows affected unavailable")
	if err := requireAffected(resultStub{err: rowsErr}, "missing", "not found"); !errors.Is(err, rowsErr) {
		t.Fatalf("expected rows affected error, got %v", err)
	}
}

func TestMapPostgresError(t *testing.T) {
	if err := mapPostgresError(nil); err != nil {
		t.Fatalf("nil error should stay nil: %v", err)
	}

	plain := errors.New("plain error")
	if err := mapPostgresError(plain); !errors.Is(err, plain) {
		t.Fatalf("plain error should pass through: %v", err)
	}

	assertRepositoryErrorCode(t, mapPostgresError(&pgconn.PgError{Code: "23503"}), "reference_not_found")
	assertRepositoryErrorCode(t, mapPostgresError(&pgconn.PgError{Code: "23505"}), "record_exists")

	unknown := &pgconn.PgError{Code: "99999"}
	if err := mapPostgresError(unknown); !errors.Is(err, unknown) {
		t.Fatalf("unknown PostgreSQL error should pass through: %v", err)
	}
}

type resultStub struct {
	affected int64
	err      error
}

func (resultStub) LastInsertId() (int64, error) {
	return 0, errors.New("not supported")
}

func (result resultStub) RowsAffected() (int64, error) {
	return result.affected, result.err
}

var _ sql.Result = resultStub{}

func mustParsePostDateError(value string) error {
	_, err := parsePostDate(value)
	return err
}

func assertRepositoryErrorCode(t *testing.T, err error, code string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error code %q", code)
	}
	var appErr models.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if appErr.Code != code {
		t.Fatalf("expected error code %q, got %q", code, appErr.Code)
	}
}
