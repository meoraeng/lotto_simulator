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
)

// func purchaseLottos(amount int) {

// }

// func generateBonusNumber() {

// }

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
	fmt.Println(generateRandomNumbers())
}
