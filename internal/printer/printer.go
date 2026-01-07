package printer

type Printer interface {
	Info(msg string)
	Success(msg string)
	Warn(msg string)
	Error(msg string)

	// Structured outputs
	Table(headers []string, rows [][]string)
}
