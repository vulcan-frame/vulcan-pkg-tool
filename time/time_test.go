package xtime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	Init("zh-CN")
	assert.NotNil(t, c)
}

func TestNow(t *testing.T) {
	Init("en")
	now := Now()
	assert.True(t, now.Before(time.Now().Add(time.Second)))
	assert.True(t, now.After(time.Now().Add(-time.Second)))
}

func TestNowUnix(t *testing.T) {
	Init("en")
	unix := NowUnix()
	assert.True(t, unix > 0)
	assert.True(t, unix <= time.Now().Unix())
}

func TestTime(t *testing.T) {
	Init("en")
	tests := []struct {
		name      string
		timestamp int64
		want      time.Time
	}{
		{
			name:      "zero timestamp",
			timestamp: 0,
			want:      time.Time{},
		},
		{
			name:      "valid timestamp",
			timestamp: 1577836800, // 2021-01-01 00:00:00
			want:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Time(tt.timestamp)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNextDailyTime(t *testing.T) {
	Init("en")
	tests := []struct {
		name     string
		now      time.Time
		delay    time.Duration
		expected time.Time
	}{
		{
			name:     "normal case - next day",
			now:      time.Date(2020, 3, 15, 10, 0, 0, 0, time.UTC),
			delay:    2 * time.Hour,
			expected: time.Date(2020, 3, 16, 2, 0, 0, 0, time.UTC),
		},
		{
			name:     "delay crosses midnight",
			now:      time.Date(2020, 3, 15, 23, 0, 0, 0, time.UTC),
			delay:    3 * time.Hour,
			expected: time.Date(2020, 3, 16, 3, 0, 0, 0, time.UTC),
		},
		{
			name:     "not trigger next day",
			now:      time.Date(2020, 3, 15, 2, 0, 0, 0, time.UTC),
			delay:    3 * time.Hour,
			expected: time.Date(2020, 3, 15, 3, 0, 0, 0, time.UTC),
		},
		{
			name:     "month end",
			now:      time.Date(2020, 3, 31, 20, 0, 0, 0, time.UTC),
			delay:    5 * time.Hour,
			expected: time.Date(2020, 4, 1, 5, 0, 0, 0, time.UTC),
		},
		{
			name:     "year end",
			now:      time.Date(2020, 12, 31, 22, 0, 0, 0, time.UTC),
			delay:    4 * time.Hour,
			expected: time.Date(2021, 1, 1, 4, 0, 0, 0, time.UTC),
		},
		{
			name:     "zero delay",
			now:      time.Date(2020, 3, 15, 10, 0, 0, 0, time.UTC),
			delay:    0,
			expected: time.Date(2020, 3, 16, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := NextDailyTime(tt.now, tt.delay)
			assert.Equal(t, tt.expected, next)
		})
	}
}

func TestNextWeeklyTime(t *testing.T) {
	Init("en")
	tests := []struct {
		name     string
		now      time.Time
		delay    time.Duration
		expected time.Time
	}{
		{
			name:     "normal case - next week",
			now:      time.Date(2020, 3, 15, 10, 0, 0, 0, time.UTC),
			delay:    3 * 24 * time.Hour,
			expected: time.Date(2020, 3, 19, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "delay crosses midnight",
			now:      time.Date(2020, 3, 15, 23, 0, 0, 0, time.UTC),
			delay:    3 * time.Hour,
			expected: time.Date(2020, 3, 16, 3, 0, 0, 0, time.UTC),
		},
		{
			name:     "not trigger next week",
			now:      time.Date(2020, 3, 10, 2, 0, 0, 0, time.UTC),
			delay:    3*24*time.Hour + 2*time.Hour,
			expected: time.Date(2020, 3, 12, 2, 0, 0, 0, time.UTC),
		},
		{
			name:     "month end",
			now:      time.Date(2020, 3, 31, 20, 0, 0, 0, time.UTC),
			delay:    5 * time.Hour,
			expected: time.Date(2020, 4, 6, 5, 0, 0, 0, time.UTC),
		},
		{
			name:     "year end",
			now:      time.Date(2020, 12, 31, 22, 0, 0, 0, time.UTC),
			delay:    4 * time.Hour,
			expected: time.Date(2021, 1, 4, 4, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := NextWeeklyTime(tt.now, tt.delay)
			assert.Equal(t, tt.expected, next)
		})
	}
}

func TestNextMonthlyTime(t *testing.T) {

	Init("en")
	tests := []struct {
		name     string
		now      time.Time
		delay    time.Duration
		expected time.Time
	}{
		{
			name:     "normal case - next month",
			now:      time.Date(2020, 3, 15, 10, 0, 0, 0, time.UTC),
			delay:    3 * 24 * time.Hour,
			expected: time.Date(2020, 4, 4, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "delay crosses midnight",
			now:      time.Date(2020, 3, 31, 23, 0, 0, 0, time.UTC),
			delay:    3 * time.Hour,
			expected: time.Date(2020, 4, 1, 3, 0, 0, 0, time.UTC),
		},
		{
			name:     "not trigger next month",
			now:      time.Date(2020, 3, 1, 2, 0, 0, 0, time.UTC),
			delay:    3*24*time.Hour + 2*time.Hour,
			expected: time.Date(2020, 3, 4, 2, 0, 0, 0, time.UTC),
		},
		{
			name:     "month with 31 days",
			now:      time.Date(2020, 3, 31, 20, 0, 0, 0, time.UTC),
			delay:    5 * time.Hour,
			expected: time.Date(2020, 4, 1, 5, 0, 0, 0, time.UTC),
		},
		{
			name:     "year end",
			now:      time.Date(2020, 12, 31, 22, 0, 0, 0, time.UTC),
			delay:    4 * time.Hour,
			expected: time.Date(2021, 1, 1, 4, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := NextMonthlyTime(tt.now, tt.delay)
			assert.Equal(t, tt.expected, next)
		})
	}
}

func TestStartOfDay(t *testing.T) {
	Init("en")
	input := time.Date(2024, 3, 15, 10, 30, 45, 0, time.UTC)
	expected := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)

	result := StartOfDay(input)
	assert.Equal(t, expected, result)
}

func TestStartOfWeek(t *testing.T) {
	Init("en")
	input := time.Date(2020, 3, 15, 10, 30, 45, 0, time.UTC)
	expected := time.Date(2020, 3, 9, 0, 0, 0, 0, time.UTC)

	result := StartOfWeek(input)
	assert.Equal(t, expected, result)
}

func TestStartOfMonth(t *testing.T) {
	Init("en")
	input := time.Date(2024, 3, 15, 10, 30, 45, 0, time.UTC)
	expected := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)

	result := StartOfMonth(input)
	assert.Equal(t, expected, result)
}
