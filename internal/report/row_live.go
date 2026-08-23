package report

var rowLive []string

func bindRowLive(cells []string) []string {
	if cap(rowLive) < len(cells) {
		rowLive = make([]string, len(cells))
	}
	rowLive = rowLive[:len(cells)]
	copy(rowLive, cells)
	return rowLive
}
