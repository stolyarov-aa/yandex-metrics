package service

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

func ParseFlags(host *string, pollInterval *int, reportInterval *int, args []string) error {
	fs := flag.NewFlagSet("flagSet", flag.ContinueOnError)

	if host != nil {
		fs.StringVar(host, "a", "localhost:8080", "a - адрес эндпоинта HTTP-сервера (по умолчанию localhost:8080)")
	}

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

	if err := getEnvVars(host, pollInterval, reportInterval); err != nil {
		return fmt.Errorf("ошибка при обработке переменных среды [%v]", err)
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

func getEnvVars(host *string, pollInterval *int, reportInterval *int) error {
	if host != nil {
		if val := os.Getenv("ADDRESS"); val != "" {
			*host = val
		}
	}

	if pollInterval != nil {
		if b, val, err := parseEnvDuration("POLL_INTERVAL"); err != nil {
			return err
		} else {
			if b {
				*pollInterval = val
			}
		}
	}

	if reportInterval != nil {
		if b, val, err := parseEnvDuration("REPORT_INTERVAL"); err != nil {
			return err
		} else {
			if b {
				*reportInterval = val
			}
		}
	}

	return nil
}

func parseEnvDuration(name string) (bool, int, error) {
	val := os.Getenv(name)
	if val == "" {
		return false, 0, nil
	}
	sec, err := strconv.Atoi(val)
	if err != nil || sec <= 0 {
		return true, 0, fmt.Errorf("%s=%q: %v", name, val, err)
	} else {
		return true, sec, nil
	}
}
