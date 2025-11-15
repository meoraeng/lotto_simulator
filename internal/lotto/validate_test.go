package lotto

import "testing"

func TestValidatePurchaseAmount(t *testing.T) {
	tests := []struct {
		name   string
		amount int
		wantOk bool
	}{
		{"정상 금액 체크", 5000, true},
		{"금액 범위 체크", 0, false},
		{"금액 범위 체크(음수)", -1000, false},
		{"1000원 단위인지 체크", 1500, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePurchaseAmount(tt.amount)

			if tt.wantOk {
				if err != nil {
					t.Errorf("유효한 금액(%d)에서 에러가 발생했습니다: %v", tt.amount, err)
				}
				return
			}

			if err == nil {
				t.Errorf("유효하지 않은 금액(%d)에서 에러가 발생하지 않았습니다.", tt.amount)
			}
		})
	}
}

func TestValidateRange(t *testing.T) {
	tests := []struct {
		value  int
		wantOk bool
	}{
		{1, true},
		{45, true},
		{0, false},
		{46, false},
	}

	for _, tt := range tests {
		err := validateRange(tt.value)

		if tt.wantOk && err != nil {
			t.Errorf("범위 내 값(%d)에서 에러 발생: %v", tt.value, err)
		}
		if !tt.wantOk && err == nil {
			t.Errorf("범위 밖 값(%d)에서 에러가 발생해야 합니다.", tt.value)
		}
	}
}

func TestValidateWinningFormat(t *testing.T) {
	tests := []struct {
		input  string
		wantOk bool
	}{
		{"1,2,3,4,5,6", true},
		{"1;2;3;4;5;6", false},
		{"1a,2,3,4,5,6", false},
		{"", true},
	}

	for _, tt := range tests {
		err := validateWinningFormat(tt.input)

		if tt.wantOk && err != nil {
			t.Errorf("정상 입력(%q)에서 에러 발생: %v", tt.input, err)
		}
		if !tt.wantOk && err == nil {
			t.Errorf("잘못된 입력(%q)에서 에러가 발생하지 않았습니다.", tt.input)
		}
	}
}

func TestValidateNoDuplicates(t *testing.T) {
	tests := []struct {
		nums   []int
		wantOk bool
	}{
		{[]int{1, 2, 3, 4, 5, 6}, true},
		{[]int{1, 1, 2, 3, 4, 5}, false},
	}

	for _, tt := range tests {
		err := validateNoDuplicates(tt.nums)

		if tt.wantOk && err != nil {
			t.Errorf("중복 없는 입력 %v 에서 에러 발생: %v", tt.nums, err)
		}
		if !tt.wantOk && err == nil {
			t.Errorf("중복 있는 입력 %v 에서 에러가 발생해야 합니다.", tt.nums)
		}
	}
}
