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
	order := []lotto.Rank{lotto.Rank1, lotto.Rank2, lotto.Rank3, lotto.Rank4, lotto.Rank5}

	fmt.Printf("총 판매액: %d원\n", out.Sales)
	fmt.Printf("라운드 잔액(RoundRemainder): %d원\n", out.RoundRemainder)
	fmt.Println()
	fmt.Println("등수 | 풀(전) | 풀(후) | 롤다운 | 당첨자 수 | 인당 지급액 | 총 지급액 | 이월")
	fmt.Println("--------------------------------------------------------------------------")

	for _, r := range order {
		label := rankLabel(r)
		winners := in.Winners[r]
		poolBefore := out.PoolBefore[r]
		poolAfter := out.PoolAfterCap[r]
		rollDown := out.RollDown[r]
		perWin := out.PaidPerWin[r]
		total := out.PaidTotal[r]
		carry := out.CarryOut[r]

		fmt.Printf("%4s | %10d | %18d | %7d | %8d | %10d | %9d | %7d\n",
			label,
			poolBefore,
			poolAfter,
			rollDown,
			winners,
			perWin,
			total,
			carry,
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
