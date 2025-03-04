package time

import (
	"time"

	"github.com/dromara/carbon/v2"
)

var (
	c carbon.Carbon
)

func Init(language string) {
	c = carbon.SetLanguage(carbon.NewLanguage().SetLocale(language))
}

func Now() time.Time {
	return c.Now().StdTime()
}

func NowUnix() int64 {
	return c.Now().Timestamp()
}

func Time(timestamp int64) time.Time {
	if timestamp == 0 {
		return time.Time{}
	}
	return c.CreateFromTimestamp(timestamp, c.Timezone()).StdTime()
}

func NextDailyTime(t time.Time, delay time.Duration) time.Time {
	return StartOfDay(t.Add(-delay)).AddDate(0, 0, 1).Add(delay)
}

func NextWeeklyTime(t time.Time, delay time.Duration) time.Time {
	return StartOfWeek(t.Add(-delay)).AddDate(0, 0, 7).Add(delay)
}

func NextMonthlyTime(t time.Time, delay time.Duration) time.Time {
	return StartOfMonth(t.Add(-delay)).AddDate(0, 1, 0).Add(delay)
}

func StartOfDay(t time.Time) time.Time {
	return t.Truncate(24 * time.Hour)
}

func StartOfWeek(t time.Time) time.Time {
	return t.Truncate(7 * 24 * time.Hour)
}

func StartOfMonth(t time.Time) time.Time {
	return carbon.CreateFromStdTime(t).StartOfMonth().StdTime()
}
