package webui

import (
	"net/http"
	"strconv"

	"github.com/meoraeng/lotto_simulator/internal/lotto"
)

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/", h.handlePlayer)
	mux.HandleFunc("/purchase", h.handlePurchase)
	mux.HandleFunc("/result", h.handleResult)
}

func (h *Handler) handlePlayer(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handlePlayerGet(w, h)
	case http.MethodPost:
		handlePlayerPost(w, r, h)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handlePlayerGet(w http.ResponseWriter, h *Handler) {
	data := playersPageData{
		Mode: lotto.ModeFixedPayout,
	}
	_ = h.tmpl.ExecuteTemplate(w, "players.gohtml", data)
}

func handlePlayerPost(w http.ResponseWriter, r *http.Request, h *Handler) {
	modeStr := r.FormValue("mode")
	modeInt, _ := strconv.Atoi(modeStr)
	mode := lotto.Mode(modeInt)

	countStr := r.FormValue("playerCount")
	count, err := strconv.Atoi(countStr)

	if !validatePlayerCount(w, h, mode, count, err) {
		return
	}

	roundCountStr := r.FormValue("roundCount")
	roundCount, _ := strconv.Atoi(roundCountStr)
	if roundCount <= 0 {
		roundCount = 1
	}

	url := buildPlayerRedirectURL(mode, count, roundCount)
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func validatePlayerCount(
	w http.ResponseWriter,
	h *Handler,
	mode lotto.Mode,
	count int,
	err error,
) bool {
	if err != nil || count <= 0 {
		data := playersPageData{
			Mode:        mode,
			PlayerCount: count,
			Error:       errorText("플레이어 수는 1 이상 입력해야 합니다"),
		}
		_ = h.tmpl.ExecuteTemplate(w, "players.gohtml", data)
		return false
	}
	return true
}

func (h *Handler) handlePurchase(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handlePurchaseGet(w, r, h)
	case http.MethodPost:
		handlePurchasePost(w, r, h)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handlePurchaseGet(w http.ResponseWriter, r *http.Request, h *Handler) {
	mode, count, roundCount := readModeAndCountFromQuery(r)
	if count <= 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := buildPurchasePageData(mode, count, roundCount, nil, 0, "")
	_ = h.tmpl.ExecuteTemplate(w, "purchase.gohtml", data)
}

func handlePurchasePost(w http.ResponseWriter, r *http.Request, h *Handler) {
	modeInt, _ := strconv.Atoi(r.FormValue("mode"))
	mode := lotto.Mode(modeInt)

	count, _ := strconv.Atoi(r.FormValue("count"))
	if count <= 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	_, _, roundCount := readModeAndCountFromQuery(r)

	players, totalSales, err := parsePlayersFromForm(r, count)
	if err != nil {
		data := buildPurchasePageData(mode, count, roundCount, nil, 0, err.Error())
		_ = h.tmpl.ExecuteTemplate(w, "purchase.gohtml", data)
		return
	}

	data := buildPurchasePageData(mode, count, roundCount, players, totalSales, "")
	_ = h.tmpl.ExecuteTemplate(w, "purchase.gohtml", data)
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

	req := parseResultRequest(r)
	domainPlayers := convertToDomainPlayers(req.Players)

	if req.RoundCount > 1 {
		handleMultipleRounds(w, r, h, req.Mode, req.TotalSales, req.Players, domainPlayers, req.RoundCount)
		return
	}

	handleSingleRound(w, r, h, req, domainPlayers)
}

func handleSingleRound(
	w http.ResponseWriter,
	r *http.Request,
	h *Handler,
	req resultRequest,
	domainPlayers []lotto.Player,
) {
	allTickets := flattenTickets(req.Players)
	l := &lotto.Lottos{Lottos: allTickets}

	if !validateWinningNumbers(w, r, h, req, l) {
		return
	}

	data := buildResultData(req, l, domainPlayers)
	renderResultPage(w, h, data)
}

func validateWinningNumbers(
	w http.ResponseWriter,
	r *http.Request,
	h *Handler,
	req resultRequest,
	l *lotto.Lottos,
) bool {
	if err := l.SetWinningNumbers(req.WinningInput); err != nil {
		renderPurchasePageWithError(w, h, req, err.Error())
		return false
	}

	if err := l.SetBonusNumber(req.BonusInput); err != nil {
		renderPurchasePageWithError(w, h, req, err.Error())
		return false
	}

	return true
}

func renderPurchasePageWithError(
	w http.ResponseWriter,
	h *Handler,
	req resultRequest,
	errorMsg string,
) {
	data := purchasePageData{
		Mode:         req.Mode,
		Count:        req.Count,
		RoundCount:   req.RoundCount,
		IndexList:    makeIndexList(req.Count),
		LottoPrice:   lotto.LottoPrice,
		Error:        errorText(errorMsg),
		Players:      req.Players,
		TotalSales:   req.TotalSales,
		WinningInput: req.WinningInput,
		BonusInput:   req.BonusInput,
	}
	_ = h.tmpl.ExecuteTemplate(w, "purchase.gohtml", data)
}

func renderResultPage(w http.ResponseWriter, h *Handler, data map[string]any) {
	_ = h.tmpl.ExecuteTemplate(w, "result.gohtml", data)
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
	roundResults := make([]roundResultView, 0, roundCount)
	carry := make(map[lotto.Rank]int)
	totalPayouts := make(map[string]int)

	allTickets := flattenTickets(players)

	for round := 1; round <= roundCount; round++ {
		result := processRound(
			r,
			round,
			allTickets,
			mode,
			totalSales,
			domainPlayers,
			carry,
		)
		if result == nil {
			continue
		}

		// 누적 수령액에 합산
		mergePayouts(totalPayouts, result.Payouts)

		// 다음 회차를 위해 이월 상태 업데이트
		carry = result.RoundOutput.CarryOut

		roundResults = append(roundResults, *result)
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

func processRound(
	r *http.Request,
	round int,
	allTickets []lotto.Lotto,
	mode lotto.Mode,
	totalSales int,
	domainPlayers []lotto.Player,
	carry map[lotto.Rank]int,
) *roundResultView {
	winning, ok := parseWinningNumbersForRound(r, round, allTickets)
	if !ok {
		return nil
	}

	stats := winning.CompileStatisticsParallel()
	roundIn := buildRoundInputForModeWithCarry(mode, totalSales, stats, carry)

	roundOut, err := lotto.CalculateRound(roundIn)
	if err != nil {
		return nil
	}

	payouts := lotto.DistributeRewardsParallel(domainPlayers, winning, roundOut)
	rankRows := buildRankRows(mode, stats, roundOut)
	detailRows := buildDetailRowsIfNeeded(mode, roundIn, roundOut)

	return &roundResultView{
		Round:          round,
		WinningNumbers: winning.WinningNumbers,
		BonusNumber:    winning.BonusNumber,
		Stats:          stats,
		RoundInput:     roundIn,
		RoundOutput:    roundOut,
		RankRows:       rankRows,
		DetailRows:     detailRows,
		Payouts:        payouts,
	}
}
