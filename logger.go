package decision

import "github.com/sirupsen/logrus"

var logger = logrus.New()

func SetLevel(lvl string) error {
	level, err := logrus.ParseLevel(lvl)
	if err != nil {
		return err
	}
	logger.Level = level
	return err
}
