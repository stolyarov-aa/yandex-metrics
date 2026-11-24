package service

import (
	"flag"
	"fmt"
)

func ParseFlags(host *string, pollInterval *int, reportInterval *int, args []string) error {
	fs := flag.NewFlagSet("flagSet", flag.ContinueOnError)
	fs.StringVar(host, "a", "localhost:8080", "a - адрес эндпоинта HTTP-сервера (по умолчанию localhost:8080)")

	if pollInterval != nil {
		fs.IntVar(pollInterval, "p", 2, "p - pollInterval — частота опроса метрик из пакета runtime (по умолчанию 2 секунды)")
	}
	if reportInterval != nil {
		fs.IntVar(reportInterval, "r", 10, "r - reportInterval — частота отправки метрик на сервер (по умолчанию 10 секунд)")
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() > 0 {
		return fmt.Errorf("указаны неизвестные аргументы: %v", fs.Args())
	}

	if host != nil {
		if err := ValidateHost(*host); err != nil {
			return fmt.Errorf("неверный хост: %w", err)
		}
	}
	if pollInterval != nil && *pollInterval <= 0 {
		return fmt.Errorf("pollInterval должен быть > 0")
	}
	if reportInterval != nil && *reportInterval <= 0 {
		return fmt.Errorf("reportInterval должен быть > 0")
	}

	return nil
}
