package api_tests

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/rshelekhov/jwtauth"
	"github.com/rshelekhov/reframed/internal/lib/constant/key"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/segmentio/ksuid"
)

func TestUpdateTaskTime_HappyPath(t *testing.T) {
	u := url.URL{
		Scheme: scheme,
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	// Register user
	user := e.POST("/register").
		WithJSON(model.UserRequestData{
			Email:    gofakeit.Email(),
			Password: randomFakePassword(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	accessToken := user.Value(jwtauth.AccessTokenKey).String().Raw()

	fakeTask := randomFakeTask(upcomingTasks, "", "")

	// Create task
	task := e.POST("/user/lists/default").
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(fakeTask).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	taskID := task.Value(key.Data).Object().Value(key.TaskID).String().Raw()

	startTime, endTime := randomTimeInterval()

	// Update task time
	e.PATCH("/user/tasks/{task_id}/time", taskID).
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(model.TaskRequestTimeData{
			StartTime: startTime,
			EndTime:   endTime,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().Object().NotEmpty()

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}

func TestUpdateTaskTime_FailCases(t *testing.T) {
	u := url.URL{
		Scheme: scheme,
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	// Register user
	user := e.POST("/register").
		WithJSON(model.UserRequestData{
			Email:    gofakeit.Email(),
			Password: randomFakePassword(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	accessToken := user.Value(jwtauth.AccessTokenKey).String().Raw()

	fakeTask := randomFakeTask(upcomingTasks, "", "")

	// Create task
	task := e.POST("/user/lists/default").
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(fakeTask).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	taskID := task.Value(key.Data).Object().Value(key.TaskID).String().Raw()

	startTime, endTime := randomTimeInterval()

	testCases := []struct {
		name      string
		taskID    string
		startTime string
		endTime   string
		status    int
	}{
		{
			name:      "Update task time with empty start time",
			taskID:    taskID,
			startTime: "",
			endTime:   endTime,
			status:    http.StatusBadRequest,
		},
		{
			name:      "Update task time with empty end time",
			taskID:    taskID,
			startTime: startTime,
			endTime:   "",
			status:    http.StatusBadRequest,
		},
		{
			name:      "Update task time when task not found",
			taskID:    ksuid.New().String(),
			startTime: startTime,
			endTime:   endTime,
			status:    http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Update task time
			e.PATCH("/user/tasks/{task_id}/time", tc.taskID).
				WithHeader("Authorization", "Bearer "+accessToken).
				WithJSON(model.TaskRequestTimeData{
					StartTime: tc.startTime,
					EndTime:   tc.endTime,
				}).
				Expect().
				Status(tc.status)
		})
	}

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}
