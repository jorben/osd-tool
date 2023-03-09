package helper

import "testing"

func TestInArray(t *testing.T) {
	type Args struct {
		target interface{}
		array  interface{}
		result bool
	}
	tests := []struct {
		name string
		args Args
	}{
		{
			"in, intger array",
			Args{
				123,
				[]int{111, 0, -33, 123, 234, 545},
				true,
			},
		},
		{
			"not in, intger array",
			Args{
				-2,
				[]int{111, 0, -33, 123, 234, 545},
				false,
			},
		},
		{
			"in, uint array",
			Args{
				uint32(0),
				[]uint32{111, 0, 123, 234, 545},
				true,
			},
		},
		{
			"not in, uint array",
			Args{
				uint32(8),
				[]uint32{111, 0, 123, 234, 545},
				false,
			},
		},
		{
			"in, string array",
			Args{
				"123",
				[]string{"111", "0", "-abc33", "123", "234", "545"},
				true,
			},
		},
		{
			"not in, string array",
			Args{
				"-2",
				[]string{"111", "0", "-33", "123", "234", "545"},
				false,
			},
		},
		{
			"not in, empty string array",
			Args{
				"abc",
				[]string{},
				false,
			},
		},
		{
			"not in, empty string, string array",
			Args{
				"",
				[]string{"abc", "123"},
				false,
			},
		},
		{
			"in, bool array",
			Args{
				true,
				[]bool{true, false},
				true,
			},
		},
		{
			"not in, bool array",
			Args{
				false,
				[]bool{true},
				false,
			},
		},
		{
			"in, struct array",
			Args{
				struct {
					s string
					i int
				}{"def", 456},
				[]struct {
					s string
					i int
				}{
					{"abc", 123},
					{"def", 456},
					{"ghi", 789},
				},
				true,
			},
		},
		{
			"not in, struct array",
			Args{
				struct {
					s string
					i int
				}{"abc", 0},
				[]struct {
					s string
					i int
				}{
					{"abc", 123},
					{"def", 456},
					{"ghi", 789},
				},
				false,
			},
		},
		{
			"in, slice array",
			Args{
				[]int{1, 2, 3}[:],
				[][]int{{1, 1, 1}, {1, 2, 3}, {4, 5, 6}},
				true,
			},
		},
		{
			"not in, slice array",
			Args{
				[]int{7, 8, 9},
				[][]int{{1, 1, 1}, {1, 2, 3}, {4, 5, 6}},
				false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsp := InArray(tt.args.target, tt.args.array)
			if rsp != tt.args.result {
				t.Errorf("InArray rsp got %v, want %v", rsp, tt.args.result)
			}
		})
	}
}
