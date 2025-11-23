package lotto

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func parseWinningNumbers(input string) ([]int, error) {
	if err := validateWinningFormat(input); err != nil {
		return nil, err
	}

	tokens := splitAndClean(input)

	if len(tokens) != LottoSize {
		return nil, fmt.Errorf(
			"당첨 번호는 %d개여야 합니다. 입력 개수: %d", LottoSize, len(tokens),
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
		return 0, fmt.Errorf("보너스 번호를 입력해야 합니다")
	}

	n, err := parseInt(input)
	if err != nil {
		return 0, err
	}
	if err := validateRange(n); err != nil {
		return 0, err
	}
	if contains(winning, n) {
		return 0, fmt.Errorf(
			"보너스 번호는 당첨 번호와 중복될 수 없습니다: %d", n,
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
		return 0, fmt.Errorf("숫자가 아닌 값이 포함됨: %q", s)
	}
	return n, nil
}

