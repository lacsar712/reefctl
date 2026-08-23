package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lacsar712/reefctl/internal/alarms"
	"github.com/lacsar712/reefctl/internal/clock"
	"github.com/lacsar712/reefctl/internal/compressor"
	"github.com/lacsar712/reefctl/internal/config"
	"github.com/lacsar712/reefctl/internal/defrost"
	"github.com/lacsar712/reefctl/internal/fsm"
	"github.com/lacsar712/reefctl/internal/interlock"
	"github.com/lacsar712/reefctl/internal/model"
	"github.com/lacsar712/reefctl/internal/reefer"
	"github.com/lacsar712/reefctl/internal/store"
)

type App struct {
	cfg         config.Config
	clk         clock.Clock
	mem         *store.Memory
	sched       *store.ScheduleStore
	unit        *reefer.ReeferUnit
	reeferFSM   *fsm.ReeferFSM
	zones       *reefer.ZoneTable
	alarms      *alarms.Emitter
	lock        *interlock.ValveLock
	dampers     *reefer.DamperActuator
	router      *reefer.Router
	heater      *defrost.HeaterController
	phaseCtrl   *defrost.PhaseController
	partitions  *reefer.PartitionTable
	balancer    *reefer.DuctBalancer
	staging     *compressor.StagingInterlock
	cargo          *reefer.CargoMonitor
	unitFSM        *fsm.UnitFSM
	scheduler      *clock.UnitScheduler
	zoneWindow     *clock.DefrostWindow
	unitLeases     *interlock.UnitLeaseRegistry
	unitMu         sync.Mutex
	unitCancels    map[model.UnitID]context.CancelFunc
	compressorStart *CompressorStart
}

func New(cfg config.Config) (*App, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	clk := clock.NewProcessClock(start, cfg.ProcessTick())
	mem := store.NewMemory()
	unitID, err := model.ParseUnitID(cfg.UnitID)
	if err != nil {
		return nil, err
	}
	ductID := model.DuctID("duct-main")
	zones, err := reefer.NewZoneTable(unitID, cfg.ZoneCount, ductID)
	if err != nil {
		return nil, err
	}
	plant := reefer.NewReeferUnit(cfg, clk, mem)
	plant.Ducts().Add(&reefer.Duct{ID: ductID, Capacity: 100})
	plant.BindFlow(ductID, model.AirflowSetpoint{CubicFeetPerMinute: cfg.DefaultAirflowCFM, TolerancePct: cfg.AirflowTolerancePct})
	plant.Coordinator().Register(model.CompressorID("comp-1"))
	heater := defrost.NewHeaterController(clk, time.Duration(cfg.DefrostCycleMinutes)*time.Minute)
	a := &App{
		cfg:        cfg,
		clk:        clk,
		mem:        mem,
		sched:      store.NewScheduleStore(mem),
		unit:       plant,
		zones:      zones,
		lock:       interlock.NewValveLock(clk.Now),
		router:     reefer.NewRouter([]model.DuctRoute{{From: ductID, To: "duct-evap", Valve: "damper-main", Priority: 10}}),
		heater:     heater,
		phaseCtrl:  defrost.NewPhaseController(clk, heater, ductID),
		partitions: reefer.NewPartitionTable(unitID),
		balancer:   reefer.NewDuctBalancer(plant.Ducts()),
		cargo:       reefer.NewCargoMonitor(),
		unitCancels: make(map[model.UnitID]context.CancelFunc),
	}
	a.dampers = reefer.NewDamperActuator(a.lock)
	a.alarms = alarms.NewEmitter(alarms.NewRegistry(), clk, cfg.AlarmBufferSize)
	a.reeferFSM = fsm.NewReeferFSM(unitID, a.onUnitTransition)
	a.staging = compressor.NewStagingInterlock(
		plant.Coordinator(),
		plant.DefrostActive,
		cfg.CompressorCoast,
		[]model.CompressorID{"comp-1"},
		clk.Now,
	)
	if err := a.syncPartitions(); err != nil {
		return nil, err
	}
	a.unitFSM = fsm.NewUnitFSM(unitID, a.onUnitFSMTransition)
	fsm.RegisterUnitCompressorHook(a.unitFSM.Hooks())
	if pc, ok := a.clk.(*clock.ProcessClock); ok {
		a.scheduler = clock.NewUnitScheduler(*pc)
		a.zoneWindow = clock.NewDefrostWindow(a.clk, time.Duration(cfg.ZoneWindowMinutes)*time.Minute)
	}
	a.unitLeases = interlock.NewUnitLeaseRegistry(a.clk.Now)
	a.compressorStart = NewCompressorStart(a.clk, cfg.ProcessTick(), cfg.CompressorStartSteps)
	a.persistSnapshot(unitID)
	return a, nil
}

