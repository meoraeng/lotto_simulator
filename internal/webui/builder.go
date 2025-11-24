package webui

import (
	"fmt"

	"github.com/meoraeng/lotto_simulator/internal/lotto"
)

func makeIndexList(n int) []int {
	out := make([]int, n)
	for i := 0; i < n; i++ {
		out[i] = i + 1
	}
	return out
}

func flattenTickets(players []playerTicketsView) []lotto.Lotto {
	var all []lotto.Lotto
	for _, p := range players {
		all = append(all, p.Tickets...)
	}
	return all
}

func buildPurchasePageData(
	mode lotto.Mode,
	count int,
	roundCount int,
	players []playerTicketsView,
	totalSales int,
	errorMsg string,
) purchasePageData {
	data := purchasePageData{
		Mode:       mode,
		Count:      count,
		RoundCount: roundCount,
		IndexList:  makeIndexList(count),
		LottoPrice: lotto.LottoPrice,
		Players:    players,
		TotalSales: totalSales,
	}

	if errorMsg != "" {
		data.Error = errorText(errorMsg)
	}

	return data
}

func buildPlayerSummaries(players []playerTicketsView, payouts map[string]int) []playerSummary {
	summaries := make([]playerSummary, 0, len(players))

	for _, p := range players {
		spent := p.Amount
		earned := payouts[p.Name]

		rate := 0.0
		if spent > 0 {
			rate = float64(earned) / float64(spent) * 100
		}

		summaries = append(summaries, playerSummary{
			Name:        p.Name,
			Amount:      spent,
			TicketCount: len(p.Tickets),
			Earned:      earned,
			ProfitRate:  rate,
		})
	}

	return summaries
}

func buildResultData(
	req resultRequest,
	l *lotto.Lottos,
	domainPlayers []lotto.Player,
) map[string]any {
	stats := l.CompileStatisticsParallel()
	roundIn := buildRoundInputForMode(req.Mode, req.TotalSales, stats)

	roundOut, _ := lotto.CalculateRound(roundIn)
	payouts := lotto.DistributeRewardsParallel(domainPlayers, *l, roundOut)
	rankRows := buildRankRows(req.Mode, stats, roundOut)
	playerSummaries := buildPlayerSummaries(req.Players, payouts)

	data := map[string]any{
		"Mode":            req.Mode,
		"TotalSales":      req.TotalSales,
		"WinningNumbers":  l.WinningNumbers,
		"BonusNumber":     l.BonusNumber,
		"RankRows":        rankRows,
		"PlayerSummaries": playerSummaries,
	}

	if req.Mode == lotto.ModeParimutuel {
		data["RoundOutput"] = roundOut
		data["DetailRows"] = buildDetailRows(roundIn, roundOut)
	}

	return data
}

func buildRoundInputForMode(
	mode lotto.Mode,
	totalSales int,
	stats map[lotto.Rank]int,
) lotto.RoundInput {
	if mode == lotto.ModeFixedPayout {
		// 고정 상금 모드
		fixedPayout := map[lotto.Rank]int{
			lotto.Rank1: lotto.Rank1.Prize(),
			lotto.Rank2: lotto.Rank2.Prize(),
			lotto.Rank3: lotto.Rank3.Prize(),
			lotto.Rank4: lotto.Rank4.Prize(),
			lotto.Rank5: lotto.Rank5.Prize(),
		}
		return lotto.RoundInput{
			Mode:        mode,
			Sales:       totalSales,
			Winners:     stats,
			FixedPayout: fixedPayout,
		}
	}

	// 분배 모드
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

	return lotto.RoundInput{
		Mode:           mode,
		Sales:          totalSales,
		Winners:        stats,
		CarryIn:        map[lotto.Rank]int{},
		Allocations:    allocations,
		CapPerRank:     caps,
		RoundingUnit:   100,
		RollDownMethod: lotto.RollDownProportional,
	}
}

