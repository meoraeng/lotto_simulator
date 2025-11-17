package main

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/meoraeng/lotto_simulator/internal/lotto"
)

// -------------------- 입력 처리 --------------------

func readMode(reader *bufio.Reader) lotto.Mode {
	for {
		fmt.Println("모드를 선택해 주세요.")
		fmt.Println("1: 고정 상금 모드")
		fmt.Println("2: 분배(패리뮤추얼) 모드")
		fmt.Print("> ")

		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)

		switch line {
		case "1":
			return lotto.ModeFixedPayout
		case "2":
			return lotto.ModeParimutuel
		default:
			fmt.Println(lotto.NewUserInputError("1 또는 2를 입력해야 합니다."))
		}
	}
}

func readPlayers(reader *bufio.Reader) []playerState {
	for {
		fmt.Print("플레이어 수를 입력해 주세요: ")
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)

		n, err := strconv.Atoi(line)
		if err != nil || n <= 0 {
			fmt.Println(lotto.NewUserInputError("1 이상의 숫자를 입력해야 합니다."))
			continue
		}

		states := make([]playerState, 0, n)
		for i := 0; i < n; i++ {
			fmt.Printf("\n[%d번째 플레이어]\n", i+1)

			name := readPlayerName(reader)
			amount := readPurchaseAmount(reader)

			lottos, err := lotto.PurchaseLottos(amount)
			if err != nil {
				fmt.Println("[FATAL] 로또 구매 중 오류 발생:", err)
				continue
			}

			fmt.Printf("%s님 %d개를 구매했습니다.\n", name, len(lottos.Lottos))
			for _, t := range lottos.Lottos {
				fmt.Println(formatNumbers(t.Numbers))
			}

			states = append(states, playerState{
				Player: lotto.Player{
					Name:    name,
					Tickets: lottos.Lottos,
				},
				PurchaseAmount: amount,
			})
		}

		return states
	}
}

func readPlayerName(reader *bufio.Reader) string {
	for {
		fmt.Print("플레이어 이름을 입력해 주세요: ")
		line, _ := reader.ReadString('\n')
		name := strings.TrimSpace(line)

		if name == "" {
			fmt.Println(lotto.NewUserInputError("빈 이름은 사용할 수 없습니다."))
			continue
		}
		return name
	}
}

func readPurchaseAmount(reader *bufio.Reader) int {
	for {
		fmt.Printf("구입금액을 입력해 주세요 (1장당 %d원):\n", lotto.LottoPrice)
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)

		amount, err := strconv.Atoi(line)
		if err != nil {
			fmt.Println(lotto.NewUserInputError("숫자가 아닌 값을 입력했습니다."))
			continue
		}

		if err := lotto.ValidatePurchaseAmount(amount); err != nil {
			fmt.Println(err)
			continue
		}

		return amount
	}
}

func readWinningNumbers(reader *bufio.Reader, ls *lotto.Lottos) {
	for {
		fmt.Println("\n당첨 번호를 입력해 주세요. (예: 1,2,3,4,5,6)")
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)

		if err := ls.SetWinningNumbers(line); err != nil {
			fmt.Println(err)
			continue
		}
		return
	}
}

func readBonusNumber(reader *bufio.Reader, ls *lotto.Lottos) {
	for {
		fmt.Println("\n보너스 번호를 입력해 주세요.")
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)

		if err := ls.SetBonusNumber(line); err != nil {
			fmt.Println(err)
			continue
		}
		return
	}
}
