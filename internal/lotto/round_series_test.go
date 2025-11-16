package lotto

import "testing"

// 이월 금액이 회차 사이에서 제대로 흘러가는지 검증
func TestSimulateSeries_CarryFlowsBetweenRounds(t *testing.T) {
	t.Helper()

	cfg := SeriesConfig{
		Mode: ModeParimutuel,
		Allocations: []Allocation{
			// 1등에게 판매액 100% 할당
			{Rank: Rank1, BasisPoints: 10_000},
		},
		CapPerRank: map[Rank]int{},
	}

	sales := []int{ // 1,2회차 판매액
		1_000_000,
		1_000_000,
	}

	winners := []map[Rank]int{
		// 1회차: 1등 당첨자 0명 → 풀 전액 이월 기대
		{},
		// 2회차: 1등 당첨자 1명 → (1회차 이월 + 2회차 풀) 전부 가져감
		{
			Rank1: 1,
		},
	}

	results, err := SimulateSeries(cfg, sales, winners, nil)
	if err != nil {
		t.Fatalf("시리즈 시뮬레이션 중 에러가 발생했습니다. err=%v", err)
	}

	if len(results) != 2 {
		t.Fatalf("결과 회차 수가 일치하지 않습니다. got=%d, want=%d", len(results), 2)
	}

	round1 := results[0]
	round2 := results[1]

	// 1회차 검증
	if round1.PoolBefore[Rank1] != 1_000_000 {
		t.Fatalf("1회차 1등 풀 금액이 잘못되었습니다. got=%d, want=%d",
			round1.PoolBefore[Rank1], 1_000_000)
	}
	if round1.PaidPerWin[Rank1] != 0 {
		t.Fatalf("1회차 1등 1인당 지급액은 0이어야 합니다. got=%d",
			round1.PaidPerWin[Rank1])
	}
	if round1.CarryOut[Rank1] != 1_000_000 {
		t.Fatalf("1회차 1등 이월 금액이 잘못되었습니다. got=%d, want=%d",
			round1.CarryOut[Rank1], 1_000_000)
	}

	// 2회차 검증: 1회차 이월 + 2회차 판매 풀 = 1000,000 + 1000,000 = 2,000,000
	if round2.PoolBefore[Rank1] != 2_000_000 {
		t.Fatalf("2회차 1등 풀 금액이 잘못되었습니다. got=%d, want=%d",
			round2.PoolBefore[Rank1], 2_000_000)
	}
	if round2.PaidPerWin[Rank1] != 2_000_000 {
		t.Fatalf("2회차 1등 1인당 지급액이 잘못되었습니다. got=%d, want=%d",
			round2.PaidPerWin[Rank1], 2_000_000)
	}
	if round2.CarryOut[Rank1] != 0 {
		t.Fatalf("2회차 1등 이월 금액은 0이어야 합니다. got=%d",
			round2.CarryOut[Rank1])
	}
}

// 판매 금액, 당첨자 길이가 다르면 에러 처리
func TestSimulateSeries_LengthMismatchReturnsError(t *testing.T) {
	t.Helper()

	cfg := SeriesConfig{
		Mode:        ModeParimutuel,
		Allocations: []Allocation{},
		CapPerRank:  map[Rank]int{},
	}

	sales := []int{1_000_000}
	winners := []map[Rank]int{
		{},
		{},
	}

	_, err := SimulateSeries(cfg, sales, winners, nil)
	if err == nil {
		t.Fatalf("입력 길이가 다를 때 에러가 발생해야 합니다.")
	}

	// 에러 타입이 UserInputError인지 정도만 확인 (메시지 전체 비교 X)
	if _, ok := err.(UserInputError); !ok {
		t.Fatalf("UserInputError 타입의 에러가 발생해야 합니다. got=%T", err)
	}
}
