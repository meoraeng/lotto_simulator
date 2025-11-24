package webui

import (
	"html/template"

	"github.com/meoraeng/lotto_simulator/internal/lotto"
)

type Handler struct {
	tmpl *template.Template
}

type playersPageData struct {
	Mode        lotto.Mode
	PlayerCount int
	Error       string
}

type playerTicketsView struct {
	Name    string
	Amount  int
	Tickets []lotto.Lotto
}

type purchasePageData struct {
	Mode       lotto.Mode
	Count      int
	RoundCount int
	IndexList  []int
	LottoPrice int
	Error      string

	Players    []playerTicketsView
	TotalSales int

	// invalid 값을 입력받은 경우 input창 유지를 위한 필드
	WinningInput string
	BonusInput   string
}

type rankRowView struct {
	RankLabel string
	Condition string
	Prize     int
	Count     int
}

type detailRowView struct {
	RankLabel  string
	PoolBefore int
	PoolAfter  int
	RollDown   int
	Winners    int
	PerWin     int
	Total      int
	Carry      int
}

type playerSummary struct {
	Name        string
	Amount      int
	TicketCount int
	Earned      int
	ProfitRate  float64
}

type resultRequest struct {
	Mode         lotto.Mode
	Count        int
	TotalSales   int
	RoundCount   int
	Players      []playerTicketsView
	WinningInput string
	BonusInput   string
}

type roundResultView struct {
	Round          int
	WinningNumbers []int
	BonusNumber    int
	Stats          map[lotto.Rank]int
	RoundInput     lotto.RoundInput
	RoundOutput    lotto.RoundOutput
	RankRows       []rankRowView
	DetailRows     []detailRowView
	Payouts        map[string]int
}
