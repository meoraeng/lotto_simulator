package lotto

import (
	"math/rand"
	"sort"
)

// 기능 추가 : 이월, 상한,
type Lottos struct {
	Lottos         []Lotto
	BonusNumber    int
	WinningNumbers []int
}

type Lotto struct {
	Numbers []int
}

const (
	LottoSize   = 6
	LottoMinNum = 1
	LottoMaxNum = 45
	LottoPrice  = 1000
)

func PurchaseLottos(amount int) (Lottos, error) {
	if err := validatePurchaseAmount(amount); err != nil {
		return Lottos{}, err
	}

	count := amount / LottoPrice
	lottos := make([]Lotto, 0, count)

	for i := 0; i < count; i++ {
		numbers := generateRandomNumbers()
		lottos = append(lottos, Lotto{Numbers: numbers})
	}

	return Lottos{
		Lottos: lottos,
	}, nil
}

func (l *Lottos) SetWinningNumbers(input string) error {
	parsed, err := parseWinningNumbers(input)
	if err != nil {
		return err
	}
	l.WinningNumbers = parsed
	return nil
}

func (l *Lottos) SetBonusNumber(input string) error {
	parsed, err := parseBonusNumber(input, l.WinningNumbers)
	if err != nil {
		return err
	}
	l.BonusNumber = parsed
	return nil
}

func (lt Lotto) matchCount(winning []int) int {
	count := 0
	for _, n := range lt.Numbers {
		if contains(winning, n) {
			count++
		}
	}
	return count
}

func (lt Lotto) hasBonus(bonus int) bool {
	return contains(lt.Numbers, bonus)
}

func (ls Lottos) CompileStatistics() map[Rank]int {
	stats := make(map[Rank]int)

	for _, lotto := range ls.Lottos {
		match := lotto.matchCount(ls.WinningNumbers)
		hasBonus := lotto.hasBonus(ls.BonusNumber)
		rank := DetermineRank(match, hasBonus)
		stats[rank]++
	}
	return stats
}

func CalculateTotalPrize(stats map[Rank]int) int {
	total := 0
	for rank, count := range stats {
		total += rank.Prize() * count
	}
	return total
}

func CalculateProfitRate(totalPrize, purchaseAmount int) float64 {
	if purchaseAmount == 0 {
		return 0
	}
	return float64(totalPrize) / float64(purchaseAmount) * 100
}

func generateRandomNumbers() []int {
	numbers := make([]int, 0, LottoSize)

	for len(numbers) < LottoSize {
		num := rand.Intn(LottoMaxNum) + 1
		if !contains(numbers, num) {
			numbers = append(numbers, num)
		}
	}
	sort.Ints(numbers)

	return numbers
}
