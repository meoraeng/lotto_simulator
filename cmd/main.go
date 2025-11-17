package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/meoraeng/lotto_simulator/internal/lotto"
)

type playerState struct {
	Player         lotto.Player
	PurchaseAmount int
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== 로또 시뮬레이터 ===")
	fmt.Println()
	// 모드 입력
	mode := readMode(reader)
	fmt.Println()
	// 플레이어 입력
	playerStates := readPlayers(reader)
	fmt.Println()

	totalSales, players := collectPlayers(playerStates)

	var winning lotto.Lottos
	readWinningNumbers(reader, &winning)
	readBonusNumber(reader, &winning)

	base := buildBaseRoundInput(mode, totalSales)

	in := lotto.BuildRoundInput(base, players, winning)
	out := lotto.CalculateRound(in)
	payouts := lotto.DistributeRewards(players, winning, out)

	fmt.Println("\n=== 회차 요약 ===")
	printRoundReport(in, out)

	fmt.Println("\n=== 플레이어별 정산 ===")
	printPlayerPayouts(playerStates, payouts)

	fmt.Println("\n시뮬레이션이 종료되었습니다.")
}
func collectPlayers(states []playerState) (int, []lotto.Player) {
	totalSales := 0
	players := make([]lotto.Player, 0, len(states))

	for _, ps := range states {
		totalSales += ps.PurchaseAmount
		players = append(players, ps.Player)
	}
	return totalSales, players
}

func buildBaseRoundInput(mode lotto.Mode, sales int) lotto.RoundInput {
	allocations := []lotto.Allocation{
		{Rank: lotto.Rank1, BasisPoints: 7500},
		{Rank: lotto.Rank2, BasisPoints: 1250},
		{Rank: lotto.Rank3, BasisPoints: 1250},
		{Rank: lotto.Rank4, BasisPoints: 0},
		{Rank: lotto.Rank5, BasisPoints: 0},
	}

	caps := map[lotto.Rank]int{
		lotto.Rank1: 2_000_000_000,
	}

	var fixedPayout map[lotto.Rank]int
	if mode == lotto.ModeFixedPayout {
		fixedPayout = map[lotto.Rank]int{
			lotto.Rank1: lotto.Rank1.Prize(),
			lotto.Rank2: lotto.Rank2.Prize(),
			lotto.Rank3: lotto.Rank3.Prize(),
			lotto.Rank4: lotto.Rank4.Prize(),
			lotto.Rank5: lotto.Rank5.Prize(),
		}
	}

	return lotto.RoundInput{
		Mode:        mode,
		Sales:       sales,
		CarryIn:     make(map[lotto.Rank]int),
		Allocations: allocations,
		CapPerRank:  caps,
		FixedPayout: fixedPayout,
	}
}
