package lotto

import "testing"

func TestCalculateRound_Parimutuel_BasicDistribution(t *testing.T) {
	t.Helper()

	type args struct {
		sales   int
		winners map[Rank]int
		carryIn map[Rank]int
		allocs  []Allocation
	}

	tests := []struct {
		name        string
		args        args
		wantPool1   int
		wantPerWin1 int
		wantCarry1  int
		wantPool2   int
		wantPerWin2 int
		wantCarry2  int
	}{
		{
			name: "1등 1명, 2등 2명 - 판매액 1,000,000, 1등 75%, 2등 25%",
			args: args{
				sales: 1_000_000,
				winners: map[Rank]int{
					Rank1: 1,
					Rank2: 2,
				},
				carryIn: map[Rank]int{},
				allocs: []Allocation{
					{Rank: Rank1, BasisPoints: 7_500}, // 75%
					{Rank: Rank2, BasisPoints: 2_500}, // 25%
				},
			},

			wantPool1:   750_000,
			wantPerWin1: 750_000,
			wantCarry1:  0,

			wantPool2:   250_000,
			wantPerWin2: 125_000,
			wantCarry2:  0,
		},
		{
			name: "2등 당첨자 0명인 경우 전액 이월",
			args: args{
				sales: 500_000,
				winners: map[Rank]int{
					Rank1: 1,
					// Rank2 0명
				},
				carryIn: map[Rank]int{},
				allocs: []Allocation{
					{Rank: Rank1, BasisPoints: 5_000}, // 50%
					{Rank: Rank2, BasisPoints: 5_000}, // 50%
				},
			},
			wantPool1:   250_000,
			wantPerWin1: 250_000,
			wantCarry1:  0,

			wantPool2:   250_000,
			wantPerWin2: 0,
			wantCarry2:  250_000,
		},
		{
			name: "이월 금액이 다음 회차 풀에 합산되는 경우",
			args: args{
				sales: 400_000,
				winners: map[Rank]int{
					Rank1: 2,
				},
				carryIn: map[Rank]int{
					Rank1: 100_000, // 이전 회차에서 이월된 금액
				},
				allocs: []Allocation{
					{Rank: Rank1, BasisPoints: 10_000}, // 100% 전부 1등 풀
				},
			},
			// 이번 회차 풀: 400,000 + 이월 100,000 = 500,000 -> 2인 분배 250,000
			wantPool1:   500_000,
			wantPerWin1: 250_000,
			wantCarry1:  0,
			// 2등은 배정/당첨자 모두 없음 -> 기대값 0
			wantPool2:   0,
			wantPerWin2: 0,
			wantCarry2:  0,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			in := RoundInput{
				Mode:        ModeParimutuel,
				Sales:       tt.args.sales,
				Winners:     tt.args.winners,
				CarryIn:     tt.args.carryIn,
				Allocations: tt.args.allocs,
				CapPerRank:  map[Rank]int{},
			}

			got := CalculateRound(in)

			if got.PoolBefore[Rank1] != tt.wantPool1 {
				t.Fatalf("1등 기본 풀 금액이 일치하지 않습니다. got=%d, want=%d",
					got.PoolBefore[Rank1], tt.wantPool1)
			}
			if got.PaidPerWin[Rank1] != tt.wantPerWin1 {
				t.Fatalf("1등 1인당 지급액이 일치하지 않습니다. got=%d, want=%d",
					got.PaidPerWin[Rank1], tt.wantPerWin1)
			}
			if got.CarryOut[Rank1] != tt.wantCarry1 {
				t.Fatalf("1등 이월 금액이 일치하지 않습니다. got=%d, want=%d",
					got.CarryOut[Rank1], tt.wantCarry1)
			}

			if got.PoolBefore[Rank2] != tt.wantPool2 {
				t.Fatalf("2등 기본 풀 금액이 일치하지 않습니다. got=%d, want=%d",
					got.PoolBefore[Rank2], tt.wantPool2)
			}
			if got.PaidPerWin[Rank2] != tt.wantPerWin2 {
				t.Fatalf("2등 1인당 지급액이 일치하지 않습니다. got=%d, want=%d",
					got.PaidPerWin[Rank2], tt.wantPerWin2)
			}
			if got.CarryOut[Rank2] != tt.wantCarry2 {
				t.Fatalf("2등 이월 금액이 일치하지 않습니다. got=%d, want=%d",
					got.CarryOut[Rank2], tt.wantCarry2)
			}
		})
	}
}

func TestCalculateRound_NonParimutuelMode_ReturnsEmptyResult(t *testing.T) {
	t.Helper()

	in := RoundInput{
		Mode:  ModeFixedPayout,
		Sales: 1_000_000,
		Winners: map[Rank]int{
			Rank1: 3,
		},
		CarryIn: map[Rank]int{
			Rank1: 100_000,
		},
		Allocations: []Allocation{
			{Rank: Rank1, BasisPoints: 10_000},
		},
		CapPerRank: map[Rank]int{
			Rank1: 1_000_000,
		},
	}

	got := CalculateRound(in)

	// 현재는 빈 출력만 반환하는 것으로 가정
	if got.PoolBefore[Rank1] != 0 {
		t.Fatalf("고정 상금 모드에서는 PoolBefore가 0이어야 합니다. got=%d", got.PoolBefore[Rank1])
	}
	if got.PaidPerWin[Rank1] != 0 {
		t.Fatalf("고정 상금 모드에서는 PaidPerWin이 0이어야 합니다. got=%d", got.PaidPerWin[Rank1])
	}
	if got.CarryOut[Rank1] != 0 {
		t.Fatalf("고정 상금 모드에서는 CarryOut이 0이어야 합니다. got=%d", got.CarryOut[Rank1])
	}
}
