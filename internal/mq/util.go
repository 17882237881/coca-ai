package mq

func orDefaultInt(value int, def int) int {
	if value > 0 {
		return value
	}
	return def
}
