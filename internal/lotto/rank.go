package lotto

type Rank int

const (
	RankNone Rank = iota
	Rank5
	Rank4
	Rank3
	Rank2
	Rank1
)

const (
	PrizeRank1 = 2_000_000_000
	PrizeRank2 = 30_000_000
	PrizeRank3 = 1_500_000
	PrizeRank4 = 50_000
	PrizeRank5 = 5_000
)

var prizes = [...]int{
	RankNone: 0,
	Rank5:    PrizeRank5,
	Rank4:    PrizeRank4,
	Rank3:    PrizeRank3,
	Rank2:    PrizeRank2,
	Rank1:    PrizeRank1,
}

func (r Rank) Prize() int {
	return prizes[r]
}

func DetermineRank(matchCount int, hasBonus bool) Rank {
	if matchCount == 6 {
		return Rank1
	}
	if matchCount == 5 {
		if hasBonus {
			return Rank2
		}
		return Rank3
	}
	if matchCount == 4 {
		return Rank4
	}
	if matchCount == 3 {
		return Rank5
	}
	return RankNone
}
