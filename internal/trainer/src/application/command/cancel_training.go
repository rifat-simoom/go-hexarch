package command

import (
	"context"
	hour2 "github.com/rifat-simoom/go-hexarch/internal/trainer/src/domain/hour"
	"time"

	"github.com/rifat-simoom/go-hexarch/internal/shared_kernel/decorator"
	"github.com/rifat-simoom/go-hexarch/internal/shared_kernel/errors"
	"github.com/sirupsen/logrus"
)

type CancelTraining struct {
	Hour time.Time
}

type CancelTrainingHandler decorator.CommandHandler[CancelTraining]

type cancelTrainingHandler struct {
	hourRepo hour2.Repository
}

func NewCancelTrainingHandler(
	hourRepo hour2.Repository,
	logger *logrus.Entry,
	metricsClient decorator.MetricsClient,
) CancelTrainingHandler {
	if hourRepo == nil {
		panic("nil hourRepo")
	}

	return decorator.ApplyCommandDecorators[CancelTraining](
		cancelTrainingHandler{hourRepo: hourRepo},
		logger,
		metricsClient,
	)
}

func (h cancelTrainingHandler) Handle(ctx context.Context, cmd CancelTraining) error {
	if err := h.hourRepo.UpdateHour(ctx, cmd.Hour, func(h *hour2.Hour) (*hour2.Hour, error) {
		if err := h.CancelTraining(); err != nil {
			return nil, err
		}
		return h, nil
	}); err != nil {
		return errors.NewSlugError(err.Error(), "unable-to-update-availability")
	}

	return nil
}
