package errors

import (
	"errors"
	"testing"
)

func TestInvalidMetricTypeError(t *testing.T) {
	t.Parallel()

	err := &InvalidMetricTypeError{MetricType: "invalid"}
	if err.Error() != "недопустимый тип метрики \"invalid\": разрешены только 'gauge' и 'counter'" {
		t.Errorf("некорректное сообщение об ошибке: %s", err.Error())
	}

	var target *InvalidMetricTypeError
	if !errors.As(err, &target) {
		t.Errorf("ошибка не реализует As()")
	}
}

func TestInvalidMetricNameError(t *testing.T) {
	t.Parallel()

	err := &InvalidMetricNameError{MetricName: "<<Имя не указано>>"}
	if err.Error() != "недопустимое имя метрики \"<<Имя не указано>>\"" {
		t.Errorf("некорректное сообщение об ошибке: %s", err.Error())
	}

	var target *InvalidMetricNameError
	if !errors.As(err, &target) {
		t.Errorf("ошибка не реализует As()")
	}
}

func TestInvalidMetricDataTypeError(t *testing.T) {
	t.Parallel()

	err := &InvalidMetricDataTypeError{
		MetricType: "gauge",
		MetricName: "test",
		DataType:   "string",
	}
	if err.Error() != "неверный тип данных gauge-метрики \"test\": string" {
		t.Errorf("некорректное сообщение об ошибке: %s", err.Error())
	}

	var target *InvalidMetricDataTypeError
	if !errors.As(err, &target) {
		t.Errorf("ошибка не реализует As()")
	}
}
