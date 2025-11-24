package webui

import (
	"html/template"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/meoraeng/lotto_simulator/internal/formatter"
)

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
