// Package welge 数值约定。
package welge

const (
	// Eps 是求解器内部的距离容差（饱和度单位）。
	Eps = 1e-12
	// MaxIter 是各类迭代求解的上限，防止坏输入导致死循环。
	MaxIter = 200
)
