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

	rounds := readRoundCount(reader)
	fmt.Println()

	// 플레이어 입력
	playerStates := readPlayers(reader)
	fmt.Println()

	totalSales, players := collectPlayers(playerStates)

	// 회차 간 이월 상태
	carry := make(map[lotto.Rank]int)

	// 플레이어별 누적 수령액
	totalPayouts := make(map[string]int)

	for round := 1; round <= rounds; round++ {
		fmt.Printf("\n=== %d회차 ===\n", round)

		// 당첨 번호 / 보너스 번호 입력
		var winning lotto.Lottos
		readWinningNumbers(reader, &winning)
		readBonusNumber(reader, &winning)

		// 이번 회차 입력값 구성 (판매액 + 이월 상태 포함)
		base := buildBaseRoundInput(mode, totalSales, carry)
		in := lotto.BuildRoundInput(base, players, winning)

		// 분배 계산
		out, err := lotto.CalculateRound(in)
		if err != nil {
			printError(fmt.Errorf("계산 중 오류 발생: %w", err))
			return
		}

		// 플레이어별 이번 회차 수령액 계산
		payouts := lotto.DistributeRewards(players, winning, out)

		// 회차 요약 출력
		fmt.Println("\n--- 회차 요약 ---")
		printRoundReport(in, out)

		// 플레이어별 이번 회차 정산
		fmt.Println("\n--- 플레이어별 정산 (이번 회차) ---")
		printPlayerPayouts(playerStates, payouts)

		// 누적 수령액에 합산
		for name, amount := range payouts {
			totalPayouts[name] += amount
		}

		// 다음 회차를 위해 이월 상태 업데이트
		carry = out.CarryOut
	}

	// 2회 이상 돌렸으면 누적 정산도 보여주기 (선택)
	if rounds > 1 {
		fmt.Println("\n=== 전체 누적 정산 ===")
		printPlayerTotals(playerStates, totalPayouts)
	}

	fmt.Println("\n시뮬레이션이 종료되었습니다.")
}

// 전체 판매액 합계와 Player리스트 생성
func collectPlayers(states []playerState) (int, []lotto.Player) {
	totalSales := 0
	players := make([]lotto.Player, 0, len(states))

	for _, ps := range states {
		totalSales += ps.PurchaseAmount
		players = append(players, ps.Player)
	}
	return totalSales, players
}

func buildBaseRoundInput(
	mode lotto.Mode,
	sales int,
	carryIn map[lotto.Rank]int,
) lotto.RoundInput {
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
		Mode:         mode,
		Sales:        sales,
		CarryIn:      carryIn,
		Allocations:  allocations,
		CapPerRank:   caps,
		RoundingUnit: 100,
		FixedPayout:  fixedPayout,
	}
}
