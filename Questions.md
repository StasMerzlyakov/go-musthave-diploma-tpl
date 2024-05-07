# Обсудить ADR

# Миграции базы данных
Какие типовые решения/библиотеки применяются? (ответ получен - TODO - https://github.com/pressly/goose)

# используется для получения имени вызывющей функции
# по мотивам 
- https://stackoverflow.com/questions/25927660/how-to-get-the-current-function-name

#	pc, _, _, _ := runtime.Caller(1)
#	action := runtime.FuncForPC(pc).Name() TODO - переделать в логгере


# goroutine - есть ли автоматические средства для анализа?
- https://pkg.go.dev/cmd/go (go run -race)  TODO - проверить приложение

# swagger + chi
- https://github.com/ganpatagarwal/chi-swagger


TODO - обработка 409 ошибки, zap.With








