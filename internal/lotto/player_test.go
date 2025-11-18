package lotto

import "testing"

// 플레이어들의 티켓으로 등수별 당첨자 수가 제대로 집계되는지 테스트
func TestCountWinnersFromPlayers_MultiRanks(t *testing.T) {
	winning := Lottos{
		WinningNumbers: []int{1, 2, 3, 4, 5, 6},
		BonusNumber:    7,
	}

	players := []Player{
		{
			Name: "a",
			Tickets: []Lotto{
				{Numbers: []int{1, 2, 3, 4, 5, 6}}, // 6개 일치 → 1등
				{Numbers: []int{1, 2, 3, 4, 5, 7}}, // 5개 + 보너스 → 2등
			},
		},
		{
			Name: "b",
			Tickets: []Lotto{
				{Numbers: []int{1, 2, 3, 4, 5, 8}},    // 5개 → 3등
				{Numbers: []int{1, 2, 3, 4, 9, 10}},   // 4개 → 4등
				{Numbers: []int{1, 2, 3, 11, 12, 13}}, // 3개 → 5등
			},
		},
	}

	stats := CountWinnersFromPlayers(players, winning)

	if stats[Rank1] != 1 {
		t.Fatalf("1등 당첨자 수가 예상과 다릅니다. got=%d, want=%d", stats[Rank1], 1)
	}
	if stats[Rank2] != 1 {
		t.Fatalf("2등 당첨자 수가 예상과 다릅니다. got=%d, want=%d", stats[Rank2], 1)
	}
	if stats[Rank3] != 1 {
		t.Fatalf("3등 당첨자 수가 예상과 다릅니다. got=%d, want=%d", stats[Rank3], 1)
	}
	if stats[Rank4] != 1 {
		t.Fatalf("4등 당첨자 수가 예상과 다릅니다. got=%d, want=%d", stats[Rank4], 1)
	}
	if stats[Rank5] != 1 {
		t.Fatalf("5등 당첨자 수가 예상과 다릅니다. got=%d, want=%d", stats[Rank5], 1)
	}
}

// DistributeRewards가 플레이어별 수령 금액을 제대로 계산하는지 테스트
func TestDistributeRewards_FixedMode(t *testing.T) {
	// 당첨 패턴은 위와 동일
	winning := Lottos{
		WinningNumbers: []int{1, 2, 3, 4, 5, 6},
		BonusNumber:    7,
	}

	players := []Player{
		{
			Name: "a",
			Tickets: []Lotto{ // 1등 1장
				{Numbers: []int{1, 2, 3, 4, 5, 6}},
			},
		},
		{
			Name: "b",
			Tickets: []Lotto{ // 2,3등 한장씩
				{Numbers: []int{1, 2, 3, 4, 5, 7}},
				{Numbers: []int{1, 2, 3, 4, 5, 8}},
			},
		},
	}

	// 고정 상금 모드 기준 RoundOutput에 등수별 1인당 지급액만 세팅
	out := RoundOutput{
		PaidPerWin: map[Rank]int{
			Rank1: Rank1.Prize(),
			Rank2: Rank2.Prize(),
			Rank3: Rank3.Prize(),
		},
	}

	rewards := DistributeRewards(players, winning, out)

	wantA := Rank1.Prize()
	if rewards["a"] != wantA {
		t.Fatalf("플레이어 a 수령 금액이 예상과 다릅니다. got=%d, want=%d", rewards["a"], wantA)
	}

	wantB := Rank2.Prize() + Rank3.Prize()
	if rewards["b"] != wantB {
		t.Fatalf("플레이어 b 수령 금액이 예상과 다릅니다. got=%d, want=%d", rewards["b"], wantB)
	}
}
