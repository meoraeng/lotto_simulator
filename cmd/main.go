package main

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"
)

type UserInputError struct {
	msg string
}

func (e UserInputError) Error() string {
	return "[ERROR] " + e.msg
}

func NewUserInputError(msg string) error {
	return UserInputError{msg: msg}
}

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

func (l *Lottos) SetWinningNumbers(input string) error {
	parsed, err := parseWinningNumbers(input)
	if err != nil {
		return err
	}
	l.winningNumbers = parsed
	return nil
}

func (l *Lottos) SetBonusNumber(input string) error {
	parsed, err := parseBonusNumber(input, l.winningNumbers)
	if err != nil {
		return err
	}
	l.bonusNumber = parsed
	return nil
}

func purchaseLottos(amount int) (Lottos, error) {
	if err := validatePurchaseAmount(amount); err != nil {
		return Lottos{}, err
	}

	count := amount / LottoPrice
	lottos := make([]Lotto, 0, count)

	for i := 0; i < count; i++ {
		numbers := generateRandomNumbers()
		lottos = append(lottos, Lotto{numbers})
	}

	return Lottos{
		lottos: lottos,
	}, nil
}

func validatePurchaseAmount(amount int) error {
	if amount <= 0 {
		return NewUserInputError("구매 금액은 양수여야 합니다.")
	}
	if amount%LottoPrice != 0 {
		return NewUserInputError(fmt.Sprintf("구매 금액은 %d원 단위여야 합니다.", LottoPrice))
	}
	return nil
}

func parseWinningNumbers(input string) ([]int, error) {
	tokens := splitAndClean(input)

	if len(tokens) != LottoSize {
		return nil, NewUserInputError(
			fmt.Sprintf("당첨 번호는 %d개여야 합니다. 입력 개수: %d", LottoSize, len(tokens)),
		)
	}

	nums := make([]int, 0, LottoSize)

	for _, t := range tokens {
		n, err := parseInt(t)
		if err != nil {
			return nil, err
		}
		if err := validateRange(n); err != nil {
			return nil, err
		}
		nums = append(nums, n)
	}

	if err := validateNoDuplicates(nums); err != nil {
		return nil, err
	}

	sort.Ints(nums)
	return nums, nil
}

func parseBonusNumber(input string, winning []int) (int, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return 0, NewUserInputError("보너스 번호를 입력해야 합니다.")
	}

	n, err := parseInt(input)
	if err != nil {
		return 0, err
	}
	if err := validateRange(n); err != nil {
		return 0, err
	}
	if contains(winning, n) {
		return 0, NewUserInputError(
			fmt.Sprintf("보너스 번호는 당첨 번호와 중복될 수 없습니다: %d", n),
		)
	}

	return n, nil
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

func parseInt(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, NewUserInputError(fmt.Sprintf("숫자가 아닌 값이 포함됨: %q", s))
	}
	return n, nil
}

func validateRange(n int) error {
	if n < LottoMinNum || n > LottoMaxNum {
		return NewUserInputError(
			fmt.Sprintf("번호는 %d~%d 사이여야 합니다: %d", LottoMinNum, LottoMaxNum, n),
		)
	}
	return nil
}

func validateNoDuplicates(nums []int) error {
	seen := make(map[int]bool)
	for _, n := range nums {
		if seen[n] {
			return NewUserInputError(
				fmt.Sprintf("중복된 번호가 있습니다: %d", n),
			)
		}
		seen[n] = true
	}
	return nil
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
	lottos, err := purchaseLottos(5000)
	if err != nil {
		fmt.Println(err)
		return
	}

	if err := lottos.SetWinningNumbers("1, 2, 3, 4, 5, 6"); err != nil {
		fmt.Println(err)
		return
	}

	if err := lottos.SetBonusNumber("7"); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("당첨 번호:", lottos.winningNumbers)
	fmt.Println("보너스 번호:", lottos.bonusNumber)
}
