package lotto

import "fmt"

// 여러 회차를 공통 규칙으로 돌리기 위한 설정
type SeriesConfig struct {
	Mode        Mode
	Allocations []Allocation
	CapPerRank  map[Rank]int
}

// 여러 회차에 대해 라운드 로직 순차 실행 -> 각 회차 결과 반환
// 이월 금액은 다음 회차 입력으로 전달
func SimulateSeries(
	cfg SeriesConfig,
	salesPerRound []int,
	winnersPerRound []map[Rank]int,
	carryIn map[Rank]int,
) ([]RoundOutput, error) {

	if len(salesPerRound) != len(winnersPerRound) {
		return nil, fmt.Errorf(
			"회차별 판매액 개수와 당첨자 정보 개수가 일치하지 않습니다",
		)
	}

	results := make([]RoundOutput, 0, len(salesPerRound))
	carry := cloneRankIntMap(carryIn)

	for i := 0; i < len(salesPerRound); i++ {
		input := RoundInput{
			Mode:        cfg.Mode,
			Sales:       salesPerRound[i],
			Winners:     winnersPerRound[i],
			CarryIn:     cloneRankIntMap(carry),
			Allocations: cfg.Allocations,
			CapPerRank:  cfg.CapPerRank,
		}

		out, err := CalculateRound(input)
		if err != nil { // 실패하면 리턴
			return nil, err
		}

		results = append(results, out)

		// 다음 회차 이월 금액 갱신
		carry = cloneRankIntMap(out.CarryOut)
	}

	return results, nil
}

// 맵 복사 -> 참조 공유 방지
func cloneRankIntMap(src map[Rank]int) map[Rank]int {
	dst := make(map[Rank]int, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
