package util

import (
	"strings"
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

func TestParseHostt(t *testing.T) {
	type test struct {
		host     string
		wantHost string
		wantPort int
		wantErr  string
	}

	tests := []test{
		{
			host:     "www.example.com",
			wantHost: "www.example.com",
			wantPort: 80,
			wantErr:  "",
		},
		{
			host:     "www.example.com:1234",
			wantHost: "www.example.com",
			wantPort: 1234,
			wantErr:  "",
		},
		{
			host:     "http://www.example.com",
			wantHost: "www.example.com",
			wantPort: 80,
			wantErr:  "",
		},
		{
			host:     "http://www.example.com:1234",
			wantHost: "www.example.com",
			wantPort: 1234,
			wantErr:  "",
		},
		{
			host:     "https://www.example.com",
			wantHost: "www.example.com",
			wantPort: 433,
			wantErr:  "",
		},
		{
			host:     "",
			wantHost: "",
			wantPort: 0,
			wantErr:  "empty host",
		},
		{
			host:     "http://www.example.com:xxx",
			wantHost: "",
			wantPort: 0,
			wantErr:  "port is not a number",
		},
	}

	for _, tc := range tests {
		t.Run(tc.host, func(t *testing.T) {
			h, p, err := ParseHost(tc.host)
			if err != nil {
				if tc.wantErr == "" {
					t.Errorf("Expect no error, got error %q", err)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Errorf("Expect error %q, got error %q", tc.wantErr, err)
				}
			} else {
				if tc.wantErr != "" {
					t.Errorf("Expect error %q, got no error", tc.wantErr)
				} else {
					if h != tc.wantHost {
						t.Errorf("Expect host %q, got %q", tc.wantHost, h)
					}
					if p != tc.wantPort {
						t.Errorf("Expect port %q, got %q", tc.wantPort, p)
					}
				}
			}
		})
	}
}
