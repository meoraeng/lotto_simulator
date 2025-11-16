package lotto

import "testing"

func TestCalculateRound_BasicParimutuel(t *testing.T) {
	in := RoundInput{
		Mode:  ModeParimutuel,
		Sales: 1_000_000,
		Winners: map[Rank]int{
			Rank1: 1,
			Rank2: 2,
			Rank3: 4,
		},
		Allocations: []Allocation{
			{Rank: Rank1, BasisPoints: 7_500}, // 75%
			{Rank: Rank2, BasisPoints: 1_250}, // 12.5%
			{Rank: Rank3, BasisPoints: 1_250}, // 12.5%
		},
		// CarryIn, CapPerRank는 기본값 nil -> 모두 0 취급
	}

	out := CalculateRound(in)

	// PoolBefore / PoolAfterCap 체크
	if got := out.PoolBefore[Rank1]; got != 750_000 {
		t.Errorf("PoolBefore[Rank1] = %d, want %d", got, 750_000)
	}
	if got := out.PoolBefore[Rank2]; got != 125_000 {
		t.Errorf("PoolBefore[Rank2] = %d, want %d", got, 125_000)
	}
	if got := out.PoolBefore[Rank3]; got != 125_000 {
		t.Errorf("PoolBefore[Rank3] = %d, want %d", got, 125_000)
	}
	// 상한이 없으면 AfterCap == Before 이어야 함
	if got := out.PoolAfterCap[Rank1]; got != 750_000 {
		t.Errorf("PoolAfterCap[Rank1] = %d, want %d", got, 750_000)
	}

	// PaidPerWin / PaidTotal 체크
	if got := out.PaidPerWin[Rank1]; got != 750_000 {
		t.Errorf("PaidPerWin[Rank1] = %d, want %d", got, 750_000)
	}
	if got := out.PaidPerWin[Rank2]; got != 62_500 {
		t.Errorf("PaidPerWin[Rank2] = %d, want %d", got, 62_500)
	}
	if got := out.PaidPerWin[Rank3]; got != 31_250 {
		t.Errorf("PaidPerWin[Rank3] = %d, want %d", got, 31_250)
	}

	// 잔액 이월 없음
	if got := out.CarryOut[Rank1]; got != 0 {
		t.Errorf("CarryOut[Rank1] = %d, want 0", got)
	}
	if got := out.CarryOut[Rank2]; got != 0 {
		t.Errorf("CarryOut[Rank2] = %d, want 0", got)
	}
	if got := out.CarryOut[Rank3]; got != 0 {
		t.Errorf("CarryOut[Rank3] = %d, want 0", got)
	}
}

func TestCalculateRound_WithCapAndProportionalRolldown(t *testing.T) {
	in := RoundInput{
		Mode:  ModeParimutuel,
		Sales: 1_000_000,
		Winners: map[Rank]int{
			Rank1: 2,
			Rank2: 3,
			Rank3: 4,
		},
		Allocations: []Allocation{
			{Rank: Rank1, BasisPoints: 5_000}, // 50%
			{Rank: Rank2, BasisPoints: 3_000}, // 30%
			{Rank: Rank3, BasisPoints: 2_000}, // 20%
		},
		CapPerRank: map[Rank]int{
			Rank1: 400_000, // 1등은 최대 40만까지
		},
	}

	out := CalculateRound(in)

	// 기본 풀: 1등 500_000, 2등 300_000, 3등 200_000
	// 1등 캡: 400_000, overflow: 100_000
	// 하위 비율: 2등 3000, 3등 2000 -> 총 5000
	// → 2등 60_000, 3등 40_000
	// 최종: 1등 400_000, 2등 360_000, 3등 240_000

	if got := out.PoolAfterCap[Rank1]; got != 400_000 {
		t.Errorf("PoolAfterCap[Rank1] = %d, want %d", got, 400_000)
	}
	if got := out.PoolAfterCap[Rank2]; got != 360_000 {
		t.Errorf("PoolAfterCap[Rank2] = %d, want %d", got, 360_000)
	}
	if got := out.PoolAfterCap[Rank3]; got != 240_000 {
		t.Errorf("PoolAfterCap[Rank3] = %d, want %d", got, 240_000)
	}

	// 지급액: 1등 2명 / 2등 3명 / 3등 4명
	if got := out.PaidPerWin[Rank1]; got != 200_000 {
		t.Errorf("PaidPerWin[Rank1] = %d, want %d", got, 200_000)
	}
	if got := out.PaidPerWin[Rank2]; got != 120_000 {
		t.Errorf("PaidPerWin[Rank2] = %d, want %d", got, 120_000)
	}
	if got := out.PaidPerWin[Rank3]; got != 60_000 {
		t.Errorf("PaidPerWin[Rank3] = %d, want %d", got, 60_000)
	}

	// 이월 없음
	if got := out.CarryOut[Rank1]; got != 0 {
		t.Errorf("CarryOut[Rank1] = %d, want 0", got)
	}
	if got := out.CarryOut[Rank2]; got != 0 {
		t.Errorf("CarryOut[Rank2] = %d, want 0", got)
	}
	if got := out.CarryOut[Rank3]; got != 0 {
		t.Errorf("CarryOut[Rank3] = %d, want 0", got)
	}
}

