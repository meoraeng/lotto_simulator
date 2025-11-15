package lotto

import "testing"

func TestDetermineRank(t *testing.T) {
	tests := []struct {
		name       string
		matchCount int
		hasBonus   bool
		want       Rank
	}{
		{"1등 - 6개 일치", 6, false, Rank1},
		{"2등 - 5개 + 보너스", 5, true, Rank2},
		{"3등 - 5개", 5, false, Rank3},
		{"4등 - 4개", 4, false, Rank4},
		{"5등 - 3개", 3, false, Rank5},
		{"꽝 - 2개 이하", 2, false, RankNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetermineRank(tt.matchCount, tt.hasBonus)
			if got != tt.want {
				t.Errorf("등수 판정 실패: 입력(%d, %v) → 결과 %v, 기대값 %v",
					tt.matchCount, tt.hasBonus, got, tt.want)
			}
		})
	}
}
