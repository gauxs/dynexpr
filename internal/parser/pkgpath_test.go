package parser

import "testing"

func Test_getModulePath(t *testing.T) {
	tests := map[string]struct {
		goModPath string
		want      string
	}{
		"valid go.mod without comments and deps": {
			goModPath: "./data/default.go.mod",
			want:      "example.com/user/project",
		},
		"valid go.mod with comments and without deps": {
			goModPath: "./data/comments.go.mod",
			want:      "example.com/user/project",
		},
		"valid go.mod with comments and deps": {
			goModPath: "./data/comments_deps.go.mod",
			want:      "example.com/user/project",
		},
		"actual dynexpr go.mod": {
			goModPath: "../../go.mod",
			want:      "github.com/gauxs/dynexpr",
		},
		"invalid go.mod with missing module": {
			goModPath: "./data/missing_module.go",
			want:      "",
		},
	}
	for name := range tests {
		tt := tests[name]
		t.Run(name, func(t *testing.T) {
			if got := getModulePath(tt.goModPath); got != tt.want {
				t.Errorf("getModulePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
