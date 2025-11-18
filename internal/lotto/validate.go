package lotto

import (
	"errors"
	"fmt"
)

func validatePurchaseAmount(amount int) error {
	if amount <= 0 {
		return errors.New("구매 금액은 양수여야 합니다")
	}
	if amount%LottoPrice != 0 {
		return fmt.Errorf("구매 금액은 %d원 단위여야 합니다", LottoPrice)
	}
	return nil
}

func validateRange(n int) error {
	if n < LottoMinNum || n > LottoMaxNum {
		return fmt.Errorf(
			"번호는 %d~%d 사이여야 합니다: %d", LottoMinNum, LottoMaxNum, n,
		)
	}
	return nil
}

func validateNoDuplicates(nums []int) error {
	seen := make(map[int]bool)
	for _, n := range nums {
		if seen[n] {
			return fmt.Errorf("중복된 번호가 있습니다: %d", n)
		}
		seen[n] = true
	}
	return nil
}

func validateWinningFormat(input string) error {
	for _, r := range input {
		switch {
		case r >= '0' && r <= '9':
		case r == ',':
		case r == ' ' || r == '\t':
		default:
			return fmt.Errorf(
				"잘못된 문자 포함: %q (숫자와 콤마(,)만 허용됩니다)", r,
			)
		}
	}
	return nil
}

// 외부 wrapper
func ValidatePurchaseAmount(amount int) error {
	return validatePurchaseAmount(amount)
}

func contains(slice []int, target int) bool {
	for _, v := range slice {
		if v == target {
			return true
		}
	}
	return false
}
