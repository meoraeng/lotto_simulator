package lotto

import "testing"

// 분배 모드에서 기본 분배 동작 검증
func TestCalculateRound_DistributeMode_BasicDistribute(t *testing.T) {
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
		CarryIn:      map[Rank]int{},
		CapPerRank:   map[Rank]int{},
		RoundingUnit: 1, // 라운딩 없이 1원 단위
	}

	out, err := CalculateRound(in)
	if err != nil {
		t.Fatalf("라운드 계산 중 에러가 발생했습니다: %v", err)
	}

	// PoolBefore 체크
	if got := out.PoolBefore[Rank1]; got != 750_000 {
		t.Errorf("1등 풀(전) 값이 예상과 다릅니다. got=%d, want=%d", got, 750_000)
	}
	if got := out.PoolBefore[Rank2]; got != 125_000 {
		t.Errorf("2등 풀(전) 값이 예상과 다릅니다. got=%d, want=%d", got, 125_000)
	}
	if got := out.PoolBefore[Rank3]; got != 125_000 {
		t.Errorf("3등 풀(전) 값이 예상과 다릅니다. got=%d, want=%d", got, 125_000)
	}

	// 상한이 없는 경우 AfterCap == Before
	if got := out.PoolAfterCap[Rank1]; got != 750_000 {
		t.Errorf("1등 풀(후) 값이 예상과 다릅니다. got=%d, want=%d", got, 750_000)
	}

	// 인당 지급액 / 총 지급액 체크
	if got := out.PaidPerWin[Rank1]; got != 750_000 {
		t.Errorf("1등 1인당 지급액이 예상과 다릅니다. got=%d, want=%d", got, 750_000)
	}
	if got := out.PaidPerWin[Rank2]; got != 62_500 {
		t.Errorf("2등 1인당 지급액이 예상과 다릅니다. got=%d, want=%d", got, 62_500)
	}
	if got := out.PaidPerWin[Rank3]; got != 31_250 {
		t.Errorf("3등 1인당 지급액이 예상과 다릅니다. got=%d, want=%d", got, 31_250)
	}

	// 이월 없음
	if got := out.CarryOut[Rank1]; got != 0 {
		t.Errorf("1등 이월 금액이 0이 아닙니다. got=%d", got)
	}
	if got := out.CarryOut[Rank2]; got != 0 {
		t.Errorf("2등 이월 금액이 0이 아닙니다. got=%d", got)
	}
	if got := out.CarryOut[Rank3]; got != 0 {
		t.Errorf("3등 이월 금액이 0이 아닙니다. got=%d", got)
	}
}

// 상한 + 비례 롤다운이 정상 작동 검증
func TestCalculateRound_분배모드_상한과비례롤다운(t *testing.T) {
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
		CarryIn:      map[Rank]int{},
		RoundingUnit: 1,
	}

	out, err := CalculateRound(in)
	if err != nil {
		t.Fatalf("라운드 계산 중 에러가 발생했습니다: %v", err)
	}

	// 기본 풀: 1등 500_000, 2등 300_000, 3등 200_000
	// 1등 캡: 400_000, overflow: 100_000
	// 하위 비율: 2등 3000, 3등 2000 -> 총 5000
	// -> 2등 60_000, 3등 40_000
	// 최종: 1등 400_000, 2등 360_000, 3등 240_000
	if got := out.PoolAfterCap[Rank1]; got != 400_000 {
		t.Errorf("1등 풀(후) 값이 예상과 다릅니다. got=%d, want=%d", got, 400_000)
	}
	if got := out.PoolAfterCap[Rank2]; got != 360_000 {
		t.Errorf("2등 풀(후) 값이 예상과 다릅니다. got=%d, want=%d", got, 360_000)
	}
	if got := out.PoolAfterCap[Rank3]; got != 240_000 {
		t.Errorf("3등 풀(후) 값이 예상과 다릅니다. got=%d, want=%d", got, 240_000)
	}

	if got := out.PaidPerWin[Rank1]; got != 200_000 {
		t.Errorf("1등 1인당 지급액이 예상과 다릅니다. got=%d, want=%d", got, 200_000)
	}
	if got := out.PaidPerWin[Rank2]; got != 120_000 {
		t.Errorf("2등 1인당 지급액이 예상과 다릅니다. got=%d, want=%d", got, 120_000)
	}
	if got := out.PaidPerWin[Rank3]; got != 60_000 {
		t.Errorf("3등 1인당 지급액이 예상과 다릅니다. got=%d, want=%d", got, 60_000)
	}

	if got := out.CarryOut[Rank1]; got != 0 {
		t.Errorf("1등 이월 금액이 0이 아닙니다. got=%d", got)
	}
	if got := out.CarryOut[Rank2]; got != 0 {
		t.Errorf("2등 이월 금액이 0이 아닙니다. got=%d", got)
	}
	if got := out.CarryOut[Rank3]; got != 0 {
		t.Errorf("3등 이월 금액이 0이 아닙니다. got=%d", got)
	}
}

