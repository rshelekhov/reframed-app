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
)

func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, true, passwordDefaultLength)
}

func randomTimeRange(isSet bool, start, end time.Time) time.Time {
	if isSet {
		return gofakeit.DateRange(start, end)
	}
	return time.Time{}
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

func randomFakeTask(isStartDate, isDeadline, isStartTime, isEndTime, isTags bool, listID, headingID string) model.TaskRequestData {
	return model.TaskRequestData{
		Title:       gofakeit.Sentence(titleDefaultLength),
		Description: gofakeit.Paragraph(paragraphDefaultCount, sentenceDefaultCount, wordDefaultCount, paragraphDefaultSeparator),
		StartDate:   randomTimeRange(isStartDate, time.Now(), time.Now().AddDate(0, 0, 1)),
		Deadline:    randomTimeRange(isDeadline, time.Now(), time.Now().AddDate(0, 0, 1)),
		StartTime:   randomTimeRange(isStartTime, time.Now(), time.Now().AddDate(0, 0, 1)),
		EndTime:     randomTimeRange(isEndTime, time.Now(), time.Now().AddDate(0, 0, 1)),
		ListID:      listID,
		HeadingID:   headingID,
		Tags:        randomTags(isTags),
	}
}
