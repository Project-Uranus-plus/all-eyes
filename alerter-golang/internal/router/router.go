package router

import (
	"github.com/Project-Uranus-plus/all-eyes/configs"
	"github.com/Project-Uranus-plus/all-eyes/internal/pkg/core"
	"github.com/Project-Uranus-plus/all-eyes/internal/repository/mysql"
	"github.com/Project-Uranus-plus/all-eyes/pkg/errors"
	"github.com/Project-Uranus-plus/all-eyes/pkg/file"

	"go.uber.org/zap"
)

type resource struct {
	mux          core.Mux
	logger       *zap.Logger
	db           mysql.Repo
}

type Server struct {
	Mux core.Mux
	Db  mysql.Repo
}

func NewHTTPServer(logger *zap.Logger) (*Server, error) {
	if logger == nil {
		return nil, errors.New("logger required")
	}

	r := new(resource)
	r.logger = logger

	openBrowserUri := configs.ProjectDomain + configs.ProjectPort

	_, ok := file.IsExists(configs.ProjectInstallMark)
	if !ok { // 未安装
		openBrowserUri += "/install"
	} else { // 已安装

		// 初始化 DB
		dbRepo, err := mysql.New()
		if err != nil {
			logger.Fatal("new db err", zap.Error(err))
		}
		r.db = dbRepo

	}

	mux, err := core.New(logger,
		core.WithEnableOpenBrowser(openBrowserUri),
		core.WithEnableCors(),
		core.WithEnableRate(),
	)

	if err != nil {
		panic(err)
	}

	r.mux = mux

	// 设置 API 路由
	setApiRouter(r)

	s := new(Server)
	s.Mux = mux
	s.Db = r.db

	return s, nil
}
