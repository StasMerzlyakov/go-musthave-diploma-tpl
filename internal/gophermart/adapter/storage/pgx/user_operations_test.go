package pgx_test

import (
	"context"
	"testing"

	"github.com/StasMerzlyakov/gophermart/internal/config"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/storage/pgx"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthFunctions(t *testing.T) {
	ctx, cancelFN := context.WithCancel(context.Background())

	defer cancelFN()

	connString, err := postgresContainer.ConnectionString(ctx)

	require.NoError(t, err)

	logger := createLogger()
	domain.SetMainLogger(logger)
	storage := pgx.NewStorage(ctx, &config.GophermartConfig{
		MaxConns:    5,
		DatabaseUri: connString,
	})

	defer func() {
		err = clear(ctx)
		require.NoError(t, err)
	}()

	err = storage.Ping(ctx)
	require.NoError(t, err)

	err = clear(ctx)
	require.NoError(t, err)

	// Все бизнес-операции выполняются с ctx, содержащим logger
	requestUUID := uuid.New()
	loggedCtx := domain.EnrichWithRequestIDLogger(ctx, requestUUID, logger)

	login := "login"
	passHash := "hash"
	salt := "salt"
	ldata, err := storage.GetUserData(loggedCtx, login)
	require.NoError(t, err)
	require.Nil(t, ldata)

	ldata = &domain.LoginData{
		Login: login,
		Hash:  passHash,
		Salt:  salt,
	}

	_, err = storage.RegisterUser(loggedCtx, ldata)
	require.NoError(t, err)

	_, err = storage.RegisterUser(loggedCtx, ldata)
	require.ErrorIs(t, err, domain.ErrLoginIsBusy)

	ldata, err = storage.GetUserData(loggedCtx, login)
	require.NoError(t, err)
	require.NotNil(t, ldata)

	assert.Equal(t, login, ldata.Login)
	assert.Equal(t, passHash, ldata.Hash)
	assert.Equal(t, salt, ldata.Salt)
	assert.True(t, ldata.UserID > 0)
}
