package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/meoraeng/lotto_simulator/internal/lotto"
)

func FormatRoundReport(in lotto.RoundInput, out lotto.RoundOutput) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("총 판매액: %s원\n", Comma(out.Sales)))
	b.WriteString(fmt.Sprintf("라운드 잔액: %s원\n\n", Comma(out.RoundRemainder)))

	// 헤더
	b.WriteString(
		fmt.Sprintf(
			"%4s | %12s | %12s | %10s | %10s | %12s | %12s | %10s\n",
			"Rank", "Pool(Before)", "Pool(After)", "Rolldown",
			"Winners", "PerWin", "Total", "Carry",
		),
	)
	b.WriteString(strings.Repeat("-", 110) + "\n")

	order := []lotto.Rank{lotto.Rank1, lotto.Rank2, lotto.Rank3, lotto.Rank4, lotto.Rank5}

	for i, r := range order {
		rankNo := i + 1
		b.WriteString(fmt.Sprintf(
			"%4d | %12s | %12s | %10s | %10d | %12s | %12s | %10s\n",
			rankNo,
			Comma(out.PoolBefore[r]),
			Comma(out.PoolAfterCap[r]),
			Comma(out.RollDown[r]),
			in.Winners[r],
			Comma(out.PaidPerWin[r]),
			Comma(out.PaidTotal[r]),
			Comma(out.CarryOut[r]),
		))
	}

	return b.String()
}

func Comma(n int) string {
	s := strconv.Itoa(n)
	neg := false
	if n < 0 {
		neg = true
		s = s[1:]
	}

	var out strings.Builder
	for i, c := range s {
		if i != 0 && (len(s)-i)%3 == 0 {
			out.WriteByte(',')
		}
		out.WriteRune(c)
	}

	if neg {
		return "-" + out.String()
	}
	return out.String()
}
