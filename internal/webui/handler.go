package webui

import (
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

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

	players := rebuildPlayersFromForm(r, count)

	winningInput := r.FormValue("winningNumbers")
	bonusInput := r.FormValue("bonusNumber")

	allTickets := flattenTickets(players)
	l := lotto.Lottos{Lottos: allTickets}

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

	stats := l.CompileStatistics()
	totalPrize := lotto.CalculateTotalPrize(stats)
	profitRate := lotto.CalculateProfitRate(totalPrize, totalSales)

	rankRows := []rankRowView{
		{
			RankLabel: "1등",
			Condition: "6개 일치",
			Prize:     lotto.PrizeRank1,
			Count:     stats[lotto.Rank1],
		},
		{
			RankLabel: "2등",
			Condition: "5개 일치 + 보너스",
			Prize:     lotto.PrizeRank2,
			Count:     stats[lotto.Rank2],
		},
		{
			RankLabel: "3등",
			Condition: "5개 일치",
			Prize:     lotto.PrizeRank3,
			Count:     stats[lotto.Rank3],
		},
		{
			RankLabel: "4등",
			Condition: "4개 일치",
			Prize:     lotto.PrizeRank4,
			Count:     stats[lotto.Rank4],
		},
		{
			RankLabel: "5등",
			Condition: "3개 일치",
			Prize:     lotto.PrizeRank5,
			Count:     stats[lotto.Rank5],
		},
	}

	data := resultPageData{
		Mode:           mode,
		Players:        players,
		TotalSales:     totalSales,
		WinningNumbers: l.WinningNumbers,
		BonusNumber:    l.BonusNumber,
		RankCounts:     stats,
		TotalPrize:     totalPrize,
		ProfitRate:     profitRate,
		RankRows:       rankRows,
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
