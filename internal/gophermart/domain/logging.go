package domain

var mainLogger Logger

func GetMainLogger() Logger {
	return mainLogger
}

func SetMainLogger(logger Logger) {
	mainLogger = logger
}
