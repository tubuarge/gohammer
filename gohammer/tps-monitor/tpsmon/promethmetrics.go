package tpsmon

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

type PrometheusMetricsService struct {
	tpsGauge    prometheus.Gauge
	txnsGauge   prometheus.Gauge
	blocksGauge prometheus.Gauge
	port        int
}

func NewPrometheusMetricsService(port int) *PrometheusMetricsService {
	ps := &PrometheusMetricsService{
		tpsGauge: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "Quorum",
			Subsystem: "TransactionProcessing",
			Name:      "TPS",
			Help:      "Transactions processed per second",
		}),
		blocksGauge: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "Quorum",
			Subsystem: "TransactionProcessing",
			Name:      "total_blocks",
			Help:      "total blocks processed",
		}),
		txnsGauge: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "Quorum",
			Subsystem: "TransactionProcessing",
			Name:      "total_transactions",
			Help:      "total transactions processed",
		}),
		port: port,
	}
	return ps
}

func (ps *PrometheusMetricsService) Start() {
	prometheus.MustRegister(ps.tpsGauge)
	prometheus.MustRegister(ps.txnsGauge)
	prometheus.MustRegister(ps.blocksGauge)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", ps.port), nil))
}

func (ps *PrometheusMetricsService) publishMetrics(ref time.Time, tps uint64, txnCnt uint64, blkCnt uint64) {
	ps.tpsGauge.SetToCurrentTime()
	ps.tpsGauge.Set(float64(tps))
	ps.txnsGauge.SetToCurrentTime()
	ps.txnsGauge.Set(float64(txnCnt))
	ps.blocksGauge.SetToCurrentTime()
	ps.blocksGauge.Set(float64(blkCnt))
	log.Debug("published metrics to prometheus")
}
