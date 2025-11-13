package main

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
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

// 당첨 번호 설정
func (l *Lottos) SetWinningNumbers(input string) {
	parsed := parseWinningNumbers(input)
	l.winningNumbers = parsed
}

// 보너스 번호 설정
func (l *Lottos) SetBonusNumber(input string) {
	parsed := parseBonusNumber(input, l.winningNumbers)
	l.bonusNumber = parsed
}

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

func parseWinningNumbers(input string) []int {
	tokens := splitAndClean(input)

	if len(tokens) != LottoSize {
		panic(fmt.Sprintf("당첨 번호는 %d개여야 합니다. 입력 개수: %d",
			LottoSize, len(tokens)))
	}

	nums := make([]int, 0, LottoSize)

	for _, t := range tokens {
		n := parseInt(t)
		validateRange(n)
		nums = append(nums, n)
	}

	validateNoDuplicates(nums)

	sort.Ints(nums)
	return nums
}

func parseBonusNumber(input string, winning []int) int {
	input = strings.TrimSpace(input)
	if input == "" {
		panic("보너스 번호를 입력해야 합니다.")
	}

	n := parseInt(input)
	validateRange(n)

	if contains(winning, n) {
		panic(fmt.Sprintf("보너스 번호는 당첨 번호와 중복될 수 없습니다: %d", n))
	}

	return n
}

func splitAndClean(input string) []string {
	parts := strings.Split(input, ",")
	clean := make([]string, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			clean = append(clean, p)
		}
	}
	return clean
}

func parseInt(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		panic(fmt.Sprintf("숫자가 아닌 값이 포함됨: %q", s))
	}
	return n
}

func validateRange(n int) {
	if n < LottoMinNum || n > LottoMaxNum {
		panic(fmt.Sprintf("번호는 %d~%d 사이여야 합니다: %d",
			LottoMinNum, LottoMaxNum, n))
	}
}

func validateNoDuplicates(nums []int) {
	seen := make(map[int]bool)
	for _, n := range nums {
		if seen[n] {
			panic(fmt.Sprintf("중복된 번호가 있습니다: %d", n))
		}
		seen[n] = true
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
	lottos := purchaseLottos(5000)

	lottos.SetWinningNumbers("1, 2, 3, 4, 5, 6")

	lottos.SetBonusNumber("7")

	fmt.Println("당첨 번호:", lottos.winningNumbers)
	fmt.Println("보너스 번호:", lottos.bonusNumber)
}
