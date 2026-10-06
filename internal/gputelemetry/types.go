package gputelemetry

// NVIDIADevice is one direct local hardware sample from the NVIDIA driver.
// Index is the CUDA ordinal, not merely the NVML enumeration index.
type NVIDIADevice struct {
	Index          int
	NVMLIndex      int
	Name           string
	TemperatureC   int
	FanPercent     int
	UtilPercent    int
	MemoryUsedMiB  int
	MemoryTotalMiB int
	PowerW         float64
	PowerLimitW    float64
	CoreClockMHz   int
	MemoryClockMHz int
	PState         string
	Source         string
	SampleUnixMS   int64
}
