package parser

import (
	"testing"
)

func TestParseResult(t *testing.T) {
	type args struct {
		output string
	}
	type status struct {
		pass, fail int
	}
	tests := []struct {
		name string
		args args
		want status
	}{
		{
			name: "it counts pass and fail",
			args: args{
				output: `
				=== RUN   TestSum
				=== RUN   TestSum/it_sums_collections_of_any_size
				--- PASS: TestSum (0.00s)
					--- PASS: TestSum/it_sums_collections_of_any_size (0.00s)
				=== RUN   TestSumAll
				--- FAIL: TestSumAll (0.00s)
					sum_test.go:26: want [3 2], got [3 6]
				=== RUN   TestSumAllTails
				=== RUN   TestSumAllTails/it_sums_the_tails_of_slices
				=== RUN   TestSumAllTails/it_sums_the_tails_for_empty_slices
				--- PASS: TestSumAllTails (0.00s)
					--- PASS: TestSumAllTails/it_sums_the_tails_of_slices (0.00s)
					--- PASS: TestSumAllTails/it_sums_the_tails_for_empty_slices (0.00s)
				`,
			},
			want: status{
				pass: 5,
				fail: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pass, fail := ParseResult(tt.args.output)
			if pass != tt.want.pass || fail != tt.want.fail {
				t.Errorf("want %d %d, got %d %d", tt.want.pass, tt.want.fail, pass, fail)
			}
		})
	}
}
