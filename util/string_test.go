package util

import (
	"testing"
)

func TestStringReact(t *testing.T) {
	type test struct {
		s    string
		n    int
		want string
	}

	tests := []test{
		{
			s:    "aaa",
			n:    0,
			want: "aaa",
		},
		{
			s:    "aaa",
			n:    1,
			want: "aax",
		},
		{
			s:    "aaa",
			n:    3,
			want: "xxx",
		},
		{
			s:    "aaa",
			n:    -1,
			want: "aaa",
		},
		{
			s:    "aaa",
			n:    4,
			want: "xxx",
		},
		{
			s:    "",
			n:    1,
			want: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.s, func(t *testing.T) {
			s := RedactLastN(tc.s, tc.n)
			if s != tc.want {
				t.Errorf("Expect %q(%d), got %q", s, tc.n, tc.want)
			}
		})
	}
}
