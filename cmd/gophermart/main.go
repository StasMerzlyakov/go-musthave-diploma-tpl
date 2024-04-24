package main

import (
	"context"

	"github.com/StasMerzlyakov/gophermart/internal/config"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/storage/pgx"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
	"go.uber.org/zap"
)

func main() {

	// -------- Контекст сервера ---------
	srvCtx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	// -------- Конфигурация ----------
	srvConf, err := config.LoadGophermartConfig()
	if err != nil {
		panic(err)
	}

	// -------- Логгер -------------------
	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic("cannot initialize zap")
	}
	defer logger.Sync()

	sugarLog := logger.Sugar()
	domain.SetMainLogger(sugarLog)

	// -------- Запуск хранилища --------
	storage := pgx.NewStorage(srvCtx, srvConf)

	if err := storage.Ping(srvCtx); err != nil {
		panic(err)
	}

	defer storage.Close()

}
