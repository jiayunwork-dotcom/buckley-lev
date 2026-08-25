package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

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
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("算例 JSON 在主对象后还有内容（应只有一个对象）")
	}
	return &c, nil
}

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
