package tui

// scrollOffset returns an adjusted offset such that the window
// [offset, offset+size) contains cursor, clamped to [0, total-size].
// total <= size yields offset 0.
func scrollOffset(cursor, offset, size, total int) int {
	if size <= 0 || total <= size {
		return 0
	}
	if cursor < offset {
		offset = cursor
	}
	if cursor >= offset+size {
		offset = cursor - size + 1
	}
	if offset < 0 {
		offset = 0
	}
	if offset > total-size {
		offset = total - size
	}
	return offset
}

// scrollIndicators returns strings to render at the top and bottom of a
// scrolled list. Either may be empty if no indicator is needed.
func scrollIndicators(offset, size, total int) (above, below string) {
	if offset > 0 {
		above = "↑ " + itoa(offset) + " more"
	}
	rest := total - (offset + size)
	if rest > 0 {
		below = "↓ " + itoa(rest) + " more"
	}
	return
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := [20]byte{}
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
