package ccw

import "testing"

func Test_IsoDurationToMonthsFloat(t *testing.T) {
	type test struct {
		input string
		want  float64
	}

	tests := []test{
		{input: "P0Y0M0DT0H0M", want: 0},
		{input: "P0Y12M0DT0H0M", want: 12},
		{input: "P0Y26M16DT0H0M", want: 26.53},
	}

	for _, tc := range tests {
		got, err := isoDurationToMonthsFloat(tc.input)
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.want {
			t.Logf("exptected: %v, got: %v", tc.want, got)
		}
	}
}

func Test_IsoDurationToDaysFloat(t *testing.T) {
	type test struct {
		input string
		want  float64
	}

	tests := []test{
		{input: "P0Y0M203DT0H0M", want: 203},
		{input: "P0Y0M14DT0H0M", want: 14},
		{input: "P0Y0M3DT0H0M", want: 3},
	}

	for _, tc := range tests {
		got, err := isoDurationToDaysFloat(tc.input)
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.want {
			t.Logf("exptected: %v, got: %v", tc.want, got)
		}
	}
}
