package kc

import (
	"os/exec"
)

const serviceName = "pot"

func AddPassword(host string, user string, password string) error {
	secCmd := exec.Command(
		"security",
		"add-generic-password",
		"-s", serviceName,
		"-a", host,
		"-l", user,
		"-w", password,
		"-U",
	)

	return secCmd.Run()
}

func GetPassword(host string, user string) (string, error) {
	cmd := exec.Command(
		"security",
		"find-generic-password",
		"-s", serviceName,
		"-a", host,
		"-l", user,
		"-w",
	)

	passwordBytes, err := cmd.Output()

	if err != nil {
		return "", err
	}

	passwordBytes = passwordBytes[:len(passwordBytes)-1]
	return string(passwordBytes), nil
}

func DeletePassword(host string) error {
	cmd := exec.Command(
		"security",
		"delete-generic-password",
		"-s", serviceName,
		"-a", host,
	)

	return cmd.Run()
}
