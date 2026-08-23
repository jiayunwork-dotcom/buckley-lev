package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Parse 从字节流解析算例 JSON。未知字段、尾随垃圾、字段类型错误
// 都视为非法输入并返回错误——避免静默忽略写错的键名。
func Parse(data []byte) (*Case, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, fmt.Errorf("算例为空：没有可解析的 JSON")
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	var c Case
	if err := dec.Decode(&c); err != nil {
		return nil, fmt.Errorf("解析算例 JSON 失败: %w", err)
	}
	// 检查主对象之后是否还有多余内容（例如两个 JSON 拼接）。
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("算例 JSON 在主对象后还有内容（应只有一个对象）")
	}
	return &c, nil
}

// ParseFile 读取并按 Parse 解析一个算例文件。
// 文件不存在、不可读或内容非法时返回带路径的错误。
func ParseFile(path string) (*Case, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取算例文件失败: %w", err)
	}
	c, err := Parse(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return c, nil
}
