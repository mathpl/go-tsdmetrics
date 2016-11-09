package prometheusreporter

import "github.com/Sirupsen/logrus"

var log = logrus.FieldLogger(logrus.New())

func SetLogger(logger logrus.FieldLogger) {
	log = logger
}
