package helper

import (
	"fmt"

	"github.com/thanksloving/starriver"
)

type response struct {
	Data         map[string]interface{} `json:"data"` //task result data, it's optional, and can only use by the next task
	Pass         bool                   `json:"pass"` // the task node's result
	FailureLevel starriver.FailureLevel `json:"failure_level"`
	Error        error                  `json:"error"`
	Status       starriver.TaskStatus   `json:"status"`
}

func (r *response) GetStatus() starriver.TaskStatus {
	if r.Status == "" {
		r.Status = starriver.TaskStatusInit
	}
	return r.Status
}

func (r *response) GetError() error {
	return r.Error
}

func (r *response) GetData() map[string]interface{} {
	return r.Data
}

func (r *response) GetFailureLevel() starriver.FailureLevel {
	return r.FailureLevel
}

func (r *response) SetPass(pass bool) {
	r.Pass = pass
}

func (r *response) SetFailureLevel(level starriver.FailureLevel) {
	r.FailureLevel = level
}

func (r *response) IsPass() bool {
	return r.Pass
}

func (r *response) String() string {
	return fmt.Sprintf("[Response]%+v", *r)
}

func NewErrorResponse(err error) starriver.Response {
	return &response{
		Pass:         false,
		FailureLevel: starriver.FailureLevelError,
		Error:        err,
		Status:       starriver.TaskStatusFailure,
	}
}

func NewFatalResponse(err error) starriver.Response {
	return &response{
		Pass:         false,
		FailureLevel: starriver.FailureLevelFatal,
		Error:        err,
		Status:       starriver.TaskStatusFailure,
	}
}

func NewWarnResponse(err error) starriver.Response {
	return &response{
		Pass:         true,
		FailureLevel: starriver.FailureLevelWarning,
		Error:        err,
		Status:       starriver.TaskStatusSuccess,
	}
}

func NewBlockedResponse() starriver.Response {
	return &response{
		Pass:         false,
		FailureLevel: starriver.FailureLevelNormal,
		Status:       starriver.TaskStatusBlocked,
	}
}

func NewSuccessDataResponse(data map[string]interface{}) starriver.Response {
	return &response{
		Pass:         true,
		FailureLevel: starriver.FailureLevelNormal,
		Status:       starriver.TaskStatusSuccess,
		Data:         data,
	}
}

func NewSuccessResponse() starriver.Response {
	return NewSuccessDataResponse(nil)
}
