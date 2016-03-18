package checkin

import (
	"time"

	"github.com/go-kit/kit/log"
	"github.com/micromdm/mdm"
)

type loggingMiddleware struct {
	logger log.Logger
	MDMCheckinService
}

func (mw loggingMiddleware) Authenticate(cmd mdm.CheckinCommand) (err error) {
	defer func(begin time.Time) {
		_ = mw.logger.Log(
			"MessageType", cmd.MessageType,
			"err", err,
			"udid", cmd.UDID,
			"took", time.Since(begin),
		)
	}(time.Now())
	err = mw.MDMCheckinService.Authenticate(cmd)
	return err
}

func (mw loggingMiddleware) TokenUpdate(cmd mdm.CheckinCommand) (err error) {
	defer func(begin time.Time) {
		_ = mw.logger.Log(
			"MessageType", cmd.MessageType,
			"err", err,
			"udid", cmd.UDID,
			"took", time.Since(begin),
		)
	}(time.Now())
	err = mw.MDMCheckinService.TokenUpdate(cmd)
	return err
}

func (mw loggingMiddleware) Checkout(cmd mdm.CheckinCommand) (err error) {
	defer func(begin time.Time) {
		_ = mw.logger.Log(
			"MessageType", cmd.MessageType,
			"err", err,
			"udid", cmd.UDID,
			"took", time.Since(begin),
		)
	}(time.Now())
	err = mw.MDMCheckinService.Checkout(cmd)
	return err
}
