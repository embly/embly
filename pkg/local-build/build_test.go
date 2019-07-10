package localbuild

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestCreate(t *testing.T) {
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: false,
	})
	logrus.SetLevel(logrus.DebugLevel)
	// err := Create()
	t.Error(nil)
}
