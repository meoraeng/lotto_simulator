package main

import "testing"

func TestDetermineRank(t *testing.T) {
	type args struct {
		matchCount int
		hasBonus   bool
	}

	tests := []struct {
		name string
		args args
		want Rank
	}{
		{"1등 - 6개 일치", args{6, false}, Rank1},
		{"2등 - 5개 + 보너스", args{5, true}, Rank2},
		{"3등 - 5개, 보너스 없음", args{5, false}, Rank3},
		{"4등 - 4개 일치", args{4, false}, Rank4},
		{"5등 - 3개 일치", args{3, false}, Rank5},
		{"꽝 - 2개 일치", args{2, false}, RankNone},
		{"꽝 - 0개 일치", args{0, false}, RankNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetermineRank(tt.args.matchCount, tt.args.hasBonus)
			if got != tt.want {
				t.Errorf("DetermineRank(%d, %v) = %v, want %v",
					tt.args.matchCount, tt.args.hasBonus, got, tt.want)
			}
		})
	}
}

func TestRankPrize(t *testing.T) {
	tests := []struct {
		rank Rank
		want int
	}{
		{Rank1, PrizeRank1},
		{Rank2, PrizeRank2},
		{Rank3, PrizeRank3},
		{Rank4, PrizeRank4},
		{Rank5, PrizeRank5},
		{RankNone, 0},
	}

	for _, tt := range tests {
		if got := tt.rank.Prize(); got != tt.want {
			t.Errorf("Rank(%v).Prize() = %d, want %d", tt.rank, got, tt.want)
		}
	}
}
