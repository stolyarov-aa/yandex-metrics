package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRuntimeMetrics(t *testing.T) {
	want := []string{
		"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys",
		"HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased",
		"HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys",
		"MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC",
		"NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys",
		"Sys", "TotalAlloc",
	}

	got := make(map[string]float64)
	for _, m := range GetRuntimeMetrics() {
		_, dup := got[m.Name]
		require.False(t, dup, "метрика %q встречается дважды", m.Name)
		got[m.Name] = m.Value
	}

	assert.Len(t, got, len(want))
	for _, name := range want {
		assert.Contains(t, got, name)
	}
	assert.Positive(t, got["Sys"])
	assert.Positive(t, got["HeapAlloc"])
}
