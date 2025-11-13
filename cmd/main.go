package main

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

// 기능 추가 : 이월, 상한,
type Lottos struct {
	lottos         []Lotto
	bonusNumber    int
	winningNumbers []int
}

type Lotto struct {
	numbers []int
}

const (
	LottoSize   = 6
	LottoMinNum = 1
	LottoMaxNum = 45
	LottoPrice  = 1000
)

func purchaseLottos(amount int) Lottos {
	validatePurchaseAmount(amount)

	count := amount / LottoPrice
	lottos := make([]Lotto, 0, count)

	for i := 0; i < count; i++ {
		numbers := generateRandomNumbers()
		lottos = append(lottos, Lotto{numbers})
	}

	return Lottos{
		lottos: lottos,
	}
}

func validatePurchaseAmount(amount int) {
	if amount <= 0 {
		panic("구매 금액은 양수여야 합니다.")
	}
	if amount%LottoPrice != 0 {
		panic(fmt.Sprintf("구매 금액은 %d원 단위여야 합니다.", LottoPrice))
	}
}

func generateRandomNumbers() []int {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	numbers := make([]int, 0, LottoSize)

	for len(numbers) < LottoSize {
		num := r.Intn(LottoMaxNum) + 1
		if !contains(numbers, num) {
			numbers = append(numbers, num)
		}
	}
	return sortLottoNumbers(numbers)
}

func contains(slice []int, target int) bool {
	for _, v := range slice {
		if v == target {
			return true
		}
	}
	return false
}

func sortLottoNumbers(n []int) []int {
	sort.Ints(n)
	return n
}

func main() {
	result := purchaseLottos(5000)
	fmt.Println(result)
}
