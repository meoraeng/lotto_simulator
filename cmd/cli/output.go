package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/meoraeng/lotto_simulator/internal/lotto"
	"github.com/meoraeng/lotto_simulator/internal/lotto/ui"
)

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
	switch in.Mode {
	case lotto.ModeParimutuel:
		fmt.Println(ui.FormatRoundReport(in, out))
	case lotto.ModeFixedPayout:
		printFixedPayoutReport(in, out)
	default:
		printError(errors.New("지원하지 않는 모드입니다"))
	}
}

func printFixedPayoutReport(in lotto.RoundInput, out lotto.RoundOutput) {
	// 헤더(등수, 당첨자 수, 인당 지급액, 총 지급액)
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
