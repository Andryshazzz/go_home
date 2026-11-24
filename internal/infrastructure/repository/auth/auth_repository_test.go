package authrepository

import (
	"context"
	"errors"
	"testing"
	"time"

	"go_home/internal/domain/entity"
	apperrors "go_home/pkg/errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type registrationStore struct {
	users    map[uuid.UUID]bool
	sessions map[uuid.UUID]bool
}

type registrationTx struct {
	pgx.Tx
	store       *registrationStore
	userID      uuid.UUID
	sessionID   uuid.UUID
	sessionErr  error
	userErr     error
	commitErr   error
	rollbackErr error
	closed      bool
}

func (tx *registrationTx) Exec(_ context.Context, _ string, args ...any) (pgconn.CommandTag, error) {
	if len(args) == 6 {
		if tx.userErr != nil {
			return pgconn.CommandTag{}, tx.userErr
		}
		tx.userID = args[0].(uuid.UUID)
		return pgconn.NewCommandTag("INSERT 0 1"), nil
	}
	if tx.sessionErr != nil {
		return pgconn.CommandTag{}, tx.sessionErr
	}
	tx.sessionID = args[0].(uuid.UUID)
	return pgconn.NewCommandTag("INSERT 0 1"), nil
}

func (tx *registrationTx) Commit(context.Context) error {
	if tx.commitErr != nil {
		return tx.commitErr
	}
	tx.store.users[tx.userID] = true
	tx.store.sessions[tx.sessionID] = true
	tx.closed = true
	return nil
}

func (tx *registrationTx) Rollback(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if tx.closed {
		return pgx.ErrTxClosed
	}
	tx.closed = true
	return tx.rollbackErr
}

func TestCreateUserWithSessionRollsBackAndCanRetry(t *testing.T) {
	store := &registrationStore{users: make(map[uuid.UUID]bool), sessions: make(map[uuid.UUID]bool)}
	user := entity.User{ID: uuid.New()}
	session := entity.Session{ID: uuid.New(), UserID: user.ID}
	sessionErr := errors.New("session insert failed")
	tx := &registrationTx{store: store, sessionErr: sessionErr}

	err := createUserWithSession(context.Background(), tx, user, session, time.Second)
	if !errors.Is(err, sessionErr) || !errors.Is(err, apperrors.ErrInternal) {
		t.Fatalf("expected session failure, got %v", err)
	}
	if !tx.closed || len(store.users) != 0 || len(store.sessions) != 0 {
		t.Fatal("failed registration was not rolled back")
	}

	retry := &registrationTx{store: store}
	if err := createUserWithSession(context.Background(), retry, user, session, time.Second); err != nil {
		t.Fatalf("retry failed: %v", err)
	}
	if !store.users[user.ID] || !store.sessions[session.ID] {
		t.Fatal("retry did not commit user and session")
	}
}

func TestCreateUserWithSessionErrors(t *testing.T) {
	commitErr := errors.New("commit failed")
	rollbackErr := errors.New("rollback failed")
	tests := []struct {
		name        string
		userErr     error
		commitErr   error
		rollbackErr error
		want        error
	}{
		{name: "duplicate user", userErr: &pgconn.PgError{Code: uniqueViolation}, want: apperrors.ErrUserAlreadyExists},
		{name: "commit failure", commitErr: commitErr, want: commitErr},
		{name: "rollback failure", userErr: &pgconn.PgError{Code: uniqueViolation}, rollbackErr: rollbackErr, want: apperrors.ErrInternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &registrationStore{users: make(map[uuid.UUID]bool), sessions: make(map[uuid.UUID]bool)}
			tx := &registrationTx{store: store, userErr: tt.userErr, commitErr: tt.commitErr, rollbackErr: tt.rollbackErr}
			err := createUserWithSession(context.Background(), tx, entity.User{ID: uuid.New()}, entity.Session{ID: uuid.New()}, time.Second)
			if !errors.Is(err, tt.want) || !tx.closed || len(store.users) != 0 || len(store.sessions) != 0 {
				t.Fatalf("unexpected result: err=%v, transaction closed=%v", err, tx.closed)
			}
		})
	}
}

func TestCreateUserWithSessionRollsBackAfterCancellation(t *testing.T) {
	store := &registrationStore{users: make(map[uuid.UUID]bool), sessions: make(map[uuid.UUID]bool)}
	tx := &registrationTx{store: store, sessionErr: context.Canceled}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := createUserWithSession(ctx, tx, entity.User{ID: uuid.New()}, entity.Session{ID: uuid.New()}, time.Second)
	if !errors.Is(err, context.Canceled) || !tx.closed {
		t.Fatalf("expected rollback after cancellation, got err=%v closed=%v", err, tx.closed)
	}
}
