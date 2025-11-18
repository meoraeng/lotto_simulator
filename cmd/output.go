package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/meoraeng/lotto_simulator/internal/lotto"
)

// -------------------- 출력 처리 --------------------

func formatNumbers(nums []int) string {
	var b strings.Builder
	b.WriteString("[")
	for i, n := range nums {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(strconv.Itoa(n))
	}
	b.WriteString("]")
	return b.String()
}

func printRoundReport(in lotto.RoundInput, out lotto.RoundOutput) {
	fmt.Printf("총 판매액: %d원\n", in.Sales)
	fmt.Printf("라운드 잔액(RoundRemainder): %d원\n\n", out.RoundRemainder)

	switch in.Mode {
	case lotto.ModeParimutuel:
		printParimutuelReport(in, out)
	case lotto.ModeFixedPayout:
		printFixedPayoutReport(in, out)
	default:
		fmt.Println("(알 수 없는 모드입니다)")
	}
}

func printParimutuelReport(in lotto.RoundInput, out lotto.RoundOutput) {
	// 헤더
	fmt.Printf(
		"%4s | %10s | %10s | %8s | %10s | %12s | %12s | %10s\n",
		"Rank", "PoolB", "PoolA", "RDown", "WinCnt", "PerWin", "Total", "Carry",
	)
	fmt.Println(strings.Repeat("-", 100))

	order := []lotto.Rank{lotto.Rank1, lotto.Rank2, lotto.Rank3, lotto.Rank4, lotto.Rank5}

	for i, r := range order {
		rankNo := i + 1

		fmt.Printf(
			"%4d | %10d | %10d | %8d | %10d | %12d | %12d | %10d\n",
			rankNo,
			out.PoolBefore[r],
			out.PoolAfterCap[r],
			out.RollDown[r],
			in.Winners[r],
			out.PaidPerWin[r],
			out.PaidTotal[r],
			out.CarryOut[r],
		)
	}
}

func printFixedPayoutReport(in lotto.RoundInput, out lotto.RoundOutput) {
	fmt.Printf(
		"%4s | %10s | %12s | %12s\n",
		"Rank", "WinCnt", "PerWin", "Total",
	)
	fmt.Println(strings.Repeat("-", 48))

	order := []lotto.Rank{lotto.Rank1, lotto.Rank2, lotto.Rank3, lotto.Rank4, lotto.Rank5}

	for i, r := range order {
		rankNo := i + 1

		fmt.Printf(
			"%4d | %10d | %12d | %12d\n",
			rankNo,
			in.Winners[r],
			out.PaidPerWin[r],
			out.PaidTotal[r],
		)
	}
}

func printPlayerPayouts(states []playerState, payouts map[string]int) {
	for _, ps := range states {
		name := ps.Player.Name
		earned := payouts[name]
		spent := ps.PurchaseAmount

		var profitRate float64
		if spent > 0 {
			profitRate = float64(earned) / float64(spent) * 100
		}

		fmt.Printf("%s: 사용 금액 %d원, 수령 금액 %d원, 수익률 %.1f%%\n",
			name, spent, earned, profitRate)
	}
}

func rankLabel(r lotto.Rank) string {
	switch r {
	case lotto.Rank1:
		return "1등"
	case lotto.Rank2:
		return "2등"
	case lotto.Rank3:
		return "3등"
	case lotto.Rank4:
		return "4등"
	case lotto.Rank5:
		return "5등"
	default:
		return "-"
	}
}

func printPlayerTotals(states []playerState, totals map[string]int) {
	for _, ps := range states {
		name := ps.Player.Name
		spent := ps.PurchaseAmount
		earned := totals[name]

		rate := 0.0
		if spent > 0 {
			rate = float64(earned) / float64(spent) * 100
		}

		fmt.Printf("%s: 사용 금액 %d원, 누적 수령 금액 %d원, 누적 수익률 %.1f%%\n",
			name, spent, earned, rate)
	}
}
