package domain

import (
	"errors"
	"net/http"
)

var (
	ErrServerInternal             = errors.New("InternalError")              // Ошибка на сервере
	ErrDataFormat                 = errors.New("DataFormatError")            // Неверный формат запроса
	ErrWrongOrderNumber           = errors.New("WrongOrderNumber")           // Неверный номера заказа
	ErrLoginIsBusy                = errors.New("LoginIsBusy")                // Логин занят
	ErrWrongLoginPassword         = errors.New("WrongLoginPassword")         // Не верный логин/пароль
	ErrAuthDataIncorrect          = errors.New("AuthDataIncorrect")          // Неверный JWT
	ErrNotEnoughPoints            = errors.New("NotEnoughPoints")            // Средств не достаточно
	ErrNotFound                   = errors.New("NoDataFound")                // Данные не найдены
	ErrOrderNumberAlreadyUploaded = errors.New("OrderNumberAlreadyUploaded") // Данные пользователя уже были приняты в обработку
	ErrDublicateOrderNumber       = errors.New("DublicateOrderNumber")       // Данные уже были приняты в обработку от другого пользователя
	ErrUserIsNotAuthorized        = errors.New("UserIsNotAuthorized")        // Пользователь не авторизован
	ErrBalanceChanged             = errors.New("BalanceChanged")             // Внутренняя ошибка - баланс был изменен; нужно повторить операцию
)

func MapDomainErrorToHTTPStatusErr(err error) int {
	if errors.Is(err, ErrDataFormat) {
		return http.StatusBadRequest
	}

	if errors.Is(err, ErrWrongOrderNumber) {
		return http.StatusUnprocessableEntity
	}

	if errors.Is(err, ErrLoginIsBusy) {
		return http.StatusConflict
	}

	if errors.Is(err, ErrWrongLoginPassword) {
		return http.StatusUnauthorized
	}

	if errors.Is(err, ErrAuthDataIncorrect) {
		return http.StatusUnauthorized
	}

	if errors.Is(err, ErrNotEnoughPoints) {
		return http.StatusPaymentRequired
	}

	if errors.Is(err, ErrNotFound) {
		return http.StatusNoContent
	}

	if errors.Is(err, ErrOrderNumberAlreadyUploaded) {
		return http.StatusOK
	}

	if errors.Is(err, ErrDublicateOrderNumber) {
		return http.StatusConflict
	}

	if errors.Is(err, ErrUserIsNotAuthorized) {
		return http.StatusUnauthorized
	}

	if errors.Is(err, ErrBalanceChanged) {
		return http.StatusInternalServerError
	}

	if errors.Is(err, ErrServerInternal) {
		return http.StatusInternalServerError
	}

	return http.StatusInternalServerError
}
