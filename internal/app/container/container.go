package container

import (
	"context"
	"fmt"
	"go_home/config"
	authservice "go_home/internal/domain/services/auth"
	profileservice "go_home/internal/domain/services/profile"
	authhandler "go_home/internal/endpoint/controller/http/auth"
	profilehandler "go_home/internal/endpoint/controller/http/profile"
	authrepository "go_home/internal/infrastructure/repository/auth"
	profilerepository "go_home/internal/infrastructure/repository/profile"
	httpmiddleware "go_home/internal/middleware"
	"go_home/pkg/database"
	httpserver "go_home/pkg/httpserver/server"
	"go_home/pkg/jwt"
	"go_home/pkg/logger"

	"go_home/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Container struct {
	cfg        *config.Config
	logger     *logger.Logger
	pool       *database.ConnectionPool
	server     *httpserver.HTTPServer
	jwtService *jwt.JWTService

	authRepository *authrepository.AuthRepository
	authService    *authservice.AuthService
	authHandler    *authhandler.AuthHTTPHandler

	profileRepository *profilerepository.ProfileRepository
	profileService    *profileservice.ProfileService
	profileHandler    *profilehandler.ProfileHTTPHandler
}

func New(cfg *config.Config) *Container {
	return &Container{cfg: cfg}
}

func (c *Container) Start(ctx context.Context) error {
	if err := c.initLogger(); err != nil {
		return fmt.Errorf("init logger: %w", err)
	}

	if err := c.initDatabase(ctx); err != nil {
		return fmt.Errorf("init database: %w", err)
	}

	if err := c.initHTTPServer(); err != nil {
		return fmt.Errorf("init HTTP server: %w", err)
	}

	return c.server.Run(ctx)
}

func (c *Container) Stop() {
	if c.pool != nil {
		c.pool.Close()
	}

	if c.logger != nil {
		c.logger.Close()
	}
}

func (c *Container) initLogger() error {
	log, err := logger.NewLogger(logger.Config{
		Level:  c.cfg.Logger.Level,
		Folder: c.cfg.Logger.Folder,
	})
	if err != nil {
		return err
	}

	c.logger = log

	return nil
}

func (c *Container) initDatabase(ctx context.Context) error {
	pool, err := database.NewConnectionPool(ctx, database.Config{
		Host:     c.cfg.Postgres.Host,
		Port:     c.cfg.Postgres.Port,
		User:     c.cfg.Postgres.User,
		Password: c.cfg.Postgres.Password,
		DB:       c.cfg.Postgres.DB,
		SSLMode:  c.cfg.Postgres.SSLMode,
		Timeout:  c.cfg.Postgres.Timeout,
	})
	if err != nil {
		return err
	}
	c.pool = pool
	return nil
}

func (c *Container) initHTTPServer() error {
	server, err := c.getHTTPServer()
	if err != nil {
		return fmt.Errorf("init HTTP server: %w", err)
	}
	c.server = server

	return nil

}
func (c *Container) getJWTService() (*jwt.JWTService, error) {
	if c.jwtService == nil {
		service, err := jwt.NewJWTService(jwt.Config{
			Secret:            c.cfg.JWT.Secret,
			AccessExpiration:  c.cfg.JWT.AccessExpiration,
			RefreshExpiration: c.cfg.JWT.RefreshExpiration,
		})
		if err != nil {
			return nil, fmt.Errorf("create JWT service: %w", err)
		}

		c.jwtService = service
	}

	return c.jwtService, nil
}

func (c *Container) getLogger() *logger.Logger {
	if c.logger == nil {
		log, err := logger.NewLogger(logger.Config{
			Level:  c.cfg.Logger.Level,
			Folder: c.cfg.Logger.Folder,
		})
		if err != nil {
			panic(fmt.Errorf("failed to create logger: %w", err))
		}
		c.logger = log
	}
	return c.logger
}

func (c *Container) getPool() *database.ConnectionPool {
	if c.pool == nil {
		pool, err := database.NewConnectionPool(context.Background(), database.Config{
			Host:     c.cfg.Postgres.Host,
			Port:     c.cfg.Postgres.Port,
			User:     c.cfg.Postgres.User,
			Password: c.cfg.Postgres.Password,
			DB:       c.cfg.Postgres.DB,
			SSLMode:  c.cfg.Postgres.SSLMode,
			Timeout:  c.cfg.Postgres.Timeout,
		})
		if err != nil {
			panic(fmt.Errorf("failed to create database pool: %w", err))
		}
		c.pool = pool
	}
	return c.pool
}

func (c *Container) getHTTPServer() (*httpserver.HTTPServer, error) {
	if c.server != nil {
		return c.server, nil
	}

	log := c.getLogger()

	jwtService, err := c.getJWTService()
	if err != nil {
		return nil, fmt.Errorf("get JWT service: %w", err)
	}

	server := httpserver.NewHTTPServer(
		httpserver.Config{
			Addr:            c.cfg.HTTP.Addr,
			ShutdownTimeout: c.cfg.HTTP.ShutdownTimeout,
		},
		log,
	)

	router := server.Router()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	c.registerSwagger(router)

	apiV1 := router.Group("/api/v1")

	c.getAuthHandler().RegisterRoutes(apiV1)

	protected := apiV1.Group("")
	protected.Use(httpmiddleware.NewJWTAuth(jwtService).Handle())

	c.getProfileHandler().RegisterRoutes(protected)

	c.server = server

	return c.server, nil
}

func (c *Container) getAuthRepository() *authrepository.AuthRepository {
	if c.authRepository == nil {
		c.authRepository = authrepository.NewAuthRepository(
			c.getPool(),
		)
	}

	return c.authRepository
}

func (c *Container) getAuthService() *authservice.AuthService {
	if c.authService == nil {
		jwtService, err := c.getJWTService()
		if err != nil {
			panic(fmt.Errorf("get JWT service: %w", err))
		}

		c.authService = authservice.NewAuthService(
			c.getAuthRepository(),
			jwtService,
		)
	}

	return c.authService
}

func (c *Container) getAuthHandler() *authhandler.AuthHTTPHandler {
	if c.authHandler == nil {
		c.authHandler = authhandler.NewAuthHTTPHandler(
			c.getAuthService(),
		)
	}

	return c.authHandler
}

func (c *Container) getProfileHandler() *profilehandler.ProfileHTTPHandler {
	if c.profileHandler == nil {
		c.profileHandler = profilehandler.NewProfileHTTPHandler(
			c.getProfileService(),
		)
	}

	return c.profileHandler
}

func (c *Container) getProfileService() *profileservice.ProfileService {
	if c.profileService == nil {
		c.profileService = profileservice.NewProfileService(
			c.getProfileRepository(),
		)
	}

	return c.profileService
}

func (c *Container) getProfileRepository() *profilerepository.ProfileRepository {
	if c.profileRepository == nil {
		c.profileRepository = profilerepository.NewProfileRepository(c.getPool())
	}

	return c.profileRepository
}

func (c *Container) registerSwagger(router gin.IRoutes) {
	if docs.SwaggerInfo.ReadDoc() == "" {
		c.getLogger().Warn("Swagger documentation is empty, skipping registration")

		return
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
