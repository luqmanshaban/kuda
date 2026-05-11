package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	JobsEnqueued = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "kuda_jobs_enqueued_total",
		Help: "Total number of jobs submitted",
	})
	JobsCompleted = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "kuda_jobs_completed_total",
		Help: "Total number of jobs completed",
	})
	JobsFailed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "kuda_jobs_failed_total",
		Help: "Total number of jobs failed",
	})
	JobDeliveryDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "kuda_job_delivery_duration_seconds",
		Help: "webhook deliver in seconds",
		Buckets: prometheus.DefBuckets,
	})
)

func Init() {
	prometheus.MustRegister(JobsEnqueued)
	prometheus.MustRegister(JobsCompleted)
	prometheus.MustRegister(JobsFailed)
	prometheus.MustRegister(JobDeliveryDuration)
}