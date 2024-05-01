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

const pathURL = "/api/orders/{number}"

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

	resp, err := req.Get(pathURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrServerInternal, err.Error())
	}

	statusCode := resp.StatusCode()
	switch statusCode {
	case http.StatusOK:
		var accrualData domain.AccrualData
		if err = json.Unmarshal(resp.Body(), &accrualData); err != nil {
			return nil, fmt.Errorf("%w: %v", domain.ErrServerInternal, err.Error())
		}
		logger.Infow("acrual.GetStatus", "msg", fmt.Sprintf("order %v found", orderNum))
		return &accrualData, nil
	case http.StatusNoContent:
		logger.Infow("acrual.GetStatus", "msg", fmt.Sprintf("order %v not found", orderNum))
		return nil, nil
	case http.StatusTooManyRequests:
		logger.Errorw("acrual.GetStatus", "err", "too many requests")
		return nil, fmt.Errorf("%w: too many requests", domain.ErrServerInternal)
	default:
		errMsg := fmt.Sprintf("unexpected status code %v", statusCode)
		logger.Errorw("acrual.GetStatus", "err", errMsg)
		return nil, fmt.Errorf("%w: %v", domain.ErrServerInternal, errMsg)
	}

}
