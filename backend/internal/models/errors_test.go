package models

import (
	"errors"
	"testing"
)

func TestAppErrorKindsAndPayloads(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		kind    error
		code    string
		message string
	}{
		{name: "bad request", err: BadRequest("bad", "bad request"), kind: ErrBadRequest, code: "bad", message: "bad request"},
		{name: "conflict", err: Conflict("exists", "already exists"), kind: ErrConflict, code: "exists", message: "already exists"},
		{name: "internal", err: Internal("internal", "internal error"), kind: ErrInternal, code: "internal", message: "internal error"},
		{name: "not found", err: NotFound("missing", "not found"), kind: ErrNotFound, code: "missing", message: "not found"},
		{name: "unauthorized", err: Unauthorized("auth", "unauthorized"), kind: ErrUnauthorized, code: "auth", message: "unauthorized"},
		{name: "validation", err: Validation("invalid", "invalid value"), kind: ErrValidation, code: "invalid", message: "invalid value"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !errors.Is(test.err, test.kind) {
				t.Fatalf("expected error kind %v, got %v", test.kind, test.err)
			}
			if test.err.Error() != test.message {
				t.Fatalf("expected message %q, got %q", test.message, test.err.Error())
			}
			payload := ErrorPayloadFrom(test.err, "fallback", "fallback message")
			if payload.Code != test.code || payload.Message != test.message {
				t.Fatalf("unexpected payload: %+v", payload)
			}
		})
	}
}

func TestErrorPayloadFromFallsBackForUnknownError(t *testing.T) {
	payload := ErrorPayloadFrom(errors.New("database unavailable"), "internal_error", "internal server error")
	if payload.Code != "internal_error" || payload.Message != "internal server error" {
		t.Fatalf("unexpected fallback payload: %+v", payload)
	}
}