func TestCalculateRound_RolldownWithoutLowerAllocations(t *testing.T) {
	in := RoundInput{
		Mode:  ModeParimutuel,
		Sales: 1_000_000,
		Winners: map[Rank]int{
			Rank5: 4,
		},
		Allocations: []Allocation{
			{Rank: Rank1, BasisPoints: 10_000}, // 전부 1등에 몰아줌
			// 나머지 등수는 0
		},
		CapPerRank: map[Rank]int{
			Rank1: 600_000,
		},
	}

	out := CalculateRound(in)

	// 기본 풀: 1등 1,000,000 / 나머지 0
	// 상한 1등 600,000, overflow = 400,000
	// 하위 등수 비율 합이 0 -> Rank5에 전부 몰아주기
	if got := out.PoolAfterCap[Rank1]; got != 600_000 {
		t.Errorf("PoolAfterCap[Rank1] = %d, want %d", got, 600_000)
	}
	if got := out.PoolAfterCap[Rank5]; got != 400_000 {
		t.Errorf("PoolAfterCap[Rank5] = %d, want %d", got, 400_000)
	}

	// Rank5 4명 -> 1인당 100,000
	if got := out.PaidPerWin[Rank5]; got != 100_000 {
		t.Errorf("PaidPerWin[Rank5] = %d, want %d", got, 100_000)
	}
	if got := out.PaidTotal[Rank5]; got != 400_000 {
		t.Errorf("PaidTotal[Rank5] = %d, want %d", got, 400_000)
	}
	if got := out.CarryOut[Rank5]; got != 0 {
		t.Errorf("CarryOut[Rank5] = %d, want 0", got)
	}
}

func TestCalculateRound_NoWinners_CarryOverAll(t *testing.T) {
	in := RoundInput{
		Mode:  ModeParimutuel,
		Sales: 500_000,
		Winners: map[Rank]int{
			Rank1: 0, // 1등 당첨자 없음
		},
		Allocations: []Allocation{
			{Rank: Rank1, BasisPoints: 10_000},
		},
		CarryIn: map[Rank]int{
			Rank1: 100_000, // 이월 있으면 같이 합쳐져야 함
		},
	}

	out := CalculateRound(in)

	// PoolBefore = Sales + CarryIn = 600,000
	if got := out.PoolBefore[Rank1]; got != 600_000 {
		t.Errorf("PoolBefore[Rank1] = %d, want %d", got, 600_000)
	}
	// 당첨자 없음 → 전액 이월
	if got := out.CarryOut[Rank1]; got != 600_000 {
		t.Errorf("CarryOut[Rank1] = %d, want %d", got, 600_000)
	}
	// 지급액은 0
	if got := out.PaidTotal[Rank1]; got != 0 {
		t.Errorf("PaidTotal[Rank1] = %d, want 0", got)
	}
}

func TestCalculateRound_FixedPayoutMode_IsNoopForNow(t *testing.T) {
	in := RoundInput{
		Mode:  ModeFixedPayout,
		Sales: 1_000_000,
		Winners: map[Rank]int{
			Rank1: 1,
		},
		Allocations: []Allocation{
			{Rank: Rank1, BasisPoints: 10_000},
		},
	}

	out := CalculateRound(in)

	// 비어있는 상태 가정
	if out.Sales != in.Sales {
		t.Errorf("Sales = %d, want %d", out.Sales, in.Sales)
	}
	if len(out.PoolBefore) != 0 {
		t.Errorf("PoolBefore len = %d, want 0", len(out.PoolBefore))
	}
	if len(out.PaidTotal) != 0 {
		t.Errorf("PaidTotal len = %d, want 0", len(out.PaidTotal))
	}
}
