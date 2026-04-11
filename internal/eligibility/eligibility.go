// Package eligibility computes malus scores, maps them to tiers, and determines
// spawn eligibility for agents. It ports eligibility.py from the Python armies CLI.
package eligibility

import (
	"log/slog"
	"math"
	"time"
)

// nowFunc is the clock source used by EffectiveMalus. Override via SetNowFunc in tests.
var nowFunc = func() time.Time { return time.Now() }

// SetNowFunc replaces the clock function and returns the previous one.
// Use in tests to pin time; always defer restoring the original.
func SetNowFunc(f func() time.Time) func() time.Time {
	old := nowFunc
	nowFunc = f
	return old
}

// KnownRoles returns every role the eligibility system tracks.
// A function is used rather than a variable to prevent callers from mutating the slice.
func KnownRoles() []string {
	return []string{"coordinator", "emergency_reserve", "specialist", "validator"}
}

// Tier describes a malus band and the spawn gate for each role within that band.
type Tier struct {
	Name             string
	Min              float64
	Coordinator      string
	EmergencyReserve string
	Specialist       string
	Validator        string
}

// MalusEntry is one line from an agent's malus ledger.
type MalusEntry struct {
	RawMalus float64
	Share    float64
	Date     string
	Decays   bool
}

// Status is the computed eligibility result for one agent.
type Status struct {
	EffectiveMalus float64
	Tier           string
	Roles          map[string]string
	Overall        string // "eligible" | "restricted" | "blocked"
}

// tiers is the ordered tier table. Entries are half-open intervals [Min, next.Min).
// The last entry covers [400, ∞). Keep in ascending Min order.
var tiers = []Tier{
	{Name: "Clean", Min: 0, Coordinator: "CLEAR", EmergencyReserve: "CLEAR", Specialist: "CLEAR", Validator: "CLEAR"},
	{Name: "Warning", Min: 100, Coordinator: "BLOCKED", EmergencyReserve: "FOUNDER", Specialist: "CLEAR", Validator: "CLEAR"},
	{Name: "Probation", Min: 200, Coordinator: "BLOCKED", EmergencyReserve: "BLOCKED", Specialist: "REVIEW", Validator: "CLEAR"},
	{Name: "Demotion risk", Min: 300, Coordinator: "BLOCKED", EmergencyReserve: "BLOCKED", Specialist: "ESCALATE", Validator: "CLEAR"},
	{Name: "Suspension", Min: 400, Coordinator: "BLOCKED", EmergencyReserve: "BLOCKED", Specialist: "BLOCKED", Validator: "BLOCKED"},
}

// TierFor returns the Tier that contains effectiveMalus.
// The input is rounded to 10 decimal places to eliminate floating-point noise
// before the interval comparison.
func TierFor(effectiveMalus float64) Tier {
	rounded := math.Round(effectiveMalus*1e10) / 1e10

	// Walk backwards: the last tier whose Min ≤ rounded wins.
	matched := tiers[0]
	for _, t := range tiers {
		if rounded >= t.Min {
			matched = t
		}
	}
	return matched
}

// EffectiveMalus computes the total effective malus from a pre-filtered list of entries.
// Caller is responsible for filtering entries to a single agent (case-insensitive) and
// expanding allocation lists into per-entry rows before calling this function.
func EffectiveMalus(entries []MalusEntry) float64 {
	today := nowFunc()
	var total float64

	for _, e := range entries {
		// Clamp share to [0, 100].
		share := max(0.0, min(100.0, e.Share))

		var contribution float64
		if e.Decays {
			t, err := time.Parse("2006-01-02", e.Date)
			if err != nil {
				slog.Warn("malus entry has unparseable date, treating as today", "date", e.Date)
				daysSince := 0.0
				// Half-life of 14 days: factor = 0.5^(days/14)
				decayFactor := math.Pow(0.5, daysSince/14)
				contribution = e.RawMalus * (share / 100) * decayFactor
			} else {
				daysSince := today.Sub(t).Hours() / 24
				if daysSince < 0 {
					// Future date → clamp to 0.
					daysSince = 0
				}
				// Half-life of 14 days: factor = 0.5^(days/14)
				decayFactor := math.Pow(0.5, daysSince/14)
				contribution = e.RawMalus * (share / 100) * decayFactor
			}
		} else {
			contribution = e.RawMalus * (share / 100)
		}

		total += contribution
	}
	return total
}

// EligibilityStatus computes eligibility for an agent.
// agentName is accepted for parity with the Python API and future logging/metrics use.
func EligibilityStatus(agentName string, effectiveMalus float64) Status {
	tier := TierFor(effectiveMalus)

	roles := map[string]string{
		"coordinator":       tier.Coordinator,
		"emergency_reserve": tier.EmergencyReserve,
		"specialist":        tier.Specialist,
		"validator":         tier.Validator,
	}

	// Determine overall gate from the role values.
	allClear := true
	allBlocked := true
	for _, v := range roles {
		if v != "CLEAR" {
			allClear = false
		}
		if v != "BLOCKED" {
			allBlocked = false
		}
	}

	var overall string
	switch {
	case allClear:
		overall = "eligible"
	case allBlocked:
		overall = "blocked"
	default:
		overall = "restricted"
	}

	return Status{
		EffectiveMalus: effectiveMalus,
		Tier:           tier.Name,
		Roles:          roles,
		Overall:        overall,
	}
}
