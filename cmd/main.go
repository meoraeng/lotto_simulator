package main

import (
	"fmt"

	"github.com/meoraeng/lotto_simulator/internal/lotto"
)

func main() {
	lottos, err := lotto.PurchaseLottos(5000)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("구매 확인 %d장\n", len(lottos.Lottos))
	fmt.Println("발매된 번호")
	for _, t := range lottos.Lottos {
		fmt.Printf("%v\n", t.Numbers)
	}

	if err := lottos.SetWinningNumbers("1,2,3,4,5,6"); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("당첨 번호: %v\n", lottos.WinningNumbers)

	if err := lottos.SetBonusNumber("7"); err != nil {
		fmt.Println(err)
		return
	}
}
