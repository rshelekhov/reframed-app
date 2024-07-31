package api_tests

import (
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/rshelekhov/reframed/internal/lib/constant/key"
	"github.com/rshelekhov/reframed/internal/lib/middleware/jwtoken"
	"github.com/rshelekhov/reframed/internal/model"
	"math/rand"
	"net/http"
	"testing"
	"time"
)

const (
	scheme = "http"
	host   = "localhost:8082"

	cookieDomain = "localhost"
	cookiePath   = "/"

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

type taskType int

const (
	todayTasks taskType = iota
	upcomingTasks
	overdueTasks
	somedayTasks
)

func randomFakeTask(tt taskType, listID, headingID string) model.TaskRequestData {
	var startDate, startTime, deadline, endTime string

	switch tt {
	case todayTasks:
		startDate = time.Now().Format(time.DateOnly)
	case upcomingTasks:
		startDate = randomDateRange(true, time.Now().AddDate(0, 0, 1), time.Now().AddDate(0, 1, 0))
	case overdueTasks:
		deadline = randomDateRange(true, time.Now().AddDate(0, -1, 0), time.Now().AddDate(0, 0, -1))
	case somedayTasks:
		// Leaving startDate, deadline, startTime, and endTime as empty strings
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
		Tags:        randomTags(),
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
	minValue := rand.Intn(3600)
	maxValue := rand.Intn(3600) + 3600
	return time.Duration(rand.Int63n(int64(maxValue-minValue))+int64(minValue)) * time.Second
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

func randomTags() []string {
	tagCount := rand.Intn(5) + 3
	tags := make([]string, tagCount)
	for i := 0; i < tagCount; i++ {
		tags[i] = gofakeit.Word()
	}
	return tags
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

func createTasks(e *httpexpect.Expect, accessToken string, taskType taskType, lists []*httpexpect.Object, n int) []*httpexpect.Object {
	var tasks []*httpexpect.Object

	for _, list := range lists {
		listID := list.Value(key.Data).Object().Value(key.ListID).String().Raw()

		fakeTask := randomFakeTask(taskType, listID, "")

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

func countTasks(t *testing.T, response *httpexpect.Object, printDetails bool) int {
	formattedData, err := json.MarshalIndent(response.Raw(), "", "  ")
	if err != nil {
		t.Fatalf("Error formatting response to JSON: %v", err)
	}

	var result map[string]interface{}

	err = json.Unmarshal(formattedData, &result)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return 0
	}

	data, ok := result[key.Data].([]interface{})
	if !ok {
		fmt.Println("Error: 'data' is not an array")
		return 0
	}

	totalTasks := 0

	for i := 0; i < len(data); i++ {
		_, ok = data[i].(map[string]interface{})
		if !ok {
			fmt.Println("Error: 'item' is not a map")
			continue
		}
		totalTasks += 1
	}

	if printDetails {
		fmt.Printf("Total tasks: %d\n", totalTasks)
	}

	return totalTasks
}

func countGroups(t *testing.T, response *httpexpect.Object, printDetails bool) int {
	formattedData, err := json.MarshalIndent(response.Raw(), "", "  ")
	if err != nil {
		t.Fatalf("Error formatting response to JSON: %v", err)
	}

	var result map[string]interface{}

	err = json.Unmarshal(formattedData, &result)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return 0
	}

	data, ok := result[key.Data].([]interface{})
	if !ok {
		fmt.Println("Error: 'data' is not an array")
		return 0
	}

	return len(data)
}

func countTasksInGroups(t *testing.T, response *httpexpect.Object, printDetails bool) int {
	formattedData, err := json.MarshalIndent(response.Raw(), "", "  ")
	if err != nil {
		t.Fatalf("Error formatting response to JSON: %v", err)
	}

	var result map[string]interface{}

	err = json.Unmarshal(formattedData, &result)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return 0
	}

	data, ok := result[key.Data].([]interface{})
	if !ok {
		fmt.Println("Error: 'data' is not an array")
		return 0
	}

	totalTasks := 0

	for _, item := range data {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			fmt.Println("Error: 'item' is not a map")
			continue
		}

		tasks, ok := itemMap[key.Tasks].([]interface{})
		if !ok {
			fmt.Printf("Error: '%s' is not an array", key.Tasks)
			continue
		}

		taskCount := len(tasks)
		totalTasks += taskCount

		if printDetails {
			fmt.Printf("Number of %s: %d\n", key.Tasks, taskCount)
		}
	}

	return totalTasks
}

// Clearing the SSO gRPC service storage after testing to avoid data collision
func cleanupAuthService(e *httpexpect.Expect, o *httpexpect.Object) {
	accessToken := o.Value(jwtoken.AccessTokenKey).String().Raw()
	e.DELETE("/user/").
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK)
}
