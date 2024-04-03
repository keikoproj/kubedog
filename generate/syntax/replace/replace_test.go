package replace

import (
	"reflect"
	"testing"
)

// func TestReplacement_Replace(t *testing.T) {
// 	type fields struct {
// 		Replacee string
// 		Replacer string
// 	}
// 	type args struct {
// 		src string
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 		want   string
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			r := Replacement{
// 				Replacee: tt.fields.Replacee,
// 				Replacer: tt.fields.Replacer,
// 			}
// 			if got := r.Replace(tt.args.src); got != tt.want {
// 				t.Errorf("Replacement.Replace() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestReplacements_Replace(t *testing.T) {
// 	type args struct {
// 		src string
// 	}
// 	tests := []struct {
// 		name string
// 		rs   Replacements
// 		args args
// 		want string
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := tt.rs.Replace(tt.args.src); got != tt.want {
// 				t.Errorf("Replacements.Replace() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func TestBracketsReplacement_Replace(t *testing.T) {
	type fields struct {
		Opening Replacement
		Closing Replacement
	}
	type args struct {
		src string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "Positive Test: '(?:' & ' )?' Case 1",
			fields: fields{
				Opening: Replacement{
					Replacee: "(?:",
					Replacer: "[",
				},
				Closing: Replacement{
					Replacee: " )?",
					Replacer: "] ",
				},
			},
			args: args{
				src: `(?:I )?wait (?:for )?(\d+) (minutes|seconds)`,
			},
			want: `[I] wait [for] (\d+) (minutes|seconds)`,
		},
		{
			name: "Positive Test: '(?:' & ' )?' Case 2",
			fields: fields{
				Opening: Replacement{
					Replacee: "(?:",
					Replacer: "[",
				},
				Closing: Replacement{
					Replacee: " )?",
					Replacer: "] ",
				},
			},
			args: args{
				src: `(?:all )?(?:the )?(?:pod|pods) in (?:the )?namespace (\S+) with (?:the )?label selector (\S+) (?:should )?converge to (?:the )?field selector (\S+)`,
			},
			want: `[all] [the] (?:pod|pods) in [the] namespace (\S+) with [the] label selector (\S+) [should] converge to [the] field selector (\S+)`,
		},
		{
			name: "Positive Test: '(?:' & ')' Case 1",
			fields: fields{
				Opening: Replacement{
					Replacee: "(?:",
					Replacer: "(",
				},
				Closing: Replacement{
					Replacee: ")",
					Replacer: ")",
				},
			},
			args: args{
				src: `[all] [the] (?:pod|pods) in [the] namespace (\S+) with [the] label selector (\S+) [should] converge to [the] field selector (\S+)`,
			},
			want: `[all] [the] (pod|pods) in [the] namespace (\S+) with [the] label selector (\S+) [should] converge to [the] field selector (\S+)`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := BracketsReplacement{
				Opening: tt.fields.Opening,
				Closing: tt.fields.Closing,
			}
			if got := br.Replace(tt.args.src); got != tt.want {
				t.Errorf("BracketsReplacement.Replace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBracketsReplacement_replaceSingle(t *testing.T) {
	type fields struct {
		Opening Replacement
		Closing Replacement
	}
	type args struct {
		src []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []byte
	}{
		{
			name: "Positive Test",
			fields: fields{
				Opening: Replacement{
					Replacee: "(?:",
					Replacer: "(",
				},
				Closing: Replacement{
					Replacee: " )",
					Replacer: ")",
				},
			},
			args: args{
				src: []byte("(?:I )"),
			},
			want: []byte("(I)"),
		},
		{
			name: "Positive Test",
			fields: fields{
				Opening: Replacement{
					Replacee: "(?:",
					Replacer: "(",
				},
				Closing: Replacement{
					Replacee: ")",
					Replacer: ")",
				},
			},
			args: args{
				src: []byte("(?:pod|pods)"),
			},
			want: []byte("(pod|pods)"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := BracketsReplacement{
				Opening: tt.fields.Opening,
				Closing: tt.fields.Closing,
			}
			if got := br.replaceSingle(tt.args.src); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BracketsReplacement.replaceSingle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBracketsReplacement_getRegExp(t *testing.T) {
	type fields struct {
		Opening Replacement
		Closing Replacement
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Positive Test",
			fields: fields{
				Opening: Replacement{
					Replacee: "(?:",
					Replacer: "(",
				},
				Closing: Replacement{
					Replacee: " )",
					Replacer: ")",
				},
			},
			want: `\(\?\:` + regExp_CharsWithinBrackets + `\)`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := BracketsReplacement{
				Opening: tt.fields.Opening,
				Closing: tt.fields.Closing,
			}
			if got := br.getRegExp(); got != tt.want {
				t.Errorf("BracketsReplacement.getRegExp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_escapeEveryCharacter(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Positive Test",
			args: args{
				s: "(?:",
			},
			want: `\(\?\:`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := escapeEveryCharacter(tt.args.s); got != tt.want {
				t.Errorf("escapeEveryCharacter() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func TestBracketsReplacements_Replace(t *testing.T) {
// 	type args struct {
// 		src string
// 	}
// 	tests := []struct {
// 		name string
// 		brs  BracketsReplacements
// 		args args
// 		want string
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := tt.brs.Replace(tt.args.src); got != tt.want {
// 				t.Errorf("BracketsReplacements.Replace() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
