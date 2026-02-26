package container

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/jamesphm04/splose-clone-be/internal/config"
	"github.com/jamesphm04/splose-clone-be/internal/database"
	"github.com/jamesphm04/splose-clone-be/internal/handlers"
	"github.com/jamesphm04/splose-clone-be/internal/repositories"
	"github.com/jamesphm04/splose-clone-be/internal/services"
	"github.com/jamesphm04/splose-clone-be/pkg/auth"
	"github.com/jamesphm04/splose-clone-be/pkg/storage"
)

type Container struct {
	// Global
	cfg *config.Config
	log *zap.Logger
	db  *gorm.DB

	// Infrastructure
	JWTManager *auth.Manager
	S3Client   *storage.Client

	// Repositories
	UserRepo    repositories.UserRepository
	PatientRepo repositories.PatientRepository
	NoteRepo    repositories.NoteRepository

	// Services
	UserSvc     *services.UserService
	PatientSvc  *services.PatientService
	NoteService *services.NoteService

	// Handlers
	AuthHandler    *handlers.AuthHandler
	UserHandler    *handlers.UserHandler
	PatientHandler *handlers.PatientHandler
	NoteHandler    *handlers.NoteHandler
}

// New wires the fill dependency graph and returns a ready Container
// Any error here is fatal - the caller should log and exit
func New(cfg *config.Config, log *zap.Logger) (*Container, error) {
	c := &Container{
		cfg: cfg,
		log: log,
	}

	if err := c.buildInfrastructure(); err != nil {
		return nil, err
	}
	c.buildRepositories()
	c.buildServices()
	c.buildHandlers()

	return c, nil
}

func (c *Container) buildInfrastructure() error {
	// Database
	db, err := database.Connect(c.cfg.DB, c.cfg.AppEnv, c.log)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	if err := database.Migrate(db, c.log); err != nil {
		return fmt.Errorf("migration: %w", err)
	}
	c.db = db

	// JWT
	c.JWTManager = auth.NewManager(
		c.cfg.JWT.Secret,
		c.cfg.JWT.AccessTTL,
		c.cfg.JWT.RefreshTTL,
	)
	// S3
	s3, err := storage.NewClient(
		context.Background(),
		c.cfg.AWS.Region,
		c.cfg.AWS.AccessKeyID,
		c.cfg.AWS.SecretAccessKey,
		c.cfg.AWS.S3Bucket,
		c.cfg.AWS.S3Endpoint,
		c.log,
	)
	if err != nil {
		return fmt.Errorf("s3: %w", err)
	}
	c.S3Client = s3

	return nil
}

func (c *Container) buildRepositories() {
	c.UserRepo = repositories.NewUserRepository(c.db, c.log)
	c.PatientRepo = repositories.NewPatientRepository(c.db, c.log)
	c.NoteRepo = repositories.NewNoteRepository(c.db, c.log)
}

func (c *Container) buildServices() error {
	c.UserSvc = services.NewUserService(c.UserRepo, c.JWTManager, c.cfg.Security.BcryptCost, c.log)
	c.PatientSvc = services.NewPatientService(c.PatientRepo, c.log)
	c.NoteService = services.NewNoteService(c.NoteRepo, c.log)
	return nil
}

func (c *Container) buildHandlers() error {
	c.AuthHandler = handlers.NewAuthHandler(c.UserSvc, c.log)
	c.UserHandler = handlers.NewUserHandler(c.UserSvc, c.log)
	c.PatientHandler = handlers.NewPatientHandler(c.PatientSvc, c.log)
	c.NoteHandler = handlers.NewNoteHandler(c.NoteService, c.log)
	return nil
}

// Router returns a fully configured *gin.Engine by assembling the handler deps.
func (c *Container) Router() interface{} {
	return handlers.SetupRouter(handlers.RouterDeps{
		Log:            c.log,
		JWTManager:     c.JWTManager,
		AuthHandler:    c.AuthHandler,
		UserHandler:    c.UserHandler,
		PatientHandler: c.PatientHandler,
		NoteHandler:    c.NoteHandler,
	})
}

func (c *Container) Close() error {
	return nil
}
