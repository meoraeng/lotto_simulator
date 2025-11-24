package webui

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/meoraeng/lotto_simulator/internal/lotto"
)

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

func parseResultRequest(r *http.Request) resultRequest {
	modeInt, _ := strconv.Atoi(r.FormValue("mode"))
	mode := lotto.Mode(modeInt)

	count, _ := strconv.Atoi(r.FormValue("count"))
	totalSales, _ := strconv.Atoi(r.FormValue("totalSales"))
	players := rebuildPlayersFromForm(r, count)

	roundCount, _ := strconv.Atoi(r.FormValue("rounds"))
	if roundCount <= 0 {
		roundCount = 1
	}

	winningInput := r.FormValue("winningNumbers")
	bonusInput := r.FormValue("bonusNumber")

	return resultRequest{
		Mode:         mode,
		Count:        count,
		TotalSales:   totalSales,
		RoundCount:   roundCount,
		Players:      players,
		WinningInput: winningInput,
		BonusInput:   bonusInput,
	}
}

func convertToDomainPlayers(players []playerTicketsView) []lotto.Player {
	domainPlayers := make([]lotto.Player, 0, len(players))
	for _, p := range players {
		domainPlayers = append(domainPlayers, lotto.Player{
			Name:    p.Name,
			Tickets: p.Tickets,
		})
	}
	return domainPlayers
}

func parsePlayersFromForm(r *http.Request, count int) ([]playerTicketsView, int, error) {
	var players []playerTicketsView
	totalSales := 0

	for i := 1; i <= count; i++ {
		player, err := parseSinglePlayerFromForm(r, i)
		if err != nil {
			return nil, 0, err
		}

		totalSales += player.Amount
		players = append(players, player)
	}

	return players, totalSales, nil
}

func parseSinglePlayerFromForm(r *http.Request, index int) (playerTicketsView, error) {
	idx := strconv.Itoa(index)
	name := strings.TrimSpace(r.FormValue("name" + idx))
	amountStr := r.FormValue("amount" + idx)
	amount, _ := strconv.Atoi(amountStr)

	if name == "" {
		return playerTicketsView{}, fmt.Errorf("이름은 비울 수 없습니다")
	}

	lottos, err := lotto.PurchaseLottos(amount)
	if err != nil {
		return playerTicketsView{}, err
	}

	return playerTicketsView{
		Name:    name,
		Amount:  amount,
		Tickets: lottos.Lottos,
	}, nil
}

func rebuildPlayersFromForm(r *http.Request, count int) []playerTicketsView {
	players := make([]playerTicketsView, 0, count)

	for i := 1; i <= count; i++ {
		idx := strconv.Itoa(i)
		player := rebuildPlayerFromForm(r, idx)
		players = append(players, player)
	}

	return players
}

func rebuildPlayerFromForm(r *http.Request, idx string) playerTicketsView {
	name := r.FormValue("name" + idx)
	amountStr := r.FormValue("amount" + idx)
	amount, _ := strconv.Atoi(amountStr)
	tickets := rebuildTicketsFromForm(r, idx)

	return playerTicketsView{
		Name:    name,
		Amount:  amount,
		Tickets: tickets,
	}
}

func rebuildTicketsFromForm(r *http.Request, idx string) []lotto.Lotto {
	var tickets []lotto.Lotto
	ticketIdx := 0

	for {
		ticketKey := fmt.Sprintf("ticket_%s_%d", idx, ticketIdx)
		ticketValue := r.FormValue(ticketKey)
		if ticketValue == "" {
			break
		}

		lotto := parseTicketFromString(ticketValue)
		if lotto != nil {
			tickets = append(tickets, *lotto)
		}

		ticketIdx++
	}

	return tickets
}

func parseTicketFromString(ticketValue string) *lotto.Lotto {
	parts := strings.Split(ticketValue, ",")
	numbers := parseNumbersFromParts(parts)

	if len(numbers) != 6 {
		return nil
	}

	return &lotto.Lotto{Numbers: numbers}
}

func parseNumbersFromParts(parts []string) []int {
	numbers := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, _ := strconv.Atoi(p)
		numbers = append(numbers, n)
	}
	return numbers
}

func parseWinningNumbersForRound(
	r *http.Request,
	round int,
	allTickets []lotto.Lotto,
) (lotto.Lottos, bool) {
	winningKey := fmt.Sprintf("winningNumbers_%d", round)
	bonusKey := fmt.Sprintf("bonusNumber_%d", round)

	winningInput := r.FormValue(winningKey)
	bonusInput := r.FormValue(bonusKey)

	if winningInput == "" || bonusInput == "" {
		return lotto.Lottos{}, false
	}

	var winning lotto.Lottos
	winning.Lottos = allTickets

	if err := winning.SetWinningNumbers(winningInput); err != nil {
		return lotto.Lottos{}, false
	}
	if err := winning.SetBonusNumber(bonusInput); err != nil {
		return lotto.Lottos{}, false
	}

	return winning, true
}
