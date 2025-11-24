package lotto

type Mode int

const (
	ModeFixedPayout Mode = iota // 기존의 고정된 상금 모드
	ModeParimutuel              // 판매액 기반 분배 모드
)

// 퍼센트 단위
const BasisPoints = 10_000

// 롤다운 분배 방식
type RollDownMethod int

const (
	RollDownProportional RollDownMethod = iota // 비례 분배 (할당 비율에 비례)
	RollDownEqual                              // 균등 분배 (하위 등수에 균등하게)
)

// 각 등수에 판매액의 몇 bp를 배정할 것인지
type Allocation struct {
	Rank        Rank
	BasisPoints int
}

// 한 회차당 필요한 입력값
type RoundInput struct {
	Mode Mode `json:"mode"`
	// 회차 판매액
	Sales   int          `json:"sales"`
	Winners map[Rank]int `json:"winners"` // 등수별 당첨자 수
	CarryIn map[Rank]int `json:"carryIn"` // 등수별 이월 금액(없으면 0)
	// 분배 모드
	Allocations    []Allocation   `json:"allocations"`    // 등수별 배정 비율
	CapPerRank     map[Rank]int   `json:"capPerRank"`     // 등수별 상한 금액
	RoundingUnit   int            `json:"roundingUnit"`   // 라운딩 단위 (1, 10, 100단위 내림)
	RollDownMethod RollDownMethod `json:"rollDownMethod"` // 롤다운 분배 방식
	// 고정 모드
	FixedPayout map[Rank]int `json:"fixedPayout"`
}

// 한 회차 분배 결과
type RoundOutput struct {
	Sales int `json:"sales"`

	PoolBefore   map[Rank]int `json:"poolBefore"`   // 이월 포함, 상한/롤다운 적용 전 풀 금액
	PoolAfterCap map[Rank]int `json:"poolAfterCap"` // 상한 적용 후 풀 금액
	PaidPerWin   map[Rank]int `json:"paidPerWin"`   // 등수별 1인당 지급액
	PaidTotal    map[Rank]int `json:"paidTotal"`    // 등수별 총 지급액
	CarryOut     map[Rank]int `json:"carryOut"`     // 등수별 다음 회차로 이월되는 금액
	RollDown     map[Rank]int `json:"rollDown"`     // 상한 초과로 하위 등수로 내려보낸 금액

	RoundRemainder int `json:"roundRemainder"` // 판매액 중 풀에 배정되지 않은 라운드 잔액
}

type roundCalculator func(*RoundOutput, RoundInput)

var modeCalculators = map[Mode]roundCalculator{
	ModeParimutuel:  calcParimutuelRound,
	ModeFixedPayout: calcFixedPayoutRound,
}

func CalculateRound(in RoundInput) (RoundOutput, error) {
	if in.Sales < 0 { // 판매액 검증
		return RoundOutput{}, ErrNegativeSales
	}

	out := newRoundOutput(in)

	calc, exists := modeCalculators[in.Mode]
	if !exists {
		return RoundOutput{}, ErrInvalidMode
	}

	calc(&out, in)
	return out, nil
}

func calcParimutuelRound(out *RoundOutput, in RoundInput) {
	order := []Rank{Rank1, Rank2, Rank3, Rank4, Rank5}
	allocBps := buildAllocationMap(in.Allocations)

	allocatedFromSales := calcBasePools(out, in, order, allocBps)
	// 잔액 기록
	out.RoundRemainder = in.Sales - allocatedFromSales
	if out.RoundRemainder < 0 {
		out.RoundRemainder = 0
	}

	applyCapAndRolldown(out.PoolAfterCap, in.CapPerRank, allocBps, order, out.RollDown, in.RollDownMethod)
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
		// 라운딩 단위 설정(0이하면 1원 단위 취급)
		unit := in.RoundingUnit
		if unit <= 0 {
			unit = 1
		}

		// 1인당 받아야 할 금액
		rawPer := pool / winners

		// 라운딩 규칙 적용(unit 단위로 내림)
		roundedPer := (rawPer / unit) * unit

		//모든 당첨자에게 roundedPer씩 지급
		total := roundedPer * winners

		out.PaidPerWin[r] = roundedPer
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
	method RollDownMethod,
) {
	for i, r := range order {
		overflow := calculateOverflow(pool, caps, r, rollDown)
		if overflow <= 0 {
			continue
		}

		lowerRanks := order[i+1:]
		if len(lowerRanks) == 0 {
			continue
		}

		distributeOverflow(pool, lowerRanks, overflow, allocBps, method)
	}
}

func calculateOverflow(pool map[Rank]int, caps map[Rank]int, rank Rank, rollDown map[Rank]int) int {
	cap, hasCap := caps[rank]
	if !hasCap {
		return 0
	}
	if pool[rank] <= cap {
		return 0
	}

	overflow := pool[rank] - cap
	pool[rank] = cap
	rollDown[rank] += overflow
	return overflow
}

func distributeOverflow(
	pool map[Rank]int,
	lowerRanks []Rank,
	overflow int,
	allocBps map[Rank]int,
	method RollDownMethod,
) {
	if method == RollDownEqual {
		distributeEqually(pool, lowerRanks, overflow)
		return
	}
	distributeProportionally(pool, lowerRanks, overflow, allocBps)
}

func distributeEqually(pool map[Rank]int, lowerRanks []Rank, overflow int) {
	perRank := overflow / len(lowerRanks)
	remainder := overflow % len(lowerRanks)

	for idx, lr := range lowerRanks {
		add := perRank
		if idx == len(lowerRanks)-1 {
			add += remainder
		}
		pool[lr] += add
	}
}

func distributeProportionally(
	pool map[Rank]int,
	lowerRanks []Rank,
	overflow int,
	allocBps map[Rank]int,
) {
	totalBasicPoints := calculateTotalBasicPoints(lowerRanks, allocBps)
	if totalBasicPoints == 0 {
		pool[lowerRanks[len(lowerRanks)-1]] += overflow
		return
	}

	distributed := 0
	for idx, lr := range lowerRanks {
		if idx == len(lowerRanks)-1 {
			pool[lr] += overflow - distributed
			break
		}
		add := overflow * allocBps[lr] / totalBasicPoints
		distributed += add
		pool[lr] += add
	}
}

func calculateTotalBasicPoints(lowerRanks []Rank, allocBps map[Rank]int) int {
	total := 0
	for _, lr := range lowerRanks {
		total += allocBps[lr]
	}
	return total
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
	remainder := in.Sales - totalPaid
	if remainder < 0 {
		remainder = 0
	}
	out.RoundRemainder = remainder
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
