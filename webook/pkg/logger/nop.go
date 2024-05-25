package logger

type NopLogger struct{}

func NewNopLogger() LoggerV1 {
	return &NopLogger{}
}

func (z *NopLogger) Debug(msg string, args ...Field) {
}

func (z *NopLogger) Info(msg string, args ...Field) {
}

func (z *NopLogger) Warn(msg string, args ...Field) {
}

func (z *NopLogger) Error(msg string, args ...Field) {
}
