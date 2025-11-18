package lotto

import "errors"

var (
	ErrInvalidMode       = errors.New("유효하지 않은 모드입니다")
	ErrInvalidAllocation = errors.New("배당 비율 합이 100%가 아닙니다")
	ErrNegativeSales     = errors.New("판매액은 음수가 될 수 없습니다")
	ErrInvalidRank       = errors.New("유효하지 않은 등수(rank) 값입니다")
)
