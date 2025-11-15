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
