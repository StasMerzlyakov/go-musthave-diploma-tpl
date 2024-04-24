package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type RetriableInvokerConf struct {
	RetriableErr    error
	FirstRetryDelay time.Duration
	DelayIncrement  time.Duration
	RetryCount      int
	PreProccFn      ErrPreProcessFn
}

type RetriableInvoker interface {
	Invoke(ctx context.Context, fn InvokableFn) error
}

type InvokableFn func(ctx context.Context) error

type ErrPreProcessFn func(err error) error

func CreateRetriableInvokerByConf(conf *RetriableInvokerConf) RetriableInvoker {
	return &retriableInvoker{
		*conf,
	}
}

func CreateRetriableInvokerByErr(retriableErr error) RetriableInvoker {
	return CreateInvokerByErrAndFn(retriableErr, nil)
}

func CreateInvokerByErrAndFn(retriableErr error, preProccFn ErrPreProcessFn) RetriableInvoker {
	return &retriableInvoker{
		RetriableInvokerConf{
			RetriableErr:    retriableErr,
			FirstRetryDelay: time.Duration(time.Second),
			DelayIncrement:  time.Duration(2 * time.Second),
			RetryCount:      4,
			PreProccFn:      preProccFn,
		},
	}
}

type retriableInvoker struct {
	RetriableInvokerConf
}

func (r *retriableInvoker) Invoke(ctx context.Context, fn InvokableFn) error {
	var err error
	iter := 1

	logger, err := GetCtxLogger(ctx)
	if err != nil {
		fmt.Printf("can't extract logger")
		return err
	}

	for {
		logger.Infow("Invoke", "iteration", iter, "status", "start")
		err = fn(ctx)
		if err == nil {
			logger.Infow("Invoke", "iteration", iter, "status", "ok")
			return nil
		}

		if r.PreProccFn != nil {
			err = r.PreProccFn(err)
		}

		if !errors.Is(err, r.RetriableErr) || iter == r.RetryCount {
			logger.Infow("Invoke", "iteration", iter, "status", "err", "msg", err.Error())
			return err
		}
		nextInvokation := r.FirstRetryDelay + time.Duration(iter-1)*r.DelayIncrement
		select {
		case <-ctx.Done():
			logger.Infow("Invoke", "status", "err", "msg", "context cancelled")
			return ctx.Err()
		case <-time.After(nextInvokation):
			iter++
			continue
		}
	}
}
