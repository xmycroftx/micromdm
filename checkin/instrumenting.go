package checkin

import (
	"fmt"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/micromdm/mdm"
)

type instrumentingMiddleware struct {
	requestCount   metrics.Counter
	requestLatency metrics.TimeHistogram
	MDMCheckinService
}

func (mw instrumentingMiddleware) Authenticate(cmd mdm.CheckinCommand) (err error) {
	defer func(begin time.Time) {
		methodField := metrics.Field{Key: "MessageType", Value: cmd.MessageType}
		errorField := metrics.Field{Key: "error", Value: fmt.Sprintf("%v", err)}
		mw.requestCount.With(methodField).With(errorField).Add(1)
		mw.requestLatency.With(methodField).With(errorField).Observe(time.Since(begin))
	}(time.Now())
	err = mw.MDMCheckinService.Authenticate(cmd)
	return err
}

func (mw instrumentingMiddleware) TokenUpdate(cmd mdm.CheckinCommand) (err error) {
	defer func(begin time.Time) {
		methodField := metrics.Field{Key: "MessageType", Value: cmd.MessageType}
		errorField := metrics.Field{Key: "error", Value: fmt.Sprintf("%v", err)}
		mw.requestCount.With(methodField).With(errorField).Add(1)
		mw.requestLatency.With(methodField).With(errorField).Observe(time.Since(begin))
	}(time.Now())
	err = mw.MDMCheckinService.TokenUpdate(cmd)
	return err
}

func (mw instrumentingMiddleware) Checkout(cmd mdm.CheckinCommand) (err error) {
	defer func(begin time.Time) {
		methodField := metrics.Field{Key: "MessageType", Value: cmd.MessageType}
		errorField := metrics.Field{Key: "error", Value: fmt.Sprintf("%v", err)}
		mw.requestCount.With(methodField).With(errorField).Add(1)
		mw.requestLatency.With(methodField).With(errorField).Observe(time.Since(begin))
	}(time.Now())
	err = mw.MDMCheckinService.Checkout(cmd)
	return err
}
