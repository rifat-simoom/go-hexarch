package integration

import (
	"context"
	"github.com/rifat-simoom/go-hexarch/internal/trainer/src/infrastructure/configs"
	presentation2 "github.com/rifat-simoom/go-hexarch/internal/trainer/src/presentation/grpc"
	http2 "github.com/rifat-simoom/go-hexarch/internal/trainer/src/presentation/http"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	trainerHTTP "github.com/rifat-simoom/go-hexarch/internal/shared_kernel/client/trainer"
	"github.com/rifat-simoom/go-hexarch/internal/shared_kernel/genproto/trainer"
	"github.com/rifat-simoom/go-hexarch/internal/shared_kernel/server"
	"github.com/rifat-simoom/go-hexarch/internal/shared_kernel/tests"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestHoursAvailability(t *testing.T) {
	t.Parallel()

	token := tests.FakeTrainerJWT(t, uuid.New().String())
	client := tests.NewTrainerHTTPClient(t, token)

	hour := tests.RelativeDate(11, 12)
	expectedHour := trainerHTTP.Hour{
		Available:            true,
		HasTrainingScheduled: false,
		Hour:                 hour,
	}

	date := hour.Truncate(24 * time.Hour)
	from := date.AddDate(0, 0, -1)
	to := date.AddDate(0, 0, 1)

	getHours := func() []trainerHTTP.Hour {
		dates := client.GetTrainerAvailableHours(t, from, to)
		for _, d := range dates {
			if d.Date.Equal(date) {
				return d.Hours
			}
		}
		t.Fatalf("Date not found in dates: %+v", dates)
		return nil
	}

	client.MakeHourUnavailable(t, hour)
	require.NotContains(t, getHours(), expectedHour)

	code := client.MakeHourAvailable(t, hour)
	require.Equal(t, http.StatusNoContent, code)
	require.Contains(t, getHours(), expectedHour)

	client.MakeHourUnavailable(t, hour)
	require.NotContains(t, getHours(), expectedHour)
}

func TestUnauthorizedForAttendee(t *testing.T) {
	t.Parallel()

	token := tests.FakeAttendeeJWT(t, uuid.New().String())
	client := tests.NewTrainerHTTPClient(t, token)

	hour := tests.RelativeDate(11, 13)

	code := client.MakeHourAvailable(t, hour)
	require.Equal(t, http.StatusUnauthorized, code)
}

func startService() bool {
	app := configs.NewApplication(context.Background())

	trainerHTTPAddr := os.Getenv("TRAINER_HTTP_ADDR")
	go server.RunHTTPServerOnAddr(trainerHTTPAddr, func(router chi.Router) http.Handler {
		return http2.HandlerFromMux(http2.NewHttpServer(app), router)
	})

	trainerGrpcAddr := os.Getenv("TRAINER_GRPC_ADDR")
	go server.RunGRPCServerOnAddr(trainerGrpcAddr, func(server *grpc.Server) {
		svc := presentation2.NewGrpcServer(app)
		trainer.RegisterTrainerServiceServer(server, svc)
	})

	ok := tests.WaitForPort(trainerHTTPAddr)
	if !ok {
		log.Println("Timed out waiting for trainer HTTP to come up")
		return false
	}

	ok = tests.WaitForPort(trainerGrpcAddr)
	if !ok {
		log.Println("Timed out waiting for trainer gRPC to come up")
	}

	return ok
}

func TestMain(m *testing.M) {
	if !startService() {
		log.Println("Timed out waiting for trainings HTTP to come up")
		os.Exit(1)
	}

	os.Exit(m.Run())
}
