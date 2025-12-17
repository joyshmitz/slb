package core

import (
	"testing"

	"github.com/Dicklesworthstone/slb/internal/db"
)

func TestTierHigher(t *testing.T) {
	tests := []struct {
		name   string
		tier1  db.RiskTier
		tier2  db.RiskTier
		expect bool
	}{
		{"critical > dangerous", db.RiskTierCritical, db.RiskTierDangerous, true},
		{"critical > caution", db.RiskTierCritical, db.RiskTierCaution, true},
		{"dangerous > caution", db.RiskTierDangerous, db.RiskTierCaution, true},
		{"dangerous < critical", db.RiskTierDangerous, db.RiskTierCritical, false},
		{"caution < critical", db.RiskTierCaution, db.RiskTierCritical, false},
		{"caution < dangerous", db.RiskTierCaution, db.RiskTierDangerous, false},
		{"same tier critical", db.RiskTierCritical, db.RiskTierCritical, false},
		{"same tier dangerous", db.RiskTierDangerous, db.RiskTierDangerous, false},
		{"same tier caution", db.RiskTierCaution, db.RiskTierCaution, false},
		{"unknown tier1", db.RiskTier("unknown"), db.RiskTierCaution, false},
		{"unknown tier2", db.RiskTierCaution, db.RiskTier("unknown"), true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tierHigher(tc.tier1, tc.tier2)
			if result != tc.expect {
				t.Errorf("tierHigher(%q, %q) = %v, want %v", tc.tier1, tc.tier2, result, tc.expect)
			}
		})
	}
}
