package repository

type MemStorageIntf interface {
	SetGaugeMetric(name string, value float64) error
	GetGaugeMetric(name string) (float64, bool)
	AddCounterMetric(name string, value int64) error
	GetCounterMetric(name string) (int64, bool)
	GetAllGaugeMetrics() map[string]float64
	GetAllCounterMetrics() map[string]int64
}