func buildRoundInputForModeWithCarry(
	mode lotto.Mode,
	totalSales int,
	stats map[lotto.Rank]int,
	carryIn map[lotto.Rank]int,
) lotto.RoundInput {
	if mode == lotto.ModeFixedPayout {
		fixedPayout := map[lotto.Rank]int{
			lotto.Rank1: lotto.Rank1.Prize(),
			lotto.Rank2: lotto.Rank2.Prize(),
			lotto.Rank3: lotto.Rank3.Prize(),
			lotto.Rank4: lotto.Rank4.Prize(),
			lotto.Rank5: lotto.Rank5.Prize(),
		}
		return lotto.RoundInput{
			Mode:        mode,
			Sales:       totalSales,
			Winners:     stats,
			FixedPayout: fixedPayout,
		}
	}

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

	return lotto.RoundInput{
		Mode:           mode,
		Sales:          totalSales,
		Winners:        stats,
		CarryIn:        carryIn,
		Allocations:    allocations,
		CapPerRank:     caps,
		RoundingUnit:   100,
		RollDownMethod: lotto.RollDownProportional,
	}
}

func buildRankRows(
	mode lotto.Mode,
	stats map[lotto.Rank]int,
	roundOut lotto.RoundOutput,
) []rankRowView {
	ranks := []struct {
		rank      lotto.Rank
		label     string
		condition string
	}{
		{lotto.Rank1, "1등", "6 match"},
		{lotto.Rank2, "2등", "5 + bonus"},
		{lotto.Rank3, "3등", "5 match"},
		{lotto.Rank4, "4등", "4 match"},
		{lotto.Rank5, "5등", "3 match"},
	}

	rows := make([]rankRowView, 0, len(ranks))
	for _, r := range ranks {
		prize := calculatePrizeForMode(mode, r.rank, roundOut)

		rows = append(rows, rankRowView{
			RankLabel: r.label,
			Condition: r.condition,
			Prize:     prize,
			Count:     stats[r.rank],
		})
	}
	return rows
}

func calculatePrizeForMode(mode lotto.Mode, rank lotto.Rank, roundOut lotto.RoundOutput) int {
	if mode == lotto.ModeFixedPayout {
		return rank.Prize()
	}
	return roundOut.PaidPerWin[rank]
}

func buildDetailRows(
	roundIn lotto.RoundInput,
	roundOut lotto.RoundOutput,
) []detailRowView {
	ranks := []struct {
		rank  lotto.Rank
		label string
	}{
		{lotto.Rank1, "1등"},
		{lotto.Rank2, "2등"},
		{lotto.Rank3, "3등"},
		{lotto.Rank4, "4등"},
		{lotto.Rank5, "5등"},
	}

	rows := make([]detailRowView, 0, len(ranks))
	for _, r := range ranks {
		rows = append(rows, detailRowView{
			RankLabel:  r.label,
			PoolBefore: roundOut.PoolBefore[r.rank],
			PoolAfter:  roundOut.PoolAfterCap[r.rank],
			RollDown:   roundOut.RollDown[r.rank],
			Winners:    roundIn.Winners[r.rank],
			PerWin:     roundOut.PaidPerWin[r.rank],
			Total:      roundOut.PaidTotal[r.rank],
			Carry:      roundOut.CarryOut[r.rank],
		})
	}
	return rows
}

func buildDetailRowsIfNeeded(
	mode lotto.Mode,
	roundIn lotto.RoundInput,
	roundOut lotto.RoundOutput,
) []detailRowView {
	if mode == lotto.ModeParimutuel {
		return buildDetailRows(roundIn, roundOut)
	}
	return nil
}

func mergePayouts(dst, src map[string]int) {
	for name, amount := range src {
		dst[name] += amount
	}
}

func buildPlayerRedirectURL(mode lotto.Mode, count int, roundCount int) string {
	return fmt.Sprintf("/purchase?mode=%d&count=%d&rounds=%d", mode, count, roundCount)
}
