package formats

import "time"

func FormatTime(t *time.Time) string {
	emptytime := time.Time{}
	var r string
	if t.Compare(emptytime) == 0 {
		r = "---"
	} else {
		r = t.Format("02-01-06 15:04:05")
	}
	return r
}

func FormatTimeShort(t *time.Time) string {
	emptytime := time.Time{}
	var r string
	if t.Compare(emptytime) == 0 {
		r = "---"
	} else {
		r = t.Format("15:04:05")
	}
	return r
}

func FormatKey(k string) string {
	return k[:3] + "." + k[len(k)-3:]
}
