package connector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	labelDb      = "db"
	labelSuccess = "success"

	mysqlDb = "mysql"
	redisDb = "redis"
	mongoDb = "mongo"
)

var (
	buckets = []float64{0.5, 1, 2, 3, 4, 5, 10, 30, 60, 120, 300, 600, 1200, 1800, 3600}

	dbHandledHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_call_latency_ms",
			Help:    "Latency (milliseconds) of db request that handled by the server.",
			Buckets: buckets,
		},
		[]string{labelDb, labelSuccess},
	)
)

func addDbMetrics(db string, ts int64, err error) {
	if err == nil {
		dbHandledHistogram.WithLabelValues(db, "true").Observe(float64(ts))
	} else {
		dbHandledHistogram.WithLabelValues(db, "false").Observe(float64(ts))
	}
}
