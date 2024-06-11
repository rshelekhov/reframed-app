package api_tests

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/rshelekhov/reframed/internal/model"
	"math/rand"
	"time"
)

const (
	host                      = "localhost:8082"
	passwordDefaultLength     = 10
	titleDefaultLength        = 5
	paragraphDefaultCount     = 1
	sentenceDefaultCount      = 5
	wordDefaultCount          = 10
	paragraphDefaultSeparator = " "

	statusID      = 1
	emptyStatusID = 0
	wrongStatusID = 100
)

func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, true, passwordDefaultLength)
}

func randomFakeTask(isStartDate, isDeadline, isStartTime, isEndTime, isTags bool, listID, headingID string) model.TaskRequestData {
	startDate := randomDateRange(isStartDate, time.Now(), time.Now().AddDate(0, 0, 1))
	startTime := randomDateTimeRange(isStartTime, time.Now(), time.Now().AddDate(0, 0, 1))
	var deadline, endTime string

	if isDeadline {
		startDateParsed, _ := time.Parse(time.DateOnly, startDate)
		deadline = randomDateRange(isDeadline, startDateParsed, startDateParsed.Add(randomDays()))
	}
	if isEndTime {
		startTimeParsed, _ := time.Parse(time.DateTime, startTime)
		endTime = randomDateTimeRange(isStartTime, startTimeParsed, startTimeParsed.Add(randomTimeDuration()))
	}

	return model.TaskRequestData{
		Title:       gofakeit.Sentence(titleDefaultLength),
		Description: gofakeit.Paragraph(paragraphDefaultCount, sentenceDefaultCount, wordDefaultCount, paragraphDefaultSeparator),
		StartDate:   startDate,
		Deadline:    deadline,
		StartTime:   startTime,
		EndTime:     endTime,
		ListID:      listID,
		HeadingID:   headingID,
		Tags:        randomTags(isTags),
	}
}

func randomDateRange(isSet bool, start, end time.Time) string {
	if isSet {
		date := gofakeit.DateRange(start, end)
		date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
		return date.Format(time.DateOnly)
	}
	return ""
}

func randomDateTimeRange(isSet bool, start, end time.Time) string {
	if isSet {
		return gofakeit.DateRange(start, end).Format(time.DateTime)
	}
	return ""
}

func randomDays() time.Duration {
	return time.Duration(rand.Int63n(30)) * (24 * time.Hour)
}

func randomTimeDuration() time.Duration {
	min := rand.Intn(3600)
	max := rand.Intn(3600) + 3600
	return time.Duration(rand.Int63n(int64(max-min))+int64(min)) * time.Second
}

func randomTags(isSet bool) []string {
	if isSet {
		tagCount := rand.Intn(5) + 3
		tags := make([]string, tagCount)
		for i := 0; i < tagCount; i++ {
			tags[i] = gofakeit.Word()
		}
		return tags
	}
	return nil
}