// 분배 모드에서 내림이 제대로 적용되는지 검증
func TestCalculateRound_DistributeMode_RoundingValidate(t *testing.T) {
	in := RoundInput{
		Mode:  ModeParimutuel,
		Sales: 10_005, // 애매한 금액
		Winners: map[Rank]int{
			Rank1: 1,
		},
		CarryIn: map[Rank]int{},
		Allocations: []Allocation{
			{Rank: Rank1, BasisPoints: 10_000}, // 100% 1등
		},
		CapPerRank:   map[Rank]int{},
		RoundingUnit: 100, // 100원 단위 내림
	}

	out, err := CalculateRound(in)
	if err != nil {
		t.Fatalf("라운드 계산 중 에러가 발생했습니다: %v", err)
	}

	// 풀 자체는 10,005
	if out.PoolBefore[Rank1] != 10_005 {
		t.Fatalf("1등 풀(전) 값이 예상과 다릅니다. got=%d, want=%d", out.PoolBefore[Rank1], 10_005)
	}

	// 1인당 지급액 예상 10,005 -> 100원 단위 내림 -> 10,000
	if out.PaidPerWin[Rank1] != 10_000 {
		t.Fatalf("라운딩된 1인당 지급액이 예상과 다릅니다. got=%d, want=%d", out.PaidPerWin[Rank1], 10_000)
	}
	if out.PaidTotal[Rank1] != 10_000 {
		t.Fatalf("라운딩된 총 지급액이 예상과 다릅니다. got=%d, want=%d", out.PaidTotal[Rank1], 10_000)
	}

	// 잔액 5원은 이월되어야 함
	if out.CarryOut[Rank1] != 5 {
		t.Fatalf("라운딩 후 이월되는 잔액이 예상과 다릅니다. got=%d, want=%d", out.CarryOut[Rank1], 5)
	}
}

// 당첨자가 없을 때 전액 이월되는지 검증
func TestCalculateRound_Distribute_NoWinnner_AllCarry(t *testing.T) {
	in := RoundInput{
		Mode:  ModeParimutuel,
		Sales: 500_000,
		Winners: map[Rank]int{
			Rank1: 0, // 1등 당첨자 없음
		},
		Allocations: []Allocation{
			{Rank: Rank1, BasisPoints: 10_000},
		},
		CarryIn:      map[Rank]int{Rank1: 100_000},
		CapPerRank:   map[Rank]int{},
		RoundingUnit: 1,
	}

	out, err := CalculateRound(in)
	if err != nil {
		t.Fatalf("라운드 계산 중 에러가 발생했습니다: %v", err)
	}

	// PoolBefore = Sales + CarryIn = 600,000
	if out.PoolBefore[Rank1] != 600_000 {
		t.Errorf("1등 풀(전) 값이 예상과 다릅니다. got=%d, want=%d", out.PoolBefore[Rank1], 600_000)
	}
	if out.CarryOut[Rank1] != 600_000 {
		t.Errorf("1등 이월 금액이 예상과 다릅니다. got=%d, want=%d", out.CarryOut[Rank1], 600_000)
	}
	if out.PaidTotal[Rank1] != 0 {
		t.Errorf("1등 지급액이 0이 아닙니다. got=%d", out.PaidTotal[Rank1])
	}
}

// 고정 상금 모드에서 여러 명이 당첨되었을 때 검증(1인당 상금 고정이고, 총 지급액은 고정 상금 × 인원 수)
func TestCalculateRound_FixedMode_MultiWinners(t *testing.T) {
	in := RoundInput{
		Mode:  ModeFixedPayout,
		Sales: 0, // 고정 모드에서는 Sales와 무관하게 상금표 기준
		Winners: map[Rank]int{
			Rank5: 3, // 5등 당첨자 3명
		},
		FixedPayout: map[Rank]int{
			Rank5: Rank5.Prize(),
		},
	}

	out, err := CalculateRound(in)
	if err != nil {
		t.Fatalf("라운드 계산 중 에러가 발생했습니다: %v", err)
	}

	if out.PaidPerWin[Rank5] != Rank5.Prize() {
		t.Fatalf("5등 1인당 지급액이 예상과 다릅니다. got=%d, want=%d",
			out.PaidPerWin[Rank5], Rank5.Prize())
	}

	wantTotal := Rank5.Prize() * 3
	if out.PaidTotal[Rank5] != wantTotal {
		t.Fatalf("5등 총 지급액이 예상과 다릅니다. got=%d, want=%d",
			out.PaidTotal[Rank5], wantTotal)
	}
}
