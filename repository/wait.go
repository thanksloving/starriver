package repository

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/thanksloving/starriver"
	"github.com/thanksloving/starriver/helper"
	"github.com/thanksloving/starriver/registry"
)

type (
	Waiting struct {
		helper.SkeletonWithParameter
	}

	WaitingParams struct {
		WaitingTime string
	}
)

var _ starriver.Executable = (*Waiting)(nil)

func registerWait() {
	registry.Register("Wait", "暂停一段时间",
		func(id string) starriver.Executable {
			return &Waiting{helper.NewSkeletonWithParameter(id, &WaitingParams{})}
		},
		registry.Input([]starriver.InputParam{
			{
				Key:      "WaitingTime",
				Required: true,
				Desc:     "暂停时长，支持 d|h|m|s|ms，如 1h 表示一个小时",
			},
		}),
	)
}

func (w *Waiting) Before(ctx starriver.DataContext) {
	ctx.Debug("wait start")
}

func (w *Waiting) After(ctx starriver.DataContext) {
	ctx.Debug("wait end")
}

func (w *Waiting) Execute(ctx starriver.DataContext, params interface{}) starriver.Response {
	p := params.(*WaitingParams)
	d, err := parseDuration(p.WaitingTime)
	if err != nil {
		return helper.NewErrorResponse(err)
	}
	tc := time.NewTimer(d).C
	select {
	case <-tc:
	case <-ctx.Context().Done():
		return helper.NewWarnResponse(ctx.Context().Err())
	}
	return helper.NewSuccessResponse()
}

var durationRE = regexp.MustCompile("^([0-9]+)(d|h|m|s|ms)$")

func parseDuration(durationStr string) (time.Duration, error) {
	matches := durationRE.FindStringSubmatch(durationStr)
	if len(matches) != 3 {
		return 0, fmt.Errorf("not a valid duration string: %q", durationStr)
	}
	var (
		n, _ = strconv.Atoi(matches[1])
		dur  = time.Duration(n) * time.Millisecond
	)
	switch unit := matches[2]; unit {
	case "d":
		dur *= 1000 * 60 * 60 * 24
	case "h":
		dur *= 1000 * 60 * 60
	case "m":
		dur *= 1000 * 60
	case "s":
		dur *= 1000
	case "ms":
	default:
		return 0, fmt.Errorf("invalid time unit in duration string: %q", unit)
	}
	return dur, nil
}
