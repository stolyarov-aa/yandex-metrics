package errors

import "fmt"

type (
	InvalidMetricTypeError struct {
		MetricType string
	}
	InvalidMetricNameError struct {
		MetricName string
	}
	InvalidMetricDataTypeError struct {
		MetricType string
		MetricName string
		DataType   string
	}
)

func (e *InvalidMetricTypeError) Error() string {
	return fmt.Sprintf("недопустимый тип метрики %q: разрешены только 'gauge' и 'counter'", e.MetricType)
}

func (e *InvalidMetricTypeError) As(target any) bool {
	if target == nil {
		return false
	}
	if t, ok := target.(**InvalidMetricTypeError); ok {
		*t = e
		return true
	}
	return false
}

func (e *InvalidMetricNameError) Error() string {
	return fmt.Sprintf("недопустимое имя метрики %q", e.MetricName)
}

func (e *InvalidMetricNameError) As(target any) bool {
	if t, ok := target.(**InvalidMetricNameError); ok {
		*t = e
		return true
	}
	return false
}

func (e *InvalidMetricDataTypeError) Error() string {
	return fmt.Sprintf("неверный тип данных %s-метрики %q: %s", e.MetricType, e.MetricName, e.DataType)
}

func (e *InvalidMetricDataTypeError) As(target any) bool {
	if t, ok := target.(**InvalidMetricDataTypeError); ok {
		*t = e
		return true
	}
	return false
}
