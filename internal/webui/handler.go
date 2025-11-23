package webui

import (
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/meoraeng/lotto_simulator/internal/formatter"
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

type resultPageData struct {
	Mode           lotto.Mode
	Players        []playerTicketsView
	TotalSales     int
	WinningNumbers []int
	BonusNumber    int

	RankCounts map[lotto.Rank]int
	TotalPrize int
	ProfitRate float64

	RankRows []rankRowView
}

type playerSummary struct {
	Name        string
	Amount      int
	TicketCount int
	Earned      int
	ProfitRate  float64
}

func NewHandler(templatesDir string) (*Handler, error) {
	funcMap := template.FuncMap{
		"add1": func(i int) int {
			return i + 1
		},
		"joinInts": func(nums []int, sep string) string {
			if len(nums) == 0 {
				return ""
			}
			parts := make([]string, len(nums))
			for i, n := range nums {
				parts[i] = strconv.Itoa(n)
			}
			return strings.Join(parts, sep)
		},
		"money": formatter.Money,
	}

	pattern := filepath.Join(templatesDir, "*.gohtml")
	tmpl, err := template.New("root").Funcs(funcMap).ParseGlob(pattern)
	if err != nil {
		return nil, err
	}

	return &Handler{tmpl: tmpl}, nil
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/", h.handlePlayer)
	mux.HandleFunc("/purchase", h.handlePurchase)
	mux.HandleFunc("/result", h.handleResult)
}

func (h *Handler) handlePlayer(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		data := playersPageData{
			Mode: lotto.ModeFixedPayout,
		}
		_ = h.tmpl.ExecuteTemplate(w, "players.gohtml", data)

	case http.MethodPost:
		modeStr := r.FormValue("mode")
		modeInt, _ := strconv.Atoi(modeStr)
		mode := lotto.Mode(modeInt)

		countStr := r.FormValue("playerCount")
		count, err := strconv.Atoi(countStr)
		if err != nil || count <= 0 {
			data := playersPageData{
				Mode:        mode,
				PlayerCount: count,
				Error:       errorText("플레이어 수는 1 이상 입력해야 합니다"),
			}
			_ = h.tmpl.ExecuteTemplate(w, "players.gohtml", data)
			return
		}

		q := "?mode=" + strconv.Itoa(int(mode)) + "&count=" + strconv.Itoa(count)
		http.Redirect(w, r, "/purchase"+q, http.StatusSeeOther)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handlePurchase(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		mode, count := readModeAndCountFromQuery(r)
		if count <= 0 {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		data := purchasePageData{
			Mode:       mode,
			Count:      count,
			IndexList:  makeIndexList(count),
			LottoPrice: lotto.LottoPrice,
		}
		_ = h.tmpl.ExecuteTemplate(w, "purchase.gohtml", data)

	case http.MethodPost:
		modeInt, _ := strconv.Atoi(r.FormValue("mode"))
		mode := lotto.Mode(modeInt)

		count, _ := strconv.Atoi(r.FormValue("count"))
		if count <= 0 {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		var players []playerTicketsView
		totalSales := 0

		for i := 1; i <= count; i++ {
			idx := strconv.Itoa(i)

			name := strings.TrimSpace(r.FormValue("name" + idx))
			amountStr := r.FormValue("amount" + idx)
			amount, _ := strconv.Atoi(amountStr)

			if name == "" {
				data := purchasePageData{
					Mode:       mode,
					Count:      count,
					IndexList:  makeIndexList(count),
					LottoPrice: lotto.LottoPrice,
					Error:      errorText("이름은 비울 수 없습니다"),
				}
				_ = h.tmpl.ExecuteTemplate(w, "purchase.gohtml", data)
				return
			}

			lottos, err := lotto.PurchaseLottos(amount)
			if err != nil {
				data := purchasePageData{
					Mode:       mode,
					Count:      count,
					IndexList:  makeIndexList(count),
					LottoPrice: lotto.LottoPrice,
					Error:      errorMsg(err),
				}
				_ = h.tmpl.ExecuteTemplate(w, "purchase.gohtml", data)
				return
			}

			totalSales += amount

			players = append(players, playerTicketsView{
				Name:    name,
				Amount:  amount,
				Tickets: lottos.Lottos,
			})
		}

		data := purchasePageData{
			Mode:       mode,
			Count:      count,
			IndexList:  makeIndexList(count),
			LottoPrice: lotto.LottoPrice,
			Players:    players,
			TotalSales: totalSales,
		}
		_ = h.tmpl.ExecuteTemplate(w, "purchase.gohtml", data)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	modeInt, _ := strconv.Atoi(r.FormValue("mode"))
	mode := lotto.Mode(modeInt)

	count, _ := strconv.Atoi(r.FormValue("count"))
	totalSales, _ := strconv.Atoi(r.FormValue("totalSales"))

	// purchase.gohtml에서 넘어온 hidden 필드 기반으로
	// playerTicketsView 슬라이스 복원
	players := rebuildPlayersFromForm(r, count)

	winningInput := r.FormValue("winningNumbers")
	bonusInput := r.FormValue("bonusNumber")

	// 모든 티켓을 한 배열로 평탄화
	allTickets := flattenTickets(players)
	l := lotto.Lottos{Lottos: allTickets}

	// 당첨 번호 파싱/검증
	if err := l.SetWinningNumbers(winningInput); err != nil {
		data := purchasePageData{
			Mode:         mode,
			Count:        count,
			IndexList:    makeIndexList(count),
			LottoPrice:   lotto.LottoPrice,
			Error:        errorText(err.Error()),
			Players:      players,
			TotalSales:   totalSales,
			WinningInput: winningInput,
			BonusInput:   bonusInput,
		}
		_ = h.tmpl.ExecuteTemplate(w, "purchase.gohtml", data)
		return
	}

	// 보너스 번호 파싱/검증
	if err := l.SetBonusNumber(bonusInput); err != nil {
		data := purchasePageData{
			Mode:         mode,
			Count:        count,
			IndexList:    makeIndexList(count),
			LottoPrice:   lotto.LottoPrice,
			Error:        errorText(err.Error()),
			Players:      players,
			TotalSales:   totalSales,
			WinningInput: winningInput,
			BonusInput:   bonusInput,
		}
		_ = h.tmpl.ExecuteTemplate(w, "purchase.gohtml", data)
		return
	}

	// 등수별 통계
	stats := l.CompileStatistics()

	// 도메인 Player로 변환 (티켓 분배 계산에 사용)
	domainPlayers := make([]lotto.Player, 0, len(players))
	for _, p := range players {
		domainPlayers = append(domainPlayers, lotto.Player{
			Name:    p.Name,
			Tickets: p.Tickets,
		})
	}

	// 모드에 따라 RoundInput 생성
	roundIn := buildRoundInputForMode(mode, totalSales, stats)

	roundOut, err := lotto.CalculateRound(roundIn)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 이름별 수령 금액 계산
	payouts := lotto.DistributeRewards(domainPlayers, l, roundOut)

	// 결과 테이블용 행 생성 (모드에 따라 상금이 다름)
	rankRows := buildRankRows(mode, stats, roundOut)

	// 플레이어 요약 뷰
	playerSummaries := buildPlayerSummaries(players, payouts)

	data := map[string]any{
		"Mode":            mode,
		"TotalSales":      totalSales,
		"WinningNumbers":  l.WinningNumbers,
		"BonusNumber":     l.BonusNumber,
		"RankRows":        rankRows,
		"PlayerSummaries": playerSummaries,
	}

	_ = h.tmpl.ExecuteTemplate(w, "result.gohtml", data)
}

func readModeAndCountFromQuery(r *http.Request) (lotto.Mode, int) {
	modeStr := r.URL.Query().Get("mode")
	modeInt, _ := strconv.Atoi(modeStr)
	mode := lotto.Mode(modeInt)

	countStr := r.URL.Query().Get("count")
	count, _ := strconv.Atoi(countStr)

	return mode, count
}

func makeIndexList(n int) []int {
	out := make([]int, n)
	for i := 0; i < n; i++ {
		out[i] = i + 1
	}
	return out
}

func buildPlayerSummaries(players []playerTicketsView, payouts map[string]int) []playerSummary {
	summaries := make([]playerSummary, 0, len(players))

	for _, p := range players {
		spent := p.Amount
		earned := payouts[p.Name]

		rate := 0.0
		if spent > 0 {
			rate = float64(earned) / float64(spent) * 100
		}

		summaries = append(summaries, playerSummary{
			Name:        p.Name,
			Amount:      spent,
			TicketCount: len(p.Tickets),
			Earned:      earned,
			ProfitRate:  rate,
		})
	}

	return summaries
}

func rebuildPlayersFromForm(r *http.Request, count int) []playerTicketsView {
	var players []playerTicketsView

	for i := 1; i <= count; i++ {
		idx := strconv.Itoa(i)

		name := r.FormValue("name" + idx)
		amount, _ := strconv.Atoi(r.FormValue("amount" + idx))

		var tickets []lotto.Lotto
		for t := 0; ; t++ {
			key := "ticket_" + idx + "_" + strconv.Itoa(t)
			raw := r.FormValue(key)
			if raw == "" {
				break
			}

			nums := parseTicketNumbers(raw)
			if len(nums) == 0 {
				continue
			}
			tickets = append(tickets, lotto.Lotto{Numbers: nums})
		}

		players = append(players, playerTicketsView{
			Name:    name,
			Amount:  amount,
			Tickets: tickets,
		})
	}

	return players
}

func parseTicketNumbers(raw string) []int {
	parts := strings.Split(raw, ",")
	nums := make([]int, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			continue
		}
		nums = append(nums, n)
	}

	return nums
}

func flattenTickets(players []playerTicketsView) []lotto.Lotto {
	var all []lotto.Lotto
	for _, p := range players {
		all = append(all, p.Tickets...)
	}
	return all
}

// 모드에 따라 RoundInput 생성
func buildRoundInputForMode(
	mode lotto.Mode,
	totalSales int,
	stats map[lotto.Rank]int,
) lotto.RoundInput {
	if mode == lotto.ModeFixedPayout {
		// 고정 상금 모드
		fixedPayout := map[lotto.Rank]int{
			lotto.Rank1: lotto.Rank1.Prize(),
			lotto.Rank2: lotto.Rank2.Prize(),
			lotto.Rank3: lotto.Rank3.Prize(),
			lotto.Rank4: lotto.Rank4.Prize(),
			lotto.Rank5: lotto.Rank5.Prize(),
		}
		return lotto.RoundInput{
			Mode:        mode,
			Sales:       totalSales,
			Winners:     stats,
			FixedPayout: fixedPayout,
		}
	}

	// 분배 모드
	allocations := []lotto.Allocation{
		{Rank: lotto.Rank1, BasisPoints: 7500},
		{Rank: lotto.Rank2, BasisPoints: 1250},
		{Rank: lotto.Rank3, BasisPoints: 1250},
		{Rank: lotto.Rank4, BasisPoints: 0},
		{Rank: lotto.Rank5, BasisPoints: 0},
	}

	caps := map[lotto.Rank]int{
		lotto.Rank1: 2_000_000_000,
	}

	return lotto.RoundInput{
		Mode:         mode,
		Sales:        totalSales,
		Winners:      stats,
		CarryIn:      map[lotto.Rank]int{},
		Allocations:  allocations,
		CapPerRank:   caps,
		RoundingUnit: 100,
	}
}

// 모드에 따라 rankRows 생성
func buildRankRows(
	mode lotto.Mode,
	stats map[lotto.Rank]int,
	roundOut lotto.RoundOutput,
) []rankRowView {
	ranks := []struct {
		rank      lotto.Rank
		label     string
		condition string
	}{
		{lotto.Rank1, "1등", "6 match"},
		{lotto.Rank2, "2등", "5 + bonus"},
		{lotto.Rank3, "3등", "5 match"},
		{lotto.Rank4, "4등", "4 match"},
		{lotto.Rank5, "5등", "3 match"},
	}

	rows := make([]rankRowView, 0, len(ranks))
	for _, r := range ranks {
		var prize int
		if mode == lotto.ModeFixedPayout {
			prize = r.rank.Prize()
		} else {
			prize = roundOut.PaidPerWin[r.rank]
		}

		rows = append(rows, rankRowView{
			RankLabel: r.label,
			Condition: r.condition,
			Prize:     prize,
			Count:     stats[r.rank],
		})
	}
	return rows
}
