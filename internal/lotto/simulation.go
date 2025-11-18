package lotto

// 플레이어들 티켓과 당첨 번호 기반으로 Winners 계산
func CountWinnersFromPlayers(players []Player, winning Lottos) map[Rank]int {
	stats := make(map[Rank]int)

	for _, p := range players {
		for _, lotto := range p.Tickets {
			match := lotto.matchCount(winning.WinningNumbers)
			hasBonus := lotto.hasBonus(winning.BonusNumber)
			rank := DetermineRank(match, hasBonus)

			stats[rank]++
		}
	}

	return stats
}

// RoundInput으로 Winners 설정
func BuildRoundInput(base RoundInput, players []Player, winning Lottos) RoundInput {
	base.Winners = CountWinnersFromPlayers(players, winning)
	return base
}

// 플레이어에게 지급액 분배
func DistributeRewards(players []Player, winning Lottos, out RoundOutput) map[string]int {
	rewards := make(map[string]int)

	for _, p := range players {
		for _, lotto := range p.Tickets {
			match := lotto.matchCount(winning.WinningNumbers)
			hasBonus := lotto.hasBonus(winning.BonusNumber)
			rank := DetermineRank(match, hasBonus)

			perWin := out.PaidPerWin[rank]
			if perWin > 0 {
				rewards[p.Name] += perWin
			}
		}
	}

	return rewards
}
