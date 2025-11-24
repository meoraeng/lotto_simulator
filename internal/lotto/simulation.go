package lotto

// 플레이어들 티켓과 당첨 번호 기반으로 Winners 계산
func CountWinnersFromPlayers(players []Player, winning Lottos) map[Rank]int {
	stats := make(map[Rank]int)

	for _, p := range players {
		playerStats := countWinnersFromPlayerTickets(p.Tickets, winning)
		mergeStats(stats, playerStats)
	}

	return stats
}

func countWinnersFromPlayerTickets(tickets []Lotto, winning Lottos) map[Rank]int {
	stats := make(map[Rank]int)
	for _, lotto := range tickets {
		rank := determineTicketRank(lotto, winning)
		stats[rank]++
	}
	return stats
}

func determineTicketRank(ticket Lotto, winning Lottos) Rank {
	match := ticket.matchCount(winning.WinningNumbers)
	hasBonus := ticket.hasBonus(winning.BonusNumber)
	return DetermineRank(match, hasBonus)
}

func mergeStats(dst, src map[Rank]int) {
	for rank, count := range src {
		dst[rank] += count
	}
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
		playerReward := calculatePlayerReward(p.Tickets, p.Name, winning, out)
		rewards[p.Name] += playerReward
	}

	return rewards
}

func calculatePlayerReward(tickets []Lotto, playerName string, winning Lottos, out RoundOutput) int {
	total := 0
	for _, ticket := range tickets {
		rank := determineTicketRank(ticket, winning)
		perWin := out.PaidPerWin[rank]
		if perWin > 0 {
			total += perWin
		}
	}
	return total
}
