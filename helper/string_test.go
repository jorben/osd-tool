package helper

import "testing"

func TestHideSecret(t *testing.T) {
	type Args struct {
		source string
		num    int
		dest   string
	}
	tests := []struct {
		name string
		args Args
	}{
		{
			"empty string",
			Args{
				source: "",
				num:    0,
				dest:   "",
			},
		},
		{
			"empty string2",
			Args{
				source: "",
				num:    8,
				dest:   "",
			},
		},
		{
			"short string",
			Args{
				source: "abc",
				num:    3,
				dest:   "***",
			},
		},
		{
			"short string2",
			Args{
				source: "abc",
				num:    4,
				dest:   "***",
			},
		},
		{
			"abc hide 1 = a*c",
			Args{
				source: "abc",
				num:    1,
				dest:   "a*c",
			},
		},
		{
			"abc hide 2 = a**",
			Args{
				source: "abc",
				num:    2,
				dest:   "a**",
			},
		},
		{
			"long string",
			Args{
				source: "qwertyuiopasdfghjkl",
				num:    8,
				dest:   "qwerty********ghjkl",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsp := HideSecret(tt.args.source, uint32(tt.args.num))
			if rsp != tt.args.dest {
				t.Errorf("HideSecret rsp got %v, want %v", rsp, tt.args.dest)
			}
		})
	}
}

func TestCompareVersion(t *testing.T) {
	type Args struct {
		version1 string
		version2 string
		want     int
	}
	tests := []struct {
		name string
		args Args
	}{
		{
			"0.0.1 = 0.0.1",
			Args{
				version1: "0.0.1",
				version2: "0.0.1",
				want:     0,
			},
		},
		{
			"0.0.1 < 0.0.2",
			Args{
				version1: "0.0.1",
				version2: "0.0.2",
				want:     -1,
			},
		},
		{
			"0.0.2 > 0.0.1",
			Args{
				version1: "0.0.2",
				version2: "0.0.1",
				want:     1,
			},
		},
		{
			"0.0.2 > 0.0.1.1",
			Args{
				version1: "0.0.2",
				version2: "0.0.1.1",
				want:     1,
			},
		},
		{
			"0.0.2.1 > 0.0.1",
			Args{
				version1: "0.0.2.1",
				version2: "0.0.1",
				want:     1,
			},
		},
		{
			"1.2.2 > 1.0.4",
			Args{
				version1: "1.2.2",
				version2: "1.0.4",
				want:     1,
			},
		},
		{
			"v11.0.0 > v9.10.40",
			Args{
				version1: "v11.0.0",
				version2: "v9.10.40",
				want:     1,
			},
		},
		{
			"v0.0.2 > v0.0.1",
			Args{
				version1: "v0.0.2",
				version2: "v0.0.1",
				want:     1,
			},
		},
		{
			"v1.0.2 < v1.0.2.1",
			Args{
				version1: "v1.0.2",
				version2: "v1.0.2.1",
				want:     -1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsp := CompareVersion(tt.args.version1, tt.args.version2)
			if rsp != tt.args.want {
				t.Errorf("CompareVersion rsp got %v, want %v", rsp, tt.args.want)
			}
		})
	}
}
