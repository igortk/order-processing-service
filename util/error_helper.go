package util

import log "github.com/sirupsen/logrus"

func IsError(err error, msg string) {
	if err != nil {
		log.Errorf("%s: %s", msg, err)
		return
	}
}
