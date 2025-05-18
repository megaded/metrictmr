package collector

import (
	"math/rand/v2"
	"runtime"
	"sync"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type MetricName string

type GaugeMetric struct {
	Name  MetricName
	Value float64
}

type Counter struct {
	Name  MetricName
	Value int64
}

type Metric struct {
	GaugeMetrics   []GaugeMetric
	CounterMetrics []Counter
}

type MetricCollector struct {
	PollCount int
}

func (c *MetricCollector) IncreasePollCount() {
	var mutex sync.Mutex
	mutex.Lock()
	defer mutex.Unlock()
	c.PollCount++
}

const (
	Alloc           = MetricName("Alloc")
	BuckHashSys     = MetricName("BuckHashSys")
	Frees           = MetricName("Frees")
	GCCPUFraction   = MetricName("GCCPUFraction")
	GCSys           = MetricName("GCSys")
	HeapAlloc       = MetricName("HeapAlloc")
	HeapIdle        = MetricName("HeapIdle")
	HeapInuse       = MetricName("HeapInuse")
	HeapObjects     = MetricName("HeapObjects")
	HeapReleased    = MetricName("HeapReleased")
	HeapSys         = MetricName("HeapSys")
	LastGC          = MetricName("LastGC")
	Lookups         = MetricName("Lookups")
	MCacheInuse     = MetricName("MCacheInuse")
	MCacheSys       = MetricName("MCacheSys")
	MSpanInuse      = MetricName("MSpanInuse")
	MSpanSys        = MetricName("MSpanSys")
	Mallocs         = MetricName("Mallocs")
	NextGC          = MetricName("NextGC")
	NumForcedGC     = MetricName("NumForcedGC")
	NumGC           = MetricName("NumGC")
	OtherSys        = MetricName("OtherSys")
	PauseTotalNs    = MetricName("PauseTotalNs")
	StackInuse      = MetricName("StackInuse")
	StackSys        = MetricName("StackSys")
	Sys             = MetricName("Sys")
	TotalAlloc      = MetricName("TotalAlloc")
	PollCount       = MetricName("PollCount")
	RandomValue     = MetricName("RandomValue")
	FreeMemory      = MetricName("FreeMemory")
	TotalMemory     = MetricName("TotalMemory")
	CPUutilization1 = MetricName("CPUutilization1")
)

func (c *MetricCollector) GetRunTimeMetrics() Metric {
	defer func() {
		c.IncreasePollCount()
	}()
	return Metric{GaugeMetrics: GetGaugeMetrics(), CounterMetrics: []Counter{
		{Name: PollCount, Value: int64(c.PollCount)},
	}}
}

func GetGaugeMetrics() []GaugeMetric {
	runTimeMetrics := &runtime.MemStats{}
	runtime.ReadMemStats(runTimeMetrics)
	v, _ := mem.VirtualMemory()
	c, _ := cpu.Counts(false)
	m := []GaugeMetric{
		{Name: Alloc, Value: float64(runTimeMetrics.Alloc)},
		{Name: BuckHashSys, Value: float64(runTimeMetrics.BuckHashSys)},
		{Name: Frees, Value: float64(runTimeMetrics.Frees)},
		{Name: GCCPUFraction, Value: float64(runTimeMetrics.GCCPUFraction)},
		{Name: GCSys, Value: float64(runTimeMetrics.GCSys)},
		{Name: HeapAlloc, Value: float64(runTimeMetrics.HeapAlloc)},
		{Name: HeapIdle, Value: float64(runTimeMetrics.HeapIdle)},
		{Name: HeapInuse, Value: float64(runTimeMetrics.HeapInuse)},
		{Name: HeapObjects, Value: float64(runTimeMetrics.HeapObjects)},
		{Name: HeapReleased, Value: float64(runTimeMetrics.HeapReleased)},
		{Name: HeapSys, Value: float64(runTimeMetrics.HeapSys)},
		{Name: LastGC, Value: float64(runTimeMetrics.LastGC)},
		{Name: Lookups, Value: float64(runTimeMetrics.Lookups)},
		{Name: MCacheInuse, Value: float64(runTimeMetrics.MCacheInuse)},
		{Name: MCacheSys, Value: float64(runTimeMetrics.MCacheSys)},
		{Name: MSpanInuse, Value: float64(runTimeMetrics.MSpanInuse)},
		{Name: MSpanSys, Value: float64(runTimeMetrics.MSpanSys)},
		{Name: Mallocs, Value: float64(runTimeMetrics.Mallocs)},
		{Name: NextGC, Value: float64(runTimeMetrics.NextGC)},
		{Name: NumForcedGC, Value: float64(runTimeMetrics.NumForcedGC)},
		{Name: NumGC, Value: float64(runTimeMetrics.NumGC)},
		{Name: OtherSys, Value: float64(runTimeMetrics.OtherSys)},
		{Name: PauseTotalNs, Value: float64(runTimeMetrics.PauseTotalNs)},
		{Name: StackInuse, Value: float64(runTimeMetrics.StackInuse)},
		{Name: StackSys, Value: float64(runTimeMetrics.StackSys)},
		{Name: Sys, Value: float64(runTimeMetrics.Sys)},
		{Name: TotalAlloc, Value: float64(runTimeMetrics.TotalAlloc)},
		{Name: RandomValue, Value: rand.Float64() * 1000},
		{Name: FreeMemory, Value: float64(v.Free)},
		{Name: TotalMemory, Value: float64(v.Total)},
		{Name: CPUutilization1, Value: float64(c)},
	}
	return m
}
