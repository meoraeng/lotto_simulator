package webui

import (
	"fmt"
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
		"seq": func(start, end int) []int {
			if start > end {
				return []int{}
			}
			result := make([]int, end-start+1)
			for i := range result {
				result[i] = start + i
			}
			return result
		},
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

		roundCountStr := r.FormValue("roundCount")
		roundCount, _ := strconv.Atoi(roundCountStr)
		if roundCount <= 0 {
			roundCount = 1
		}

		q := "?mode=" + strconv.Itoa(int(mode)) + "&count=" + strconv.Itoa(count) + "&rounds=" + strconv.Itoa(roundCount)
		http.Redirect(w, r, "/purchase"+q, http.StatusSeeOther)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handlePurchase(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		mode, count, roundCount := readModeAndCountFromQuery(r)
		if count <= 0 {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		data := purchasePageData{
			Mode:       mode,
			Count:      count,
			RoundCount: roundCount,
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

		_, _, roundCount := readModeAndCountFromQuery(r)

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
					RoundCount: roundCount,
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
					RoundCount: roundCount,
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

		_, _, roundCount = readModeAndCountFromQuery(r)
		data := purchasePageData{
			Mode:       mode,
			Count:      count,
			RoundCount: roundCount,
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

	players := rebuildPlayersFromForm(r, count)

	roundCount, _ := strconv.Atoi(r.FormValue("rounds"))
	if roundCount <= 0 {
		roundCount = 1
	}

	// 도메인 Player로 변환
	domainPlayers := make([]lotto.Player, 0, len(players))
	for _, p := range players {
		domainPlayers = append(domainPlayers, lotto.Player{
			Name:    p.Name,
			Tickets: p.Tickets,
		})
	}

	// 다중 회차 처리
	if roundCount > 1 {
		handleMultipleRounds(w, r, h, mode, totalSales, players, domainPlayers, roundCount)
		return
	}

	// 단일 회차 처리
	winningInput := r.FormValue("winningNumbers")
	bonusInput := r.FormValue("bonusNumber")

	// 모든 티켓을 한 배열로 평탄화
	allTickets := flattenTickets(players)
	l := &lotto.Lottos{Lottos: allTickets}

	// 당첨 번호 파싱/검증
	if err := l.SetWinningNumbers(winningInput); err != nil {
		data := purchasePageData{
			Mode:         mode,
			Count:        count,
			RoundCount:   roundCount,
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
			RoundCount:   roundCount,
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
	stats := l.CompileStatisticsParallel()

	// 모드에 따라 RoundInput 생성
	roundIn := buildRoundInputForMode(mode, totalSales, stats)

	roundOut, err := lotto.CalculateRound(roundIn)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 이름별 수령 금액 계산
	payouts := lotto.DistributeRewardsParallel(domainPlayers, *l, roundOut)

	// 결과 테이블용 행 생성
	rankRows := buildRankRows(mode, stats, roundOut)

	playerSummaries := buildPlayerSummaries(players, payouts)

	data := map[string]any{
		"Mode":            mode,
		"TotalSales":      totalSales,
		"WinningNumbers":  l.WinningNumbers,
		"BonusNumber":     l.BonusNumber,
		"RankRows":        rankRows,
		"PlayerSummaries": playerSummaries,
	}

	// 분배 모드일 때 RoundOutput과 detail row 추가
	if mode == lotto.ModeParimutuel {
		data["RoundOutput"] = roundOut
		data["DetailRows"] = buildDetailRows(roundIn, roundOut)
	}

	_ = h.tmpl.ExecuteTemplate(w, "result.gohtml", data)
}

func readModeAndCountFromQuery(r *http.Request) (lotto.Mode, int, int) {
	modeStr := r.URL.Query().Get("mode")
	modeInt, _ := strconv.Atoi(modeStr)
	mode := lotto.Mode(modeInt)

	countStr := r.URL.Query().Get("count")
	count, _ := strconv.Atoi(countStr)

	roundCountStr := r.URL.Query().Get("rounds")
	roundCount, _ := strconv.Atoi(roundCountStr)
	if roundCount <= 0 {
		roundCount = 1
	}

	return mode, count, roundCount
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
	players := make([]playerTicketsView, 0, count)

	for i := 1; i <= count; i++ {
		idx := strconv.Itoa(i)

		name := r.FormValue("name" + idx)
		amountStr := r.FormValue("amount" + idx)
		amount, _ := strconv.Atoi(amountStr)

		var tickets []lotto.Lotto
		ticketIdx := 0
		for {
			ticketKey := fmt.Sprintf("ticket_%s_%d", idx, ticketIdx)
			ticketValue := r.FormValue(ticketKey)
			if ticketValue == "" {
				break
			}

			parts := strings.Split(ticketValue, ",")
			numbers := make([]int, 0, len(parts))
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p == "" {
					continue
				}
				n, _ := strconv.Atoi(p)
				numbers = append(numbers, n)
			}

			if len(numbers) == 6 {
				tickets = append(tickets, lotto.Lotto{Numbers: numbers})
			}

			ticketIdx++
		}

		players = append(players, playerTicketsView{
			Name:    name,
			Amount:  amount,
			Tickets: tickets,
		})
	}

	return players
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
		Mode:           mode,
		Sales:          totalSales,
		Winners:        stats,
		CarryIn:        map[lotto.Rank]int{},
		Allocations:    allocations,
		CapPerRank:     caps,
		RoundingUnit:   100,
		RollDownMethod: lotto.RollDownProportional,
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

// 분배 모드 상세 내용 작성을 위한 행 생성
func buildDetailRows(
	roundIn lotto.RoundInput,
	roundOut lotto.RoundOutput,
) []detailRowView {
	ranks := []struct {
		rank  lotto.Rank
		label string
	}{
		{lotto.Rank1, "1등"},
		{lotto.Rank2, "2등"},
		{lotto.Rank3, "3등"},
		{lotto.Rank4, "4등"},
		{lotto.Rank5, "5등"},
	}

	rows := make([]detailRowView, 0, len(ranks))
	for _, r := range ranks {
		rows = append(rows, detailRowView{
			RankLabel:  r.label,
			PoolBefore: roundOut.PoolBefore[r.rank],
			PoolAfter:  roundOut.PoolAfterCap[r.rank],
			RollDown:   roundOut.RollDown[r.rank],
			Winners:    roundIn.Winners[r.rank],
			PerWin:     roundOut.PaidPerWin[r.rank],
			Total:      roundOut.PaidTotal[r.rank],
			Carry:      roundOut.CarryOut[r.rank],
		})
	}
	return rows
}

// 다중 회차 처리
func handleMultipleRounds(
	w http.ResponseWriter,
	r *http.Request,
	h *Handler,
	mode lotto.Mode,
	totalSales int,
	players []playerTicketsView,
	domainPlayers []lotto.Player,
	roundCount int,
) {
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

	roundResults := make([]roundResultView, 0, roundCount)
	carry := make(map[lotto.Rank]int)
	totalPayouts := make(map[string]int)

	allTickets := flattenTickets(players)

	for round := 1; round <= roundCount; round++ {
		winningKey := fmt.Sprintf("winningNumbers_%d", round)
		bonusKey := fmt.Sprintf("bonusNumber_%d", round)

		winningInput := r.FormValue(winningKey)
		bonusInput := r.FormValue(bonusKey)

		if winningInput == "" || bonusInput == "" {
			continue
		}

		var winning lotto.Lottos
		winning.Lottos = allTickets

		if err := winning.SetWinningNumbers(winningInput); err != nil {
			continue
		}
		if err := winning.SetBonusNumber(bonusInput); err != nil {
			continue
		}

		// 등수별 통계
		stats := winning.CompileStatisticsParallel()

		// 모드에 따라 RoundInput 생성 (이월 포함)
		roundIn := buildRoundInputForModeWithCarry(mode, totalSales, stats, carry)

		roundOut, err := lotto.CalculateRound(roundIn)
		if err != nil {
			continue
		}

		// 이름별 수령 금액 계산
		payouts := lotto.DistributeRewardsParallel(domainPlayers, winning, roundOut)

		// 누적 수령액에 합산
		for name, amount := range payouts {
			totalPayouts[name] += amount
		}

		// 다음 회차를 위해 이월 상태 업데이트
		carry = roundOut.CarryOut

		rankRows := buildRankRows(mode, stats, roundOut)
		var detailRows []detailRowView
		if mode == lotto.ModeParimutuel {
			detailRows = buildDetailRows(roundIn, roundOut)
		}

		roundResults = append(roundResults, roundResultView{
			Round:          round,
			WinningNumbers: winning.WinningNumbers,
			BonusNumber:    winning.BonusNumber,
			Stats:          stats,
			RoundInput:     roundIn,
			RoundOutput:    roundOut,
			RankRows:       rankRows,
			DetailRows:     detailRows,
			Payouts:        payouts,
		})
	}

	// 플레이어별 누적 요약
	playerSummaries := buildPlayerSummaries(players, totalPayouts)

	data := map[string]any{
		"Mode":            mode,
		"TotalSales":      totalSales,
		"RoundCount":      roundCount,
		"RoundResults":    roundResults,
		"PlayerSummaries": playerSummaries,
	}

	_ = h.tmpl.ExecuteTemplate(w, "result_multi.gohtml", data)
}

// 이월 상태를 포함한 RoundInput 생성
func buildRoundInputForModeWithCarry(
	mode lotto.Mode,
	totalSales int,
	stats map[lotto.Rank]int,
	carryIn map[lotto.Rank]int,
) lotto.RoundInput {
	if mode == lotto.ModeFixedPayout {
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
		Mode:           mode,
		Sales:          totalSales,
		Winners:        stats,
		CarryIn:        carryIn,
		Allocations:    allocations,
		CapPerRank:     caps,
		RoundingUnit:   100,
		RollDownMethod: lotto.RollDownProportional,
	}
}
