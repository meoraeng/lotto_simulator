package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/meoraeng/lotto_simulator/internal/formatter"
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

type reportPrinter func(lotto.RoundInput, lotto.RoundOutput)

var modeReportPrinters = map[lotto.Mode]reportPrinter{
	lotto.ModeParimutuel:  func(in lotto.RoundInput, out lotto.RoundOutput) { fmt.Println(ui.FormatRoundReport(in, out)) },
	lotto.ModeFixedPayout: printFixedPayoutReport,
}

func printRoundReport(in lotto.RoundInput, out lotto.RoundOutput) {
	printer, exists := modeReportPrinters[in.Mode]
	if !exists {
		printError(errors.New("지원하지 않는 모드입니다"))
		return
	}
	printer(in, out)
}

var rankConditionLabels = map[lotto.Rank]string{
	lotto.Rank1: "6 match",
	lotto.Rank2: "5 + bonus",
	lotto.Rank3: "5 match",
	lotto.Rank4: "4 match",
	lotto.Rank5: "3 match",
}

func conditionLabel(r lotto.Rank) string {
	label, exists := rankConditionLabels[r]
	if !exists {
		return "-"
	}
	return label
}

func printFixedPayoutReport(in lotto.RoundInput, out lotto.RoundOutput) {
	// 헤더(등수, 당첨자 수, 인당 지급액, 총 지급액)
	fmt.Printf(
		"%4s | %-12s | %15s | %8s | %15s | %15s\n",
		"Rank", "Cond", "Prize", "WinCnt", "PerWin", "Total",
	)
	fmt.Println(strings.Repeat("-", 90))

	order := []lotto.Rank{
		lotto.Rank1,
		lotto.Rank2,
		lotto.Rank3,
		lotto.Rank4,
		lotto.Rank5,
	}

	for i, r := range order {
		rankNo := i + 1

		cond := conditionLabel(r)
		basePrize := formatter.Money(r.Prize())      // 기준 상금
		perWin := formatter.Money(out.PaidPerWin[r]) // 1인당 지급액
		total := formatter.Money(out.PaidTotal[r])   // 총 지급액
		winCnt := in.Winners[r]

		fmt.Printf(
			"%4d | %-12s | %15s | %8d | %15s | %15s\n",
			rankNo,
			cond,
			basePrize,
			winCnt,
			perWin,
			total,
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
