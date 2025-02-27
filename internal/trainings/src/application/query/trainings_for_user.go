package query

import (
	"context"

	"github.com/rifat-simoom/go-hexarch/internal/shared_kernel/auth"
	"github.com/rifat-simoom/go-hexarch/internal/shared_kernel/decorator"
	"github.com/sirupsen/logrus"
)

type TrainingsForUser struct {
	User auth.User
}

type TrainingsForUserHandler decorator.QueryHandler[TrainingsForUser, []Training]

type trainingsForUserHandler struct {
	readModel TrainingsForUserReadModel
}

func NewTrainingsForUserHandler(
	readModel TrainingsForUserReadModel,
	logger *logrus.Entry,
	metricsClient decorator.MetricsClient,
) TrainingsForUserHandler {
	if readModel == nil {
		panic("nil readModel")
	}

	return decorator.ApplyQueryDecorators[TrainingsForUser, []Training](
		trainingsForUserHandler{readModel: readModel},
		logger,
		metricsClient,
	)
}

type TrainingsForUserReadModel interface {
	FindTrainingsForUser(ctx context.Context, userUUID string) ([]Training, error)
}

func (h trainingsForUserHandler) Handle(ctx context.Context, query TrainingsForUser) (tr []Training, err error) {
	return h.readModel.FindTrainingsForUser(ctx, query.User.UUID)
}
