package eligibility_test

import (
	"testing"
	"time"

	"github.com/petersimmons1972/armies/internal/eligibility"
	"github.com/stretchr/testify/assert"
)

func TestTierFor_BoundaryValues(t *testing.T) {
	cases := []struct {
		malus    float64
		wantTier string
	}{
		{0.0, "Clean"},
		{99.9999, "Clean"},
		{100.0, "Warning"},
		{199.9999, "Warning"},
		{200.0, "Probation"},
		{299.9999, "Probation"},
		{300.0, "Demotion risk"},
		{399.9999, "Demotion risk"},
		{400.0, "Suspension"},
		{999.0, "Suspension"},
	}
	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			tier := eligibility.TierFor(tc.malus)
			assert.Equal(t, tc.wantTier, tier.Name, "malus=%.4f", tc.malus)
		})
	}
}

func TestTierFor_FloatPrecision(t *testing.T) {
	// Sum ten 10.0 entries → should equal 100.0 → "Warning" tier (not "Clean" due to float accumulation)
	var sum float64
	for i := 0; i < 10; i++ {
		sum += 10.0
	}
	tier := eligibility.TierFor(sum)
	assert.Equal(t, "Warning", tier.Name)
}

func TestEffectiveMalus_DecayAt14Days(t *testing.T) {
	// Pin the clock to exactly 14 days after the entry
	entryDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	original := eligibility.SetNowFunc(func() time.Time { return entryDate.AddDate(0, 0, 14) })
	defer eligibility.SetNowFunc(original)

	entries := []eligibility.MalusEntry{
		{RawMalus: 100, Share: 100, Date: entryDate.Format("2006-01-02"), Decays: true},
	}
	assert.InDelta(t, 50.0, eligibility.EffectiveMalus(entries), 0.0001)
}

func TestEffectiveMalus_NonDecaying(t *testing.T) {
	entries := []eligibility.MalusEntry{{RawMalus: 100, Share: 50, Decays: false}}
	assert.InDelta(t, 50.0, eligibility.EffectiveMalus(entries), 0.001)
}

func TestEffectiveMalus_UnparseableDate_NoPanic(t *testing.T) {
	// "not-a-date" treated as today → days_since=0 → 0.5^0 = 1.0 → contribution = 100
	entries := []eligibility.MalusEntry{{RawMalus: 100, Share: 100, Date: "not-a-date", Decays: true}}
	assert.InDelta(t, 100.0, eligibility.EffectiveMalus(entries), 0.001)
}

func TestEffectiveMalus_FutureDate_Clamped(t *testing.T) {
	// Future date → days_since=0 → contribution = 100
	future := time.Now().AddDate(0, 0, 14).Format("2006-01-02")
	entries := []eligibility.MalusEntry{{RawMalus: 100, Share: 100, Date: future, Decays: true}}
	assert.InDelta(t, 100.0, eligibility.EffectiveMalus(entries), 0.001)
}

func TestEffectiveMalus_ShareClamping(t *testing.T) {
	// Share 150 → clamped to 100 → contribution = 100 (not 150)
	entries := []eligibility.MalusEntry{{RawMalus: 100, Share: 150, Decays: false}}
	assert.InDelta(t, 100.0, eligibility.EffectiveMalus(entries), 0.001)
}

func TestEligibilityStatus_AllClear(t *testing.T) {
	status := eligibility.EligibilityStatus("clean-agent", 0.0)
	assert.Equal(t, "eligible", status.Overall)
	assert.Equal(t, "CLEAR", status.Roles["coordinator"])
	assert.Equal(t, "Clean", status.Tier)
}

func TestEligibilityStatus_WarningTier(t *testing.T) {
	// 150 → Warning tier
	status := eligibility.EligibilityStatus("warned-agent", 150.0)
	assert.Equal(t, "restricted", status.Overall)
	assert.Equal(t, "BLOCKED", status.Roles["coordinator"])
	assert.Equal(t, "FOUNDER", status.Roles["emergency_reserve"])
	assert.Equal(t, "CLEAR", status.Roles["specialist"])
	assert.Equal(t, "CLEAR", status.Roles["validator"])
}

func TestEligibilityStatus_AllBlocked(t *testing.T) {
	status := eligibility.EligibilityStatus("suspended-agent", 400.0)
	assert.Equal(t, "blocked", status.Overall)
	assert.Equal(t, "BLOCKED", status.Roles["specialist"])
	assert.Equal(t, "BLOCKED", status.Roles["validator"])
}

func TestEligibilityStatus_DemotionRisk(t *testing.T) {
	// 350 → "Demotion risk"
	status := eligibility.EligibilityStatus("risky-agent", 350.0)
	assert.Equal(t, "Demotion risk", status.Tier)
	assert.Equal(t, "ESCALATE", status.Roles["specialist"])
	assert.Equal(t, "CLEAR", status.Roles["validator"])
}

func TestEligibilityStatus_Probation(t *testing.T) {
	// 250 → "Probation"
	status := eligibility.EligibilityStatus("probation-agent", 250.0)
	assert.Equal(t, "Probation", status.Tier)
	assert.Equal(t, "REVIEW", status.Roles["specialist"])
}

func TestEffectiveMalus_Empty(t *testing.T) {
	// No entries → 0.0
	assert.Equal(t, 0.0, eligibility.EffectiveMalus(nil))
}
