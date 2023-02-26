package formats

import "time"

func FormatTime(t *time.Time) string {
	eptytime := time.Time{}
	ft := t.Format("02-01-06 15:04:05")
	if t.Compare(eptytime) == 0 {
		ft = "---"
	}
	return ft
}

func FormatKey(k string) string {
	return k[:5] + "..." + k[len(k)-5:]
}