func (a *App) syncPartitions() error {
	for i := 0; i < a.cfg.ZoneCount && i < 4; i++ {
		_ = a.zones.Enable(i, true)
	}
	return a.partitions.BuildFromZones(a.zones, a.cargo)
}

func (a *App) applyAirflowBalance() error {
	plan, err := reefer.BuildBalancePlan(a.unit.Ducts(), a.partitions, a.cfg.DefaultAirflowCFM, a.cfg.AirflowTolerancePct)
	if err != nil {
		return err
	}
	a.balancer.ApplyPlan(plan)
	for duct, sp := range plan.Setpoints() {
		a.unit.BindFlow(duct, sp)
	}
	return nil
}

func (a *App) onUnitTransition(ctx context.Context, unit model.UnitID, from, to model.UnitState) error {
	if to == model.UnitFault {
		a.staging.TripAll(ctx)
		return a.alarms.Raise(ctx, "COMPRESSOR_TRIP", unit, 3)
	}
	return nil
}

func (a *App) onUnitFSMTransition(ctx context.Context, unit model.UnitID, from, to fsm.UnitPhase) error {
	_ = ctx
	_ = unit
	_ = from
	_ = to
	return nil
}

func (a *App) persistSnapshot(id model.UnitID) {
	b := store.NewSnapshotBuilder(id).State(a.reeferFSM.State())
	for _, s := range a.zones.Slots() {
		b.Slot(model.ZoneAssignment{
			Slot:     s.ID,
			Duct:     s.Duct,
			Enabled:  s.Enabled,
			Setpoint: model.AirflowSetpoint{CubicFeetPerMinute: a.cfg.DefaultAirflowCFM, TolerancePct: a.cfg.AirflowTolerancePct},
		})
	}
	a.mem.PutUnit(b.Build(a.clk.Now()))
}

func (a *App) ApplyScheduleSnapshot(ctx context.Context, id model.ScheduleID) error {
	snap, err := a.sched.SnapshotClone(id)
	if err != nil {
		return err
	}
	now := a.clk.Now()
	entry, ok := a.sched.ActiveEntry(snap, now)
	if !ok {
		return model.Wrap("app", "schedule", model.ErrScheduleEmpty)
	}
	a.unit.BindFlow(entry.Duct, entry.Setpoint)
	win := defrost.NewWindow(now, time.Duration(a.cfg.DefrostCycleMinutes)*time.Minute, entry.TargetCelsius)
	a.unit.ArmDefrostCycle(win)
	a.phaseCtrl.Begin(win)
	return nil
}

func (a *App) RunOnce(ctx context.Context) error {
	if err := a.reeferFSM.Apply(ctx, "prime"); err != nil {
		return err
	}
	duct := model.DuctID("duct-main")
	if err := a.unit.PrimeDuct(ctx, duct); err != nil {
		return err
	}
	if err := a.applyAirflowBalance(); err != nil {
		return err
	}
	if err := a.reeferFSM.Apply(ctx, "flow_ok"); err != nil {
		return err
	}
	if err := a.staging.RequestStart(ctx, model.CompressorID("comp-1")); err != nil {
		return err
	}
	a.unit.ObserveFlow(duct, a.cfg.DefaultAirflowCFM)
	_ = a.balancer.Observe(duct, a.cfg.DefaultAirflowCFM)
	if err := a.balancer.ValidateAll(ctx); err != nil {
		return err
	}
	win := defrost.NewWindow(a.clk.Now(), time.Duration(a.cfg.DefrostCycleMinutes)*time.Minute, 6.5)
	a.unit.ArmDefrostCycle(win)
	a.phaseCtrl.Begin(win)
	if err := a.reeferFSM.Apply(ctx, "defrost_cycle"); err != nil {
		return err
	}
	reading, _ := a.unit.Sensors().Reading(model.SensorID("evap-temp"))
	_ = a.phaseCtrl.Advance(ctx, 1.5, reading)
	if pc, ok := a.clk.(*clock.ProcessClock); ok {
		pc.Advance(time.Duration(a.cfg.DefrostCycleMinutes)*time.Minute + time.Second)
	}
	reading, _ = a.unit.Sensors().Reading(model.SensorID("evap-temp"))
	_ = a.phaseCtrl.Advance(ctx, 0, reading)
	a.phaseCtrl.ForceRelease()
	if err := a.reeferFSM.Apply(ctx, "release_defrost"); err != nil {
		return err
	}
	a.persistSnapshot(model.UnitID(a.cfg.UnitID))
	return nil
}

func (a *App) StatusLine() string {
	return fmt.Sprintf("unit=%s state=%s defrost=%v phase=%s zones=%d %s",
		a.cfg.UnitID, a.reeferFSM.State(), a.unit.DefrostActive(), a.phaseCtrl.Phase(),
		a.zones.EnabledCount(), a.staging.Status())
}
