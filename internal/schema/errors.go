package schema

import "fmt"

type ValidationError struct {
	Issues []string
}

func (e *ValidationError) Error() string {
	return "算例校验失败: " + joinIssues(e.Issues)
}

func (e *ValidationError) Valid() bool { return e == nil || len(e.Issues) == 0 }

func joinIssues(issues []string) string {
	out := ""
	for i, s := range issues {
		if i > 0 {
			out += "; "
		}
		out += s
	}
	return out
}

func invalid(field, format string, a ...any) string {
	return fmt.Sprintf("%s: %s", field, fmt.Sprintf(format, a...))
}
