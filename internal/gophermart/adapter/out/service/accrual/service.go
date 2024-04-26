package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/StasMerzlyakov/gophermart/internal/config"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
	"github.com/go-resty/resty/v2"
)

func New(conf *config.GophermartConfig) *acService {
	return &acService{
		client: resty.New().SetBaseURL(conf.AccrualSystemAddress).
			SetRetryCount(3).
			SetRetryWaitTime(200 * time.Millisecond),
	}
}

const pathUrl = "/api/orders/{number}"

type acService struct {
	client *resty.Client
}

func (ac *acService) GetStatus(ctx context.Context, orderNum domain.OrderNumber) (*domain.AccrualData, error) {

	logger, err := domain.GetCtxLogger(ctx)
	if err != nil {
		fmt.Printf("can't get logger from context")
		return nil, err
	}

	req := ac.client.R()

	req.SetPathParam("number", string(orderNum))

	resp, err := req.Get(pathUrl)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrServerInternal, err.Error())
	}

	if resp.StatusCode() == http.StatusOK {
		var accrualData domain.AccrualData
		if err = json.Unmarshal(resp.Body(), &accrualData); err != nil {
			return nil, fmt.Errorf("%w: %v", domain.ErrServerInternal, err.Error())
		}
		return &accrualData, nil
	}

	if resp.StatusCode() == http.StatusNoContent {
		logger.Infow("acrual.GetStatus", "err", fmt.Sprintf("order %v not found", orderNum))
		return &domain.AccrualData{
			Number: orderNum,
			Status: domain.AccrualStatusInvalid,
		}, nil
	}

	if resp.StatusCode() == http.StatusTooManyRequests {
		logger.Infow("acrual.GetStatus", "err", "too many requests")
		return nil, fmt.Errorf("%w: too many requests", domain.ErrServerInternal)
	}

	logger.Infow("acrual.GetStatus", "err", fmt.Sprintf("unexpected status code %v", resp.StatusCode()))
	return nil, fmt.Errorf("%w: unexpected status code %v", domain.ErrServerInternal, resp.StatusCode())
}
