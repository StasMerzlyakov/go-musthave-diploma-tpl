package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/StasMerzlyakov/gophermart/internal/config"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/in/http/middleware"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/in/http/middleware/logging"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/in/http/middleware/retry"

	hauth "github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/in/http/handler/auth"
	hbalance "github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/in/http/handler/balance"
	horder "github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/in/http/handler/order"

	mauth "github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/in/http/middleware/auth"

	"github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/out/service/accrual"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/out/storage/pgx"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/app"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func main() {

	// -------- Контекст сервера ---------
	srvCtx, cancelFn := context.WithCancel(context.Background())

	defer cancelFn()

	// -------- Конфигурация ----------
	srvConf, err := config.LoadGophermartConfig()

	//srvConf.AccrualSystemAddress = "http://localhost:8080"
	//srvConf.DatabaseUri = "postgres://postgres:postgres@localhost:5432/gophermarket"
	//srvConf.RunAddress = ":8081"

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

	// ---------- Система расчета баллов лояльности -------
	accrualService := accrual.New(srvConf)

	// ---------- Запуск приложения -------------------
	auth := app.NewAuth(srvConf, storage)
	balance := app.NewBalance(srvConf, storage)
	order := app.NewOrder(srvConf, storage, accrualService)

	order.PoolAcrualSystem(srvCtx)
	balance.PoolOrders(srvCtx)

	// ----------- Настройка http.Handler ------------------
	httpHandler := chi.NewMux()

	// мидлы
	mwList := []func(http.Handler) http.Handler{
		logging.EncrichWithRequestIDMW(),
		logging.NewLoggingRequestMW(),
		logging.NewLoggingResponseMW(),
		retry.NewRetriableRequestMW(),
	}
	httpHandler.Use(mwList...)

	authMW := mauth.NewJwtRequestMW(auth)

	httpHandler.Route("/api/user", func(r chi.Router) {
		r.Post("/register", hauth.RegisterHandler(auth))
		r.Post("/login", hauth.LoginHandler(auth))

		r.Get("/orders", middleware.ConveyorFunc(horder.GetHandler(order), authMW))
		r.Post("/orders", middleware.ConveyorFunc(horder.CreateHandler(order), authMW))

		r.Get("/balance", middleware.ConveyorFunc(hbalance.GetHandler(balance), authMW))
		r.Post("/withdraw", middleware.ConveyorFunc(hbalance.WithdrawHandler(balance), authMW))

		r.Get("/withdrawals", middleware.ConveyorFunc(hbalance.GetWithdrawals(balance), authMW))
	})

	// ------- Запуск сервера -----
	srv := &http.Server{
		Addr:        srvConf.RunAddress,
		Handler:     httpHandler,
		ReadTimeout: 0,
		IdleTimeout: 0,
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			sugarLog.Fatalw("ListenAndServe", "msg", err.Error())
			panic(err)
		}
	}()

	// --------------- Обрабатываем остановку сервера --------------
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	defer func() {
		cancelFn()
		srv.Shutdown(srvCtx)
	}()
	<-exit
}
