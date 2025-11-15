package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/meoraeng/lotto_simulator/internal/lotto"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	// 1. 구입 금액 입력
	fmt.Println("구입금액을 입력해 주세요.")
	amountLine, _ := reader.ReadString('\n')
	amountLine = strings.TrimSpace(amountLine)

	amount, err := strconv.Atoi(amountLine)
	if err != nil {
		fmt.Println("[ERROR] 숫자가 아닌 값을 입력했습니다.")
		return
	}

	lottos, err := lotto.PurchaseLottos(amount)
	if err != nil {
		fmt.Println(err) // 여기서 err는 이미 [ERROR] 포맷
		return
	}

	fmt.Printf("\n%d개를 구매했습니다.\n", len(lottos.Lottos))
	for _, t := range lottos.Lottos {
		fmt.Println(formatNumbers(t.Numbers))
	}

	fmt.Println("\n당첨 번호를 입력해 주세요.")
	winningLine, _ := reader.ReadString('\n')
	winningLine = strings.TrimSpace(winningLine)

	if err := lottos.SetWinningNumbers(winningLine); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("\n보너스 번호를 입력해 주세요.")
	bonusLine, _ := reader.ReadString('\n')
	bonusLine = strings.TrimSpace(bonusLine)

	if err := lottos.SetBonusNumber(bonusLine); err != nil {
		fmt.Println(err)
		return
	}

	stats := lottos.CompileStatistics()
	totalPrize := lotto.CalculateTotalPrize(stats)
	profit := lotto.CalculateProfitRate(totalPrize, amount)

	fmt.Println("\n당첨 통계")
	fmt.Println("---")
	// 5등 -> 1등 순서로 출력
	fmt.Printf("3개 일치 (5,000원) - %d개\n", stats[lotto.Rank5])
	fmt.Printf("4개 일치 (50,000원) - %d개\n", stats[lotto.Rank4])
	fmt.Printf("5개 일치 (1,500,000원) - %d개\n", stats[lotto.Rank3])
	fmt.Printf("5개 일치, 보너스 볼 일치 (30,000,000원) - %d개\n", stats[lotto.Rank2])
	fmt.Printf("6개 일치 (2,000,000,000원) - %d개\n", stats[lotto.Rank1])

	fmt.Printf("총 수익률은 %.1f%%입니다.\n", profit)

}

func formatNumbers(nums []int) string {
	var b strings.Builder
	b.WriteString("[")
	for i, n := range nums {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(strconv.Itoa(n))
	}
	b.WriteString("]")
	return b.String()
}
