package schema

import "fmt"

// ValidationError 汇总一个算例的全部物理合法性错误。
// 用 Error() 一次性给出所有问题，方便命令行逐条展示。
type ValidationError struct {
	Issues []string
}

// Error 把所有校验问题用分号拼接为一条错误文本。
func (e *ValidationError) Error() string {
	return "算例校验失败: " + joinIssues(e.Issues)
}

// Valid 返回是否没有任何校验问题。
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

// invalid 是构造校验问题的简写。
func invalid(field, format string, a ...any) string {
	return fmt.Sprintf("%s: %s", field, fmt.Sprintf(format, a...))
}
