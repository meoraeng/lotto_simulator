package lotto

type Mode int

const (
	ModeFixedPayout Mode = iota // 기존의 고정된 상금 모드
	ModeParimutuel              // 판매액 기반 분배 모드
)

// 퍼센트 단위
const BasisPoints = 10_000

// 각 등수에 판매액의 몇 bp를 배정할 것인지
type Allocation struct {
	Rank        Rank
	BasisPoints int
}

// 한 회차당 필요한 입력값
type RoundInput struct {
	Mode Mode

	Sales   int          // 회차 판매액
	Winners map[Rank]int // 등수별 당첨자 수
	CarryIn map[Rank]int // 등수별 이월 금액(없으면 0)

	Allocations []Allocation // 등수별 배정 비율
	CapPerRank  map[Rank]int // 등수별 상한 금액
}

// 한 회차 분배 결과
type RoundOutput struct {
	Sales int

	PoolBefore   map[Rank]int // 이월 포함, 상한/롤다운 적용 전 풀 금액
	PoolAfterCap map[Rank]int // 상한 적용 후 풀 금액
	PaidPerWin   map[Rank]int // 등수별 1인당 지급액
	PaidTotal    map[Rank]int // 등수별 총 지급액
	CarryOut     map[Rank]int // 등수별 다음 회차로 이월되는 금액
	RollDown     map[Rank]int // 상한 초과로 하위 등수로 내려보낸 금액
}

func CalculateRound(in RoundInput) RoundOutput {
	out := newRoundOutput(in)

	if in.Mode != ModeParimutuel {
		return out
	}

	for _, a := range in.Allocations {
		calcRankPool(in, &out, a)
	}

	return out
}

func newRoundOutput(in RoundInput) RoundOutput {
	return RoundOutput{
		Sales:        in.Sales,
		PoolBefore:   make(map[Rank]int),
		PoolAfterCap: make(map[Rank]int),
		PaidPerWin:   make(map[Rank]int),
		PaidTotal:    make(map[Rank]int),
		CarryOut:     make(map[Rank]int),
		RollDown:     make(map[Rank]int),
	}
}

func calcRankPool(in RoundInput, out *RoundOutput, a Allocation) {
	basePool := in.Sales * a.BasisPoints / BasisPoints
	carry := in.CarryIn[a.Rank]
	pool := basePool + carry

	out.PoolBefore[a.Rank] = pool
	out.PoolAfterCap[a.Rank] = pool

	winners := in.Winners[a.Rank]
	if winners <= 0 {
		// 당첨자가 없으면 전액 이월
		out.CarryOut[a.Rank] = pool
		return
	}
	// 당첨자 있으면 균등 분배
	perWinner := pool / winners
	out.PaidPerWin[a.Rank] = perWinner
	out.PaidTotal[a.Rank] = perWinner * winners
}
