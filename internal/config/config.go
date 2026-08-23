package config

import "time"

type Config struct {
	UnitID             string
	ZoneCount          int
	DefaultAirflowCFM     float64
	AirflowTolerancePct   float64
	DefrostCycleMinutes int
	CompressorMinRun   time.Duration
	CompressorCoast    time.Duration
	DuctPrimeSec   int
	AlarmBufferSize    int
	ProcessTickMs      int
}

func Default() Config {
	return Config{
		UnitID: "reef-unit-7", ZoneCount: 8, DefaultAirflowCFM: 12.5, AirflowTolerancePct: 5,
		DefrostCycleMinutes: 1, CompressorMinRun: time.Millisecond, CompressorCoast: time.Second,
		DuctPrimeSec: 5, AlarmBufferSize: 64, ProcessTickMs: 10,
	}
}

func (c Config) Validate() error {
	if c.ZoneCount <= 0 {
		return errConfig("zone_count must be positive")
	}
	if c.DefaultAirflowCFM < 0 {
		return errConfig("default_airflow_cfm invalid")
	}
	return nil
}

func (c Config) ProcessTick() time.Duration {
	if c.ProcessTickMs <= 0 {
		return 100 * time.Millisecond
	}
	return time.Duration(c.ProcessTickMs) * time.Millisecond
}