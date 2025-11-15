package lotto

import "testing"

func TestParseWinningNumbers_Valid(t *testing.T) {
	input := "1,2,3,4,5,6"

	nums, err := parseWinningNumbers(input)
	if err != nil {
		t.Fatalf("정상 입력(%q)에서 에러 발생: %v", input, err)
	}

	expected := []int{1, 2, 3, 4, 5, 6}
	for i := range expected {
		if nums[i] != expected[i] {
			t.Errorf("파싱 오류: nums[%d] = %d, 기대값 %d",
				i, nums[i], expected[i])
		}
	}
}

func TestParseWinningNumbers_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"개수 부족한 경우", "1,2,3,4,5"},
		{"개수 초과한 경우", "1,2,3,4,5,6,7"},
		{"범위 벗어난 값을 받은 경우", "0,2,3,4,5,6"},
		{"중복이 포함된 경우", "1,1,2,3,4,5"},
		{"숫자 아닌 값이 포함된 경우", "1,a,3,4,5,6"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseWinningNumbers(tt.input)
			if err == nil {
				t.Errorf("잘못된 입력(%q)에서 에러가 발생해야 합니다.", tt.input)
			}
		})
	}
}

func TestParseBonusNumber(t *testing.T) {
	winning := []int{1, 2, 3, 4, 5, 6}

	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{"정상 파싱 결과", "7", 7, false},
		{"빈값 파싱 결과", "", 0, true},
		{"문자 파싱 결과", "a", 0, true},
		{"범위 밖의 값 파싱 결과", "0", 0, true},
		{"당첨번호 중복인 경우", "6", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseBonusNumber(tt.input, winning)

			if tt.wantErr && err == nil {
				t.Errorf("잘못된 입력(%q)에서 에러가 발생해야 합니다.", tt.input)
				return
			}
			if !tt.wantErr && err != nil {
				t.Errorf("정상 입력(%q)에서 에러 발생: %v", tt.input, err)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("보너스 번호 파싱 오류: %d, 기대값 %d", got, tt.want)
			}
		})
	}
}
