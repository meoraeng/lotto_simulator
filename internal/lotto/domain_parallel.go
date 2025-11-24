package lotto

import (
	"sync"
)

// 대량의 티켓 처리 시 goroutine을 활용해 성능 향상
func (ls Lottos) CompileStatisticsParallel() map[Rank]int {
	if len(ls.Lottos) < 100 {
		// 티켓이 적으면 오버헤드가 더 클 수 있으므로 순차 처리
		return ls.CompileStatistics()
	}

	const numWorkers = 4
	ticketsPerWorker := len(ls.Lottos) / numWorkers
	if ticketsPerWorker == 0 {
		ticketsPerWorker = 1
	}

	var wg sync.WaitGroup
	statsChan := make(chan map[Rank]int, numWorkers)

	// Worker goroutine들 생성
	for i := 0; i < numWorkers; i++ {
		start := i * ticketsPerWorker
		end := start + ticketsPerWorker
		if i == numWorkers-1 {
			end = len(ls.Lottos) // 마지막 worker는 나머지 모두 처리
		}

		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()

			localStats := make(map[Rank]int)
			for j := start; j < end; j++ {
				lotto := ls.Lottos[j]
				match := lotto.matchCount(ls.WinningNumbers)
				hasBonus := lotto.hasBonus(ls.BonusNumber)
				rank := DetermineRank(match, hasBonus)
				localStats[rank]++
			}
			statsChan <- localStats
		}(start, end)
	}

	// 모든 worker 완료 대기
	go func() {
		wg.Wait()
		close(statsChan)
	}()

	// 결과 합산
	stats := make(map[Rank]int)
	for localStats := range statsChan {
		for rank, count := range localStats {
			stats[rank] += count
		}
	}

	return stats
}

// 대량의 티켓 처리 시 goroutine을 활용해 성능 향상
func DistributeRewardsParallel(
	players []Player,
	winning Lottos,
	out RoundOutput,
) map[string]int {
	// 전체 티켓 수 계산
	totalTickets := 0
	for _, p := range players {
		totalTickets += len(p.Tickets)
	}

	if totalTickets < 100 {
		// 티켓이 적으면 오버헤드가 더 클 수 있으므로 순차 처리
		return DistributeRewards(players, winning, out)
	}

	const numWorkers = 4
	rewardsChan := make(chan map[string]int, numWorkers)
	var wg sync.WaitGroup

	playersPerWorker := len(players) / numWorkers
	if playersPerWorker == 0 {
		playersPerWorker = 1
	}

	// Worker goroutine들 생성
	for i := 0; i < numWorkers; i++ {
		start := i * playersPerWorker
		end := start + playersPerWorker
		if i == numWorkers-1 {
			end = len(players) // 마지막 worker는 나머지 모두 처리
		}

		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()

			localRewards := calculateRewardsForPlayersRange(
				players[start:end],
				winning,
				out,
			)
			rewardsChan <- localRewards
		}(start, end)
	}

	// 모든 worker 완료 대기
	go func() {
		wg.Wait()
		close(rewardsChan)
	}()

	// 결과 합산
	rewards := make(map[string]int)
	for localRewards := range rewardsChan {
		for name, amount := range localRewards {
			rewards[name] += amount
		}
	}

	return rewards
}

func calculateRewardsForPlayersRange(
	players []Player,
	winning Lottos,
	out RoundOutput,
) map[string]int {
	localRewards := make(map[string]int)
	for _, p := range players {
		playerReward := calculatePlayerReward(p.Tickets, p.Name, winning, out)
		localRewards[p.Name] += playerReward
	}
	return localRewards
}
