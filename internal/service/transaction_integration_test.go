package service_test

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
	"entaintest/internal/repository"
	"entaintest/internal/service"
	servicerequest "entaintest/internal/service/model/request"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
)

// TransactionIntegrationSuite drives the real services and repositories against
// a Postgres. It is skipped unless TEST_DATABASE_URL is set.
type TransactionIntegrationSuite struct {
	suite.Suite

	pool           *pgxpool.Pool
	service        *service.TransactionService
	balanceService *service.BalanceService
	idSeq          int64
}

func Test_TransactionIntegrationSuite(t *testing.T) {
	suite.Run(t, new(TransactionIntegrationSuite))
}

func (s *TransactionIntegrationSuite) SetupSuite() {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		return
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	s.Require().NoError(err)
	s.Require().NoError(pool.Ping(context.Background()))

	users := repository.NewUserRepository(pool)

	s.pool = pool
	s.service = service.NewTransactionService(
		repository.NewTransactor(pool),
		users,
		repository.NewTransactionRepository(pool),
	)
	s.balanceService = service.NewBalanceService(users)
	// Small, int32-safe base that differs per run to avoid id collisions.
	s.idSeq = (time.Now().Unix() % 1_000_000) * 1000
}

func (s *TransactionIntegrationSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
}

func (s *TransactionIntegrationSuite) SetupTest() {
	if s.pool == nil {
		s.T().Skip("TEST_DATABASE_URL not set; skipping integration test")
	}
}

func (s *TransactionIntegrationSuite) seedUser(balance int64) uint64 {
	ctx := context.Background()
	id := uint64(atomic.AddInt64(&s.idSeq, 1))

	_, err := s.pool.Exec(
		ctx,
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

func (s *TransactionIntegrationSuite) balanceOf(userID uint64) int64 {
	response, err := s.balanceService.GetBalance(context.Background(), userID)
	s.Require().NoError(err)

	return response.Balance
}

func (s *TransactionIntegrationSuite) request(userID uint64, state, amount, transactionID string) servicerequest.ProcessTransactionRequest {
	return servicerequest.ProcessTransactionRequest{
		UserID:        userID,
		SourceType:    "game",
		State:         state,
		Amount:        amount,
		TransactionID: transactionID,
	}
}

func (s *TransactionIntegrationSuite) Test_ProcessTransaction_ReturnsStartingBalance_InCaseRunConcurrently() {
	// Arrange
	const pairs = 200
	start := int64(pairs) * 100 // high enough that no lose ever overdraws
	userID := s.seedUser(start)

	var wg sync.WaitGroup
	errCh := make(chan error, pairs*2)

	apply := func(i int, state string) {
		defer wg.Done()

		_, err := s.service.ProcessTransaction(
			context.Background(),
			s.request(userID, state, "1.00", fmt.Sprintf("conc-%d-%s-%d", userID, state, i)),
		)
		if err != nil {
			errCh <- err
		}
	}

	// Act
	for i := 0; i < pairs; i++ {
		wg.Add(2)
		go apply(i, "win")
		go apply(i, "lose")
	}
	wg.Wait()
	close(errCh)

	// Assert
	for err := range errCh {
		s.Require().NoError(err)
	}
	s.Require().Equal(start, s.balanceOf(userID))
}

func (s *TransactionIntegrationSuite) Test_ProcessTransaction_AppliesOnce_InCaseTransactionIDIsDuplicated() {
	// Arrange
	const goroutines = 50
	userID := s.seedUser(0)
	transactionID := fmt.Sprintf("dup-%d", userID)

	var (
		wg         sync.WaitGroup
		successes  int64
		duplicates int64
	)

	apply := func() {
		defer wg.Done()

		_, err := s.service.ProcessTransaction(context.Background(), s.request(userID, "win", "1.00", transactionID))

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
	s.Require().Equal(int64(100), s.balanceOf(userID))
}

func (s *TransactionIntegrationSuite) Test_ProcessTransaction_ReturnsInsufficientFunds_InCaseBalanceTooLow() {
	// Arrange
	userID := s.seedUser(50)

	// Act
	_, err := s.service.ProcessTransaction(context.Background(), s.request(userID, "lose", "1.00", fmt.Sprintf("insuff-%d", userID)))

	// Assert
	s.Require().ErrorIs(err, businesserror.ErrInsufficientFunds)
	s.Require().Equal(int64(50), s.balanceOf(userID))
}

func (s *TransactionIntegrationSuite) Test_ProcessTransaction_ReturnsUserNotFound_InCaseUserMissing() {
	// Arrange
	missingID := uint64(atomic.AddInt64(&s.idSeq, 1))
	transactionID := fmt.Sprintf("missing-%d", missingID)

	// Act
	_, err := s.service.ProcessTransaction(context.Background(), s.request(missingID, "win", "1.00", transactionID))

	// Assert
	s.Require().ErrorIs(err, businesserror.ErrUserNotFound)

	var count int
	countErr := s.pool.QueryRow(
		context.Background(),
		`SELECT count(*) FROM transactions WHERE transaction_id = @transaction_id`,
		pgx.NamedArgs{"transaction_id": transactionID},
	).Scan(&count)
	s.Require().NoError(countErr)
	s.Require().Zero(count)
}
