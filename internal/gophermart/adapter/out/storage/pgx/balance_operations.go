package pgx

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
	"github.com/jackc/pgx/v5"
)

func (st *storage) Balance(ctx context.Context, userID int) (*domain.UserBalance, error) {

	logger, err := domain.GetCtxLogger(ctx)
	if err != nil {
		fmt.Printf("storage.Balance error: can't extract logger")
		return nil, err
	}

	logger.Infow("storage.Balance", "status", "start")

	var userBalance domain.UserBalance

	if err := st.pPool.QueryRow(ctx,
		` with ins as (
			insert into balance(userID) values ($1) on conflict (userID) do nothing 
			-- явно укащываб список полей, для исключения возможных ошибок
			 returning balanceID, userID, current, withdrawn, release 
			)
			select balanceID, userID, current, withdrawn, release from ins
			union
			select balanceID, userID, current, withdrawn, release from balance
			where userID=$1;`,
		userID).Scan(&userBalance.BalanceID, &userBalance.UserID, &userBalance.Current, &userBalance.Release, &userBalance.Release); err == nil {
		logger.Infow("storage.Balance", "status", "success")
		return &userBalance, nil
	} else {
		logger.Errorw("storage.Balance", "err", err.Error())
		return nil, domain.ErrServerInternal
	}
}

func (st *storage) UpdateBalanceByOrder(ctx context.Context, balance *domain.UserBalance, orderData *domain.OrderData) error {

	logger, err := domain.GetCtxLogger(ctx)
	if err != nil {
		fmt.Printf("storage.UpdateBalanceByOrder error: can't extract logger")
		return err
	}

	if balance == nil {
		logger.Errorw("storage.UpdateBalanceByOrder", "err", "balance is nil")
		return fmt.Errorf("%w: balance is nil", domain.ErrServerInternal)
	}

	if orderData == nil {
		logger.Errorw("storage.UpdateBalanceByOrder", "err", "orderData is nil")
		return fmt.Errorf("%w: orderData is nil", domain.ErrServerInternal)
	}

	tx, err := st.pPool.Begin(ctx)

	if err != nil {
		logger.Errorw("storage.UpdateBalanceByOrder", "err", err.Error())
		return domain.ErrServerInternal
	}

	defer tx.Rollback(ctx)

	var orderNum string
	err = tx.QueryRow(ctx,
		`update orderData set status = $1 where number = $2 and status<> $1 returning number`,
		orderData.Status,
		orderData.Number,
	).Scan(&orderNum)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Infow("storage.UpdateBalanceByOrder", "status", "not found")
			return domain.ErrNotFound
		}
		logger.Errorw("storage.UpdateBalanceByOrder", "err", err.Error())
		return domain.ErrServerInternal
	}

	var balanceID int
	err = tx.QueryRow(ctx,
		`update balance set current = $1, withdrawn = $2 ,release = release+1 where userID=$3 and release=$4 returning balanceID`,
		balance.Current,
		balance.Withdrawn,
		balance.UserID,
		balance.Release).Scan(&balanceID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Infow("storage.UpdateBalanceByOrder", "status", "not found")
			return domain.ErrBalanceChanged
		}
		logger.Errorw("storage.UpdateBalanceByOrder", "err", err.Error())
		return domain.ErrServerInternal
	}

	if err = tx.Commit(ctx); err != nil {
		logger.Errorw("storage.UpdateBalanceByOrder", "err", err.Error())
		return domain.ErrServerInternal
	}

	return nil
}

func (st *storage) UpdateBalanceByWithdraw(ctx context.Context, balance *domain.UserBalance, withdraw *domain.WithdrawalData) error {
	logger, err := domain.GetCtxLogger(ctx)
	if err != nil {
		fmt.Printf("storage.UpdateBalanceByWithdraw error: can't extract logger")
		return err
	}

	if balance == nil {
		logger.Errorw("storage.UpdateBalanceByWithdraw", "err", "balance is nil")
		return fmt.Errorf("%w: balance is nil", domain.ErrServerInternal)
	}

	if withdraw == nil {
		logger.Errorw("storage.UpdateBalanceByWithdraw", "err", "withdraw is nil")
		return fmt.Errorf("%w: withdraw is nil", domain.ErrServerInternal)
	}

	tx, err := st.pPool.Begin(ctx)

	if err != nil {
		logger.Errorw("storage.UpdateBatch", "err", err.Error())
		return domain.ErrServerInternal
	}

	defer tx.Rollback(ctx)

	tx.Exec(ctx,
		`insert into withdrawal(balancerId, number, sum, processed_at) values($1, $2, $3, $4)`,
		balance.BalanceID,
		withdraw.Order,
		withdraw.Sum,
		time.Time(withdraw.ProcessedAt),
	)

	var balanceID int
	err = tx.QueryRow(ctx,
		`update balance set current = $1, withdrawn = $2 ,release = release+1 where userID=$3 and release=$4 returning balanceID`,
		balance.Current,
		balance.Withdrawn,
		balance.UserID,
		balance.Release).Scan(&balanceID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Infow("storage.UpdateBalanceByWithdraw", "status", "not found")
			return domain.ErrBalanceChanged
		}
		logger.Errorw("storage.UpdateBalanceByWithdraw", "err", err.Error())
		return domain.ErrServerInternal
	}

	if err = tx.Commit(ctx); err != nil {
		logger.Errorw("storage.UpdateBalanceByWithdraw", "err", err.Error())
		return domain.ErrServerInternal
	}

	return nil
}

func (st *storage) Withdrawals(ctx context.Context, userID int) ([]domain.WithdrawalData, error) {
	var withdrawals []domain.WithdrawalData

	logger, err := domain.GetCtxLogger(ctx)
	if err != nil {
		fmt.Printf("storage.UpdateBalanceByWithdraw error: can't extract logger")
		return nil, err
	}

	rows, err := st.pPool.Query(ctx,
		`select w.number, w.sum, w.processed_at from withdrawal w 
		inner join balance b on b.balanceID = w.balanceID where b.userID=$1`,
		userID,
	)

	if err != nil {
		logger.Infow("storage.Withdrawals", "err", err.Error())
		return nil, domain.ErrServerInternal
	}

	defer rows.Close()

	for rows.Next() {
		var withdrawal domain.WithdrawalData
		var processedAt time.Time
		err = rows.Scan(&withdrawal.Order, &withdrawal.Sum, &processedAt)
		if err != nil {
			logger.Infow("storage.Withdrawals", "err", err.Error())
			return nil, domain.ErrServerInternal
		}
		withdrawal.ProcessedAt = domain.RFC3339Time(processedAt)
		withdrawals = append(withdrawals, withdrawal)
	}

	err = rows.Err()
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Infow("storage.Withdrawals", "status", "not found")
			return nil, domain.ErrNotFound
		}
		logger.Errorw("storage.Withdrawals", "err", err.Error())
		return nil, domain.ErrServerInternal
	}

	return withdrawals, nil
}
