package model

import "time"

type UnitState string

const (
	UnitIdle      UnitState = "idle"
	UnitPriming   UnitState = "priming"
	UnitCirculate UnitState = "circulate"
	UnitDefrost      UnitState = "hold"
	UnitFault     UnitState = "fault"
	UnitShutdown  UnitState = "shutdown"
)

type CompressorState string

const (
	CompressorOff     CompressorState = "off"
	CompressorStaging CompressorState = "staging"
	CompressorRun     CompressorState = "run"
	CompressorCoast   CompressorState = "coast"
	CompressorTrip    CompressorState = "trip"
)

type ValvePosition string

const (
	ValveClosed    ValvePosition = "closed"
	ValveOpen      ValvePosition = "open"
	ValveThrottled ValvePosition = "throttled"
)

type AirflowSetpoint struct {
	CubicFeetPerMinute float64
	TolerancePct    float64
}

func (f AirflowSetpoint) Within(actual float64) bool {
	if f.CubicFeetPerMinute <= 0 {
		return actual <= 0
	}
	lo := f.CubicFeetPerMinute * (1 - f.TolerancePct/100)
	hi := f.CubicFeetPerMinute * (1 + f.TolerancePct/100)
	return actual >= lo && actual <= hi
}

type TempReading struct {
	Sensor  SensorID
	Celsius float64
	At      time.Time
}

type ZoneAssignment struct {
	Slot     ZoneID
	Duct DuctID
	Setpoint AirflowSetpoint
	Enabled  bool
	LastFlow float64
}

type UnitSnapshot struct {
	ID          UnitID
	State       UnitState
	Slots       []ZoneAssignment
	Compressors []CompressorID
	UpdatedAt   time.Time
}

type DefrostScheduleEntry struct {
	ID          ScheduleID
	Duct    DuctID
	Start       time.Time
	End         time.Time
	Setpoint    AirflowSetpoint
	TargetCelsius float64
}

type DefrostSchedule struct {
	ID      ScheduleID
	Entries []DefrostScheduleEntry
	Version int64
}

func (s DefrostSchedule) Clone() DefrostSchedule {
	out := DefrostSchedule{ID: s.ID, Version: s.Version}
	if len(s.Entries) == 0 {
		return out
	}
	out.Entries = make([]DefrostScheduleEntry, len(s.Entries))
	copy(out.Entries, s.Entries)
	return out
}

type AlarmEvent struct {
	Code     AlarmCode
	Message  string
	Unit     UnitID
	RaisedAt time.Time
	Severity int
}

type DuctRoute struct {
	From     DuctID
	To       DuctID
	Valve    ValveID
	Priority int
}