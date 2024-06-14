package api_tests

import (
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/rshelekhov/reframed/internal/lib/constants/key"
	"github.com/rshelekhov/reframed/internal/model"
	"math/rand"
	"net/http"
	"testing"
	"time"
)

const (
	scheme                    = "http"
	host                      = "localhost:8082"
	cookieDomain              = "localhost"
	cookiePath                = "/"
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

func randomTimeInterval() (string, string) {
	gofakeit.Seed(0)

	start := time.Now().AddDate(0, 0, -30)
	end := time.Now()

	firstTime := gofakeit.DateRange(start, end)
	secondTime := gofakeit.DateRange(firstTime, end)

	firstFormattedTime := firstTime.Format(time.DateTime)
	secondFormattedTime := secondTime.Format(time.DateTime)

	return firstFormattedTime, secondFormattedTime
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

func createLists(e *httpexpect.Expect, accessToken string, n int) []*httpexpect.Object {
	var lists []*httpexpect.Object

	for i := 0; i < n; i++ {
		list := e.POST("/user/lists/").
			WithHeader("Authorization", "Bearer "+accessToken).
			WithJSON(model.ListRequestData{
				Title: gofakeit.Word(),
			}).
			Expect().
			Status(http.StatusCreated).
			JSON().Object()

		lists = append(lists, list)
	}

	return lists
}

func createTasks(e *httpexpect.Expect, accessToken string, lists []*httpexpect.Object, n int) []*httpexpect.Object {
	var tasks []*httpexpect.Object

	for _, list := range lists {
		listID := list.Value(key.Data).Object().Value(key.ListID).String().Raw()

		fakeTask := randomFakeTask(true, true, true, true, true, listID, "")

		for i := 0; i < n; i++ {
			task := e.POST("/user/lists/{list_id}/tasks", listID).
				WithHeader("Authorization", "Bearer "+accessToken).
				WithJSON(fakeTask).
				Expect().
				Status(http.StatusCreated).
				JSON().Object()

			tasks = append(tasks, task)
		}
	}

	return tasks
}

func printDataToJSON(t *testing.T, data *httpexpect.Object) {
	formattedData, err := json.MarshalIndent(data.Raw(), "", "  ")
	if err != nil {
		t.Fatalf("Error formatting data to JSON: %v", err)
	}
	fmt.Println(string(formattedData))
}
