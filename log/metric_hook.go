package log

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
)

var errLogCounter = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "go_service_log_error_count",
	}, []string{"level"},
)

type metricHook struct{}

func (m metricHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
	}
}

func (m metricHook) Fire(entry *logrus.Entry) error {
	errLogCounter.WithLabelValues(entry.Level.String()).Add(1)
	return nil
}
