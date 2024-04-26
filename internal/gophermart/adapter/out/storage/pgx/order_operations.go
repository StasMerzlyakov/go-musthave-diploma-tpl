package pgx

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
	"github.com/jackc/pgx/v5"
)

func (st *storage) Upload(ctx context.Context, data *domain.OrderData) error {

	logger, err := domain.GetCtxLogger(ctx)
	if err != nil {
		fmt.Printf("storage.Upload error: can't extract logger")
		return err
	}

	if data == nil {
		logger.Errorw("storage.Upload", "err", "data is nil")
		return domain.ErrServerInternal
	}

	var number domain.OrderNumber

	if err := st.pPool.QueryRow(ctx,
		`insert into orderData(number, userID, status, uploaded_at) values ($1, $2, $3, $4) 
		on conflict("number") do nothing returning number;
	  `, data.Number, data.UserID, data.Status, time.Time(data.UploadedAt).UTC()).Scan(&number); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// запись с таким number уже есть; проверим какому пользователю принадлежит
			var userID int
			err = st.pPool.QueryRow(ctx, `select userID from orderData where number = $1`, data.Number).Scan(&userID)
			if err != nil {
				logger.Errorw("storage.Upload", "err", err.Error())
				return domain.ErrServerInternal
			}
			if userID == data.UserID {
				return domain.ErrOrderNumberAlreadyUploaded
			} else {
				return domain.ErrDublicateOrderNumber
			}
		} else {
			logger.Errorw("storage.Upload", "err", err.Error())
			return domain.ErrServerInternal
		}
	}

	return nil
}

func (st *storage) Orders(ctx context.Context, userID int) ([]domain.OrderData, error) {
	var orders []domain.OrderData

	logger, err := domain.GetCtxLogger(ctx)
	if err != nil {
		fmt.Printf("storage.Orders error: can't extract logger")
		return nil, err
	}

	rows, err := st.pPool.Query(ctx,
		`select number, userID, status, accrual, uploaded_at from orderData where userID = $1`,
		userID,
	)

	if err != nil {
		logger.Errorw("storage.Orders", "err", err.Error())
		return nil, domain.ErrServerInternal
	}

	defer rows.Close()

	for rows.Next() {
		var data domain.OrderData
		var uploaded time.Time
		err = rows.Scan(&data.Number, &data.UserID, &data.Status, &data.Accrual, &uploaded)
		if err != nil {
			logger.Errorw("storage.Orders", "err", err.Error())
			return nil, domain.ErrServerInternal
		}
		data.UploadedAt = domain.RFC3339Time(uploaded)
		orders = append(orders, data)
	}

	err = rows.Err()
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Infow("storage.Orders", "status", "not found")
			return nil, domain.ErrNotFound
		}
		logger.Errorw("storage.Orders", "err", err.Error())
		return nil, domain.ErrServerInternal
	}

	return orders, nil
}

func (st *storage) UpdateOrder(ctx context.Context, number domain.OrderNumber, status domain.OrderStatus, accrual *float64) error {

	logger, err := domain.GetCtxLogger(ctx)
	if err != nil {
		fmt.Printf("storage.UpdateOrder error: can't extract logger")
		return err
	}

	rows, err := st.pPool.Query(ctx,
		`update orderData set status = $1, accrual = $2 where number = $3`,
		string(status), accrual, string(number),
	)

	if err != nil {
		logger.Errorw("storage.UpdateOrder", "err", err.Error())
		return domain.ErrServerInternal
	}

	defer rows.Close()

	if rows.Next() {
		return nil
	}

	err = rows.Err()
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Infow("storage.UpdateOrder", "status", "not found")
			return domain.ErrNotFound
		}
		return fmt.Errorf("%w: %v", domain.ErrServerInternal, err.Error())
	}
	return nil
}

func (st *storage) GetByStatus(ctx context.Context, status domain.OrderStatus) ([]domain.OrderData, error) {

	logger, err := domain.GetCtxLogger(ctx)
	if err != nil {
		fmt.Printf("storage.GetByStatus error: can't extract logger")
		return nil, err
	}

	var forProcessing []domain.OrderData

	rows, err := st.pPool.Query(ctx,
		`update orderData set score = $1 
		 where number in 
		   (select number from orderdata where status = $2 and score < $3 limit $4) 
		 returning 
		    number, userID, status, accrual, uploaded_at;`,
		time.Now().Add(st.processingScoreDelta),
		string(status),
		time.Now(),
		st.processingLimit,
	)

	if err != nil {
		logger.Infow("storage.ForProcessing", "err", err.Error())
		return nil, domain.ErrServerInternal
	}

	defer rows.Close()

	for rows.Next() {
		var data domain.OrderData
		var uploaded time.Time
		err = rows.Scan(&data.Number, &data.UserID, &data.Status, &data.Accrual, &uploaded)
		if err != nil {
			logger.Infow("storage.ForProcessing", "err", err.Error())
			return nil, domain.ErrServerInternal
		}
		data.UploadedAt = domain.RFC3339Time(uploaded)
		forProcessing = append(forProcessing, data)
	}

	err = rows.Err()
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		logger.Infow("storage.ForProcessing", "err", err.Error())
		return nil, domain.ErrServerInternal
	}
	return forProcessing, nil
}

func (st *storage) UpdateBatch(ctx context.Context, orders []domain.OrderData) error {

	logger, err := domain.GetCtxLogger(ctx)
	if err != nil {
		fmt.Printf("storage.UpdateBatch error: can't extract logger")
		return err
	}

	tx, err := st.pPool.Begin(ctx)
	if err != nil {
		logger.Errorw("storage.UpdateBatch", "err", err.Error())
		return domain.ErrServerInternal
	}

	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}

	for _, ord := range orders {
		batch.QueuedQueries = append(batch.QueuedQueries,
			&pgx.QueuedQuery{
				SQL:       `update orderData set status = $1, accrual = $2 where number = $3`,
				Arguments: []any{string(ord.Status), ord.Accrual, string(ord.Number)},
			},
		)
	}

	err = tx.SendBatch(context.Background(), batch).Close()

	if err != nil {
		logger.Errorw("storage.UpdateBatch", "err", err.Error())
		return domain.ErrServerInternal
	}

	if err = tx.Commit(ctx); err != nil {
		logger.Errorw("storage.UpdateBatch", "err", err.Error())
		return domain.ErrServerInternal
	}

	return nil
}
