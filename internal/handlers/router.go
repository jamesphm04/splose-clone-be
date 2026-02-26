// Package handlers also contains the router setup that wires all handlers
// together with their middleware.
package handlers

import (
	"go.uber.org/zap"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/jamesphm04/splose-clone-be/internal/middleware"
	"github.com/jamesphm04/splose-clone-be/pkg/auth"
)

// RouterDeps bundles every dependency needed to build the HTTP router.
// It is populated by the DI Container and passed to SetupRouter.
type RouterDeps struct {
	Log            *zap.Logger // root logger – middleware uses named children
	JWTManager     *auth.Manager
	AuthHandler    *AuthHandler
	UserHandler    *UserHandler
	PatientHandler *PatientHandler
	NoteHandler    *NoteHandler
	// ConvHandler    *ConversationHandler
	// PromptHandler  *PromptHandler
	// AttachHandler  *AttachmentHandler
}

// SetupRoter builds and returns a configured *gin.Engine
func SetupRouter(deps RouterDeps) *gin.Engine {
	r := gin.New()

	// Global middleware (order matters)
	r.Use(middleware.Recovery(deps.Log))      // catch panics first
	r.Use(middleware.RequestLogger(deps.Log)) // then log all requests
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Restrict in production.
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Endpoints
	v1 := r.Group("/api/v1")

	// Public
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/register", deps.AuthHandler.Register)
		authGroup.POST("/login", deps.AuthHandler.Login)
		authGroup.POST("/refresh", deps.AuthHandler.Refresh)
	}

	// Protected (JWT required)
	protected := v1.Group("")
	protected.Use(middleware.Authenticate(deps.JWTManager))
	{
		// User endpoints
		users := protected.Group("/users")
		{
			users.GET("/me", deps.UserHandler.GetMe)
			users.PATCH("/:id", deps.UserHandler.Update)
			users.DELETE("/:id", deps.UserHandler.Delete)
			users.GET("", middleware.RequireRole("admin"), deps.UserHandler.List)
		}
		// Patient endpoints
		patients := protected.Group("/patients")
		{
			patients.POST("", deps.PatientHandler.Create)
			patients.GET("/:id", deps.PatientHandler.GetByID)
			patients.GET("", deps.PatientHandler.List)
			patients.PATCH("/:id", deps.PatientHandler.Update)
		}

		// Progress note endpoints
		notes := protected.Group("/notes")
		{
			notes.POST("", deps.NoteHandler.Create)
			notes.GET("", deps.NoteHandler.List)
			notes.GET("/patient/:patientID", deps.NoteHandler.ListByPatientID)
			notes.GET("/:id", deps.NoteHandler.GetByID)
			notes.PATCH("/:id", deps.NoteHandler.Update)
		}
	}

	return r
}
