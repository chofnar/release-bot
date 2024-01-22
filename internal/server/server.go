package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"

	"github.com/chofnar/release-bot/internal/database"
	databaseLoader "github.com/chofnar/release-bot/internal/database/loader"
	"github.com/chofnar/release-bot/internal/errors"
	"github.com/chofnar/release-bot/internal/server/behaviors"
	botConfig "github.com/chofnar/release-bot/internal/server/config"
	"github.com/chofnar/release-bot/internal/server/consts"
	"github.com/chofnar/release-bot/internal/server/logger"
	myHandlers "github.com/chofnar/release-bot/internal/server/telegohandlers"
	th "github.com/mymmrac/telego/telegohandler"

	// "github.com/fasthttp/router"
	"github.com/hasura/go-graphql-client"
	"github.com/mymmrac/telego"
	"go.uber.org/zap"
	"golang.org/x/oauth2"

	mapset "github.com/deckarep/golang-set/v2"
)

func Initialize(logger zap.SugaredLogger) (*botConfig.BotConfig, database.Database) {
	conf := botConfig.LoadBotConfig()
	db := databaseLoader.GetDatabase(logger)
	return conf, db
}

func Start() {
	logger := logger.New()
	botConf, db := Initialize(*logger)

	bot, err := telego.NewBot(botConf.TelegramToken, telego.WithLogger(logger))
	if err != nil {
		logger.Error(err)

		panic(err)
	}

	// Create Github GraphQL token
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: botConf.GithubGQLToken},
	)

	httpClient := oauth2.NewClient(context.Background(), src)
	githubGQLClient := graphql.NewClient("https://api.github.com/graphql", httpClient)

	linkRegex, _ := regexp.Compile("(?:https://)github.com[:/](.*)[:/](.*)")
	directRegex, _ := regexp.Compile("(.*)[/](.*)")

	behaviorHandler := behaviors.BehaviorHandler{
		Bot:         bot,
		LinkRegex:   linkRegex,
		DirectRegex: directRegex,
		GQLClient:   githubGQLClient,
		DB:          db,
	}

	awaitingAddRepo := map[int64]struct{}{}

	handler := myHandlers.Handler{
		BehaviorHandler: behaviorHandler,
		Logger:          *logger,
		AwaitingAddRepo: awaitingAddRepo,
	}

	if botConf.ResetWebhookUrl != "" {
		err = bot.SetWebhook(&telego.SetWebhookParams{
			URL: botConf.WebhookSite + "/bot/" + botConf.TelegramToken,
		})
		if err != nil {
			panic(err)
		}
	}

	mux := http.NewServeMux()
	stats := StatsPath{}
	mux.Handle("/stats", stats.ServeHTTP(&behaviorHandler, *logger))
	up := UpdatePath{}
	mux.Handle("/updateRepos", up.UpdateRepos(&behaviorHandler, *logger))

	updates, err := bot.UpdatesViaWebhook("/bot/"+bot.Token(), telego.WithWebhookServer(telego.HTTPWebhookServer{
		Logger:   logger,
		Server:   &http.Server{},
		ServeMux: mux,
	}))
	if err != nil {
		panic(err)
	}

	botHandler, err := th.NewBotHandler(bot, updates)
	if err != nil {
		panic(err)
	}

	// register handlers
	botHandler.Handle(handler.Start(), th.CommandEqual("start"))
	botHandler.Handle(handler.About(), th.CommandEqual("about"))
	botHandler.Handle(handler.UnknownOrSent(), th.AnyMessageWithText())

	// Callback queries
	botHandler.HandleCallbackQuery(handler.SeeAll(), th.CallbackDataEqual(consts.SeeAllCallback))
	botHandler.HandleCallbackQuery(handler.Add(), th.CallbackDataEqual(consts.AddCallback))
	botHandler.HandleCallbackQuery(handler.Menu(), th.CallbackDataEqual(consts.MenuCallback))
	botHandler.HandleCallbackQuery(handler.AnyCallbackRouter(), th.AnyCallbackQuery())

	// start listening

	go func() {
		botHandler.Start()
	}()

	go func() {
		err = bot.StartWebhook("0.0.0.0:" + botConf.Port)
		if err != nil {
			panic(err)
		}
	}()

	defer func() {
		_ = bot.StopWebhook()
		botHandler.Stop()
	}()

	ctx := context.Background()
	nctx, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill)

	defer stop()
	<-nctx.Done()
}

type StatsPath struct{}

func (hp StatsPath) ServeHTTP(behaviorHandler *behaviors.BehaviorHandler, logger zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer logger.Sync()
		b, err := io.ReadAll(r.Body)
		if err != nil {
			msg := "could not read request body"
			logger.Error(msg)
			_, writeErr := w.Write([]byte(msg))
			if writeErr != nil {
				logger.Error(writeErr)
			}
			return
		}

		bodyStr := string(b)

		if bodyStr != os.Getenv("SUPER_SECRET_TOKEN") {
			logger.Error(errors.ErrUpdateIncorrectToken)
			_, writeErr := w.Write([]byte(errors.ErrUpdateIncorrectToken.Error()))
			if writeErr != nil {
				logger.Error(writeErr)
			}
			return
		}

		repos, err := behaviorHandler.DB.AllRepos()
		if err != nil {
			msg := "Something went wrong querying the database: " + err.Error()
			logger.Error([]byte(msg))
			_, writeErr := w.Write([]byte(msg))
			if writeErr != nil {
				logger.Error(writeErr)
			}
			return
		}

		uniqueUsersSet, uniqueReposSet := mapset.NewSet[string](), mapset.NewSet[string]()

		for _, repo := range repos {
			uniqueUsersSet.Add(repo.ChatID)
			uniqueReposSet.Add(repo.RepoID)
		}

		_, writeErr := w.Write([]byte("Currently serving " + strconv.Itoa(uniqueUsersSet.Cardinality()) + " users, watching " + strconv.Itoa(uniqueReposSet.Cardinality()) + " unique repos"))
		if writeErr != nil {
			logger.Error(writeErr)
		}
	}
}

type UpdatePath struct{}

func (up UpdatePath) UpdateRepos(behaviorHandler *behaviors.BehaviorHandler, logger zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer logger.Sync()
		b, err := io.ReadAll(r.Body)
		if err != nil {
			msg := "could not read request body"
			logger.Error(msg)
			_, writeErr := w.Write([]byte(msg))
			if writeErr != nil {
				logger.Error(writeErr)
			}
			return
		}

		bodyStr := string(b)

		if bodyStr != os.Getenv("SUPER_SECRET_TOKEN") {
			logger.Error(errors.ErrUpdateIncorrectToken)
			_, writeErr := w.Write([]byte(errors.ErrUpdateIncorrectToken.Error()))
			if writeErr != nil {
				logger.Error(writeErr)
			}
			return
		}

		failedRepoErrors := behaviorHandler.UpdateRepos(logger)
		marshaledErrors, err := json.Marshal(failedRepoErrors)
		if err != nil {
			logger.Error(err)
			_, writeErr := w.Write([]byte(err.Error()))
			if writeErr != nil {
				logger.Error(writeErr)
			}
			return
		}

		if len(failedRepoErrors) != 0 {
			logger.Error(failedRepoErrors)
			_, writeErr := w.Write(marshaledErrors)
			if writeErr != nil {
				logger.Error(writeErr)
			}
			return
		}

		msg := "Repos updated successfully with no funky business"
		logger.Info(msg)
		_, writeErr := w.Write([]byte(msg))
		if writeErr != nil {
			logger.Error(writeErr)
		}
	}
}
