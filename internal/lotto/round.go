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
	// 분배 모드
	Allocations []Allocation // 등수별 배정 비율
	CapPerRank  map[Rank]int // 등수별 상한 금액
	// 고정 모드
	FixedPayout map[Rank]int
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

	RoundRemainder int // 판매액 중 풀에 배정되지 않은 라운드 잔액
}

func CalculateRound(in RoundInput) RoundOutput {
	out := newRoundOutput(in)

	switch in.Mode {
	case ModeParimutuel:
		calcParimutuelRound(&out, in)
	case ModeFixedPayout:
		calcFixedPayoutRound(&out, in)
	default:
	}

	return out
}

func calcParimutuelRound(out *RoundOutput, in RoundInput) {
	order := []Rank{Rank1, Rank2, Rank3, Rank4, Rank5}
	allocBps := buildAllocationMap(in.Allocations)

	allocatedFromSales := calcBasePools(out, in, order, allocBps)
	// 잔액 기록
	out.RoundRemainder = in.Sales - allocatedFromSales
	if out.RoundRemainder < 0 {
		// 비정상 설정 방어
		out.RoundRemainder = 0
	}

	applyCapAndRolldown(out.PoolAfterCap, in.CapPerRank, allocBps, order, out.RollDown)
	calcPayoutAndCarry(out, in, order)
}

func calcBasePools(
	out *RoundOutput,
	in RoundInput,
	order []Rank,
	allocBps map[Rank]int,
) int {
	allocatedFromSales := 0

	for _, r := range order {
		bps := allocBps[r]
		basePool := in.Sales * bps / BasisPoints
		allocatedFromSales += basePool

		carry := in.CarryIn[r]
		pool := basePool + carry

		out.PoolBefore[r] = pool
		out.PoolAfterCap[r] = pool
	}

	return allocatedFromSales
}

// PoolAfterCap 기준 지급 및 이월 계산
func calcPayoutAndCarry(
	out *RoundOutput,
	in RoundInput,
	order []Rank,
) {
	for _, r := range order {
		pool := out.PoolAfterCap[r]
		winners := in.Winners[r]

		if winners <= 0 {
			// 당첨자 없으면 전체 이월
			out.CarryOut[r] = pool
			continue
		}

		per := pool / winners
		total := per * winners

		out.PaidPerWin[r] = per
		out.PaidTotal[r] = total

		// 잔액 이월
		remain := pool - total
		if remain > 0 {
			out.CarryOut[r] = remain
		}
	}
}

func applyCapAndRolldown(
	pool map[Rank]int,
	caps map[Rank]int,
	allocBps map[Rank]int,
	order []Rank,
	rollDown map[Rank]int,
) {
	for i, r := range order {
		cap, hasCap := caps[r]
		if !hasCap {
			continue
		}
		if pool[r] <= cap {
			continue
		}

		overflow := pool[r] - cap
		pool[r] = cap
		rollDown[r] += overflow

		lowerRanks := order[i+1:]
		if len(lowerRanks) == 0 {
			// 하위 등수가 없으면, 여기서는 그냥 초과분을 버리는 정책으로 둔다.
			// (필요하면 나중에 "특별 기금" 개념을 추가해도 됨)
			continue
		}

		// 하위 등수 전체 비율 합
		totalBasicPoints := 0
		for _, lr := range lowerRanks {
			totalBasicPoints += allocBps[lr]
		}
		if totalBasicPoints == 0 {
			// 비율 정보가 없으면 가장 마지막 하위 등수에 몰아준다
			last := lowerRanks[len(lowerRanks)-1]
			pool[last] += overflow
			continue
		}

		// 비례 분배 (잔액은 마지막 등수에 몰아줘서 합계 보존)
		distributed := 0
		for idx, lr := range lowerRanks {
			if idx == len(lowerRanks)-1 {
				// 마지막 등수는 나머지 전부
				add := overflow - distributed
				pool[lr] += add
				break
			}
			add := overflow * allocBps[lr] / totalBasicPoints
			distributed += add
			pool[lr] += add
		}
	}
}

func calcFixedPayoutRound(out *RoundOutput, in RoundInput) {
	// 등수별 winners 수와 FixedPayout을 이용해 직접 지급액 계산.
	totalPaid := 0

	for rank, winners := range in.Winners {
		if winners <= 0 {
			continue
		}

		fixed := in.FixedPayout[rank] // 존재하지 않으면 0
		if fixed <= 0 {
			continue
		}

		out.PaidPerWin[rank] = fixed

		total := fixed * winners
		out.PaidTotal[rank] = total
		totalPaid += total
	}
	// 판매액 있으면 잔액 기록
	if in.Sales > totalPaid {
		out.RoundRemainder = in.Sales - totalPaid
	} else {
		out.RoundRemainder = 0
	}
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

func buildAllocationMap(allocs []Allocation) map[Rank]int {
	m := make(map[Rank]int)
	for _, a := range allocs {
		m[a.Rank] = a.BasisPoints
	}
	return m
}
