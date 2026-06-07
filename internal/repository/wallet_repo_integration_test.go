package repository_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"entaintest/internal/common/businesserror"
	"entaintest/internal/common/enum"
	"entaintest/internal/model"
	"entaintest/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
)

// WalletRepositorySuite runs against a real Postgres. It is skipped unless
// TEST_DATABASE_URL is set, so the default `go test ./...` stays DB-free.
type WalletRepositorySuite struct {
	suite.Suite

	pool  *pgxpool.Pool
	repo  *repository.WalletRepository
	idSeq int64
}

func Test_WalletRepositorySuite(t *testing.T) {
	suite.Run(t, new(WalletRepositorySuite))
}

func (s *WalletRepositorySuite) SetupSuite() {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		return
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	s.Require().NoError(err)
	s.Require().NoError(pool.Ping(context.Background()))

	s.pool = pool
	s.repo = repository.NewWalletRepository(pool)
	s.idSeq = time.Now().UnixNano()
}

func (s *WalletRepositorySuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
}

func (s *WalletRepositorySuite) SetupTest() {
	if s.pool == nil {
		s.T().Skip("TEST_DATABASE_URL not set; skipping integration test")
	}
}

func (s *WalletRepositorySuite) seedUser(balance int64) uint64 {
	ctx := context.Background()
	id := uint64(atomic.AddInt64(&s.idSeq, 1))

	_, err := s.pool.Exec(ctx,
		`INSERT INTO users (id, balance) VALUES (@id, @balance)`,
		pgx.NamedArgs{"id": id, "balance": balance},
	)
	s.Require().NoError(err)

	s.T().Cleanup(func() {
		_, _ = s.pool.Exec(ctx, `DELETE FROM transactions WHERE user_id = @id`, pgx.NamedArgs{"id": id})
		_, _ = s.pool.Exec(ctx, `DELETE FROM users WHERE id = @id`, pgx.NamedArgs{"id": id})
	})

	return id
}

func (s *WalletRepositorySuite) Test_ApplyTransaction_ReturnsStartingBalance_WhenRunConcurrently() {
	// Arrange
	const (
		pairs  = 200
		amount = int64(100)
		start  = int64(pairs) * amount // high enough that no lose ever overdraws
	)
	userID := s.seedUser(start)

	var wg sync.WaitGroup
	errCh := make(chan error, pairs*2)

	apply := func(i int, state enum.State) {
		defer wg.Done()

		_, err := s.repo.ApplyTransaction(context.Background(), model.Transaction{
			TransactionID: fmt.Sprintf("conc-%d-%s-%d", userID, state, i),
			UserID:        userID,
			State:         state,
			SourceType:    enum.SourceGame,
			Amount:        amount,
		})
		if err != nil {
			errCh <- err
		}
	}

	// Act
	for i := 0; i < pairs; i++ {
		wg.Add(2)
		go apply(i, enum.StateWin)
		go apply(i, enum.StateLose)
	}
	wg.Wait()
	close(errCh)

	// Assert
	for err := range errCh {
		s.Require().NoError(err)
	}

	balance, err := s.repo.GetBalance(context.Background(), userID)
	s.Require().NoError(err)
	s.Require().Equal(start, balance)
}

func (s *WalletRepositorySuite) Test_ApplyTransaction_AppliesOnce_WhenTransactionIDIsDuplicated() {
	// Arrange
	const goroutines = 50
	userID := s.seedUser(0)
	txID := fmt.Sprintf("dup-%d", userID)

	var (
		wg         sync.WaitGroup
		successes  int64
		duplicates int64
	)

	apply := func() {
		defer wg.Done()

		_, err := s.repo.ApplyTransaction(context.Background(), model.Transaction{
			TransactionID: txID,
			UserID:        userID,
			State:         enum.StateWin,
			SourceType:    enum.SourceGame,
			Amount:        100,
		})

		switch {
		case err == nil:
			atomic.AddInt64(&successes, 1)
		case errors.Is(err, businesserror.ErrDuplicateTransaction):
			atomic.AddInt64(&duplicates, 1)
		default:
			s.Require().NoError(err)
		}
	}

	// Act
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go apply()
	}
	wg.Wait()

	// Assert
	s.Require().Equal(int64(1), successes)
	s.Require().Equal(int64(goroutines-1), duplicates)

	balance, err := s.repo.GetBalance(context.Background(), userID)
	s.Require().NoError(err)
	s.Require().Equal(int64(100), balance)
}

func (s *WalletRepositorySuite) Test_ApplyTransaction_ReturnsInsufficientFunds_WhenBalanceTooLow() {
	// Arrange
	userID := s.seedUser(50)

	// Act
	_, err := s.repo.ApplyTransaction(context.Background(), model.Transaction{
		TransactionID: fmt.Sprintf("insuff-%d", userID),
		UserID:        userID,
		State:         enum.StateLose,
		SourceType:    enum.SourcePayment,
		Amount:        100,
	})

	// Assert
	s.Require().ErrorIs(err, businesserror.ErrInsufficientFunds)

	balance, balErr := s.repo.GetBalance(context.Background(), userID)
	s.Require().NoError(balErr)
	s.Require().Equal(int64(50), balance)
}

func (s *WalletRepositorySuite) Test_ApplyTransaction_ReturnsUserNotFound_WhenUserMissing() {
	// Arrange
	missingID := uint64(atomic.AddInt64(&s.idSeq, 1))
	txID := fmt.Sprintf("missing-%d", missingID)

	// Act
	_, err := s.repo.ApplyTransaction(context.Background(), model.Transaction{
		TransactionID: txID,
		UserID:        missingID,
		State:         enum.StateWin,
		SourceType:    enum.SourceServer,
		Amount:        100,
	})

	// Assert
	s.Require().ErrorIs(err, businesserror.ErrUserNotFound)

	var count int
	countErr := s.pool.QueryRow(context.Background(),
		`SELECT count(*) FROM transactions WHERE transaction_id = @transaction_id`,
		pgx.NamedArgs{"transaction_id": txID},
	).Scan(&count)
	s.Require().NoError(countErr)
	s.Require().Zero(count)
}
