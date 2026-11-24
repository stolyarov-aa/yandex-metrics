package service

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
)

func ValidateHost(host string) error {
	if host == "" {
		return fmt.Errorf("хост не может быть пустым")
	}
	
	h, p, err := net.SplitHostPort(host)
	if err != nil {
		return fmt.Errorf("некорректный формат хоста, ожидается 'host:port', получено: %q. Ошибка: %v", host, err)
	}

	if h == "" {
		return fmt.Errorf("хост не указан")
	}

	if p == "" {
		return fmt.Errorf("порт не указан")
	}

	if err := ValidatePort(p); err != nil {
		return err
	}

	return nil
}

func ValidatePort(port string) error {
	if port == "" {
		return fmt.Errorf("порт не указан")
	}
	num, err := strconv.Atoi(port)
	if err != nil || !regexp.MustCompile(`^\d+$`).MatchString(port) {
		return fmt.Errorf("некорректный номер порта %q: не является числом", port)
	}
	if num < 1 || num > 65535 {
		return fmt.Errorf("некорректный номер порта %d: должен быть в диапазоне от 1 до 65535", num)
	}
	return nil
}
