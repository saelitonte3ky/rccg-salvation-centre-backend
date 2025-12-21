// internal/routes/routes.go
package routes

import (
	"rccg-salvation-centre-backend/internal/handlers"
	"rccg-salvation-centre-backend/internal/middleware"
	"time"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// Root health check (no rate limiting for health checks)
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "RCCG Salvation Centre Backend API is running!",
			"status":  "ok",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// Apply global rate limiting (60 requests per minute)
	api := r.Group("/api")
	api.Use(middleware.RateLimiter())
	{
		api.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Welcome to RCCG Salvation Centre API",
				"version": "1.0",
			})
		})

		// AUTH ROUTES - Strict rate limiting (5 attempts per 15 minutes)
		auth := api.Group("/auth")
		auth.Use(middleware.StrictRateLimiter())
		{
			auth.POST("/login", handlers.Login)
			auth.POST("/logout", handlers.Logout)
			auth.GET("/me", middleware.AuthRequired(), handlers.Me)
		}

		// PUBLIC ROUTES
		api.GET("/sermons", handlers.GetSermons)
		api.GET("/sermons/latest", handlers.GetLatestSermon)
		api.GET("/sermons/search", handlers.SearchSermons)

		// PUBLIC: Service Types
		api.GET("/service-types", handlers.GetServiceTypes)

		// PUBLIC: Testimonies (approved only)
		api.GET("/testimonies", handlers.GetTestimonies)

		// PUBLIC: Submit testimony (moderate rate limit - 10 per hour)
		api.POST("/testimonies",
			middleware.CustomRateLimiter(10, 1*time.Hour),
			handlers.CreateTestimony,
		)

		// PUBLIC: Submit first-timer (moderate rate limit - 5 per hour)
		api.POST("/first-timers",
			middleware.CustomRateLimiter(5, 1*time.Hour),
			handlers.CreateFirstTimer,
		)

		// PUBLIC: Submit prayer request (moderate rate limit - 10 per hour)
		api.POST("/prayer-requests",
			middleware.CustomRateLimiter(10, 1*time.Hour),
			handlers.CreatePrayerRequest,
		)

		// PUBLIC: Special Events & Regular Programs
		api.GET("/special-events", handlers.GetSpecialEvents)
		api.GET("/regular-programs", handlers.GetRegularPrograms)

		// ADMIN PROTECTED ROUTES - Higher rate limits for authenticated users
		admin := api.Group("/admin")
		admin.Use(middleware.AuthRequired())
		admin.Use(middleware.CustomRateLimiter(100, 1*time.Minute)) // 100 requests per minute for admins
		{
			// Sermon Management
			sermons := admin.Group("/sermons")
			{
				sermons.GET("", handlers.AdminGetSermons)
				sermons.POST("", middleware.RequireRoles("superadmin", "media_team"), handlers.CreateSermon)
				sermons.PUT("/:id", middleware.RequireRoles("superadmin", "media_team"), handlers.UpdateSermon)
				sermons.DELETE("/:id", middleware.RequireRoles("superadmin"), handlers.DeleteSermon)
			}

			// Testimonies Management
			testimonies := admin.Group("/testimonies")
			{
				testimonies.GET("", handlers.AdminGetTestimonies)
				testimonies.PUT("/:id", middleware.RequireRoles("superadmin", "secretariat"), handlers.UpdateTestimony)
				testimonies.DELETE("/:id", middleware.RequireRoles("superadmin"), handlers.DeleteTestimony)
			}

			// First-Timers Management
			firstTimers := admin.Group("/first-timers")
			{
				firstTimers.GET("", handlers.AdminGetFirstTimers)
				firstTimers.PUT("/:id", middleware.RequireRoles("superadmin", "visitors_welfare"), handlers.UpdateFirstTimer)
				firstTimers.DELETE("/:id", middleware.RequireRoles("superadmin"), handlers.DeleteFirstTimer)
			}

			// Attendance Management
			attendance := admin.Group("/attendance")
			{
				attendance.GET("", handlers.AdminGetAttendance)
				attendance.POST("", middleware.RequireRoles("superadmin", "secretariat"), handlers.CreateAttendance)
				attendance.PUT("/:id", middleware.RequireRoles("superadmin", "secretariat"), handlers.UpdateAttendance)
				attendance.DELETE("/:id", middleware.RequireRoles("superadmin"), handlers.DeleteAttendance)
			}

			// Prayer Requests Management
			prayerRequests := admin.Group("/prayer-requests")
			{
				prayerRequests.GET("", handlers.AdminGetPrayerRequests)
				prayerRequests.PUT("/:id", middleware.RequireRoles("superadmin", "secretariat"), handlers.UpdatePrayerRequest)
				prayerRequests.DELETE("/:id", middleware.RequireRoles("superadmin", "secretariat"), handlers.DeletePrayerRequest)
			}

			// Dashboard
			admin.GET("/dashboard", handlers.AdminGetDashboard)

			// ADMIN: Special Events Management
			specialEvents := admin.Group("/special-events")
			{
				specialEvents.GET("", handlers.AdminGetSpecialEvents)
				specialEvents.POST("", middleware.RequireRoles("superadmin", "admin"), handlers.CreateSpecialEvent)
				specialEvents.PUT("/:id", middleware.RequireRoles("superadmin", "admin"), handlers.UpdateSpecialEvent)
				specialEvents.DELETE("/:id", middleware.RequireRoles("superadmin", "admin"), handlers.DeleteSpecialEvent)
			}

			// ADMIN: Regular Programs Management
			regularPrograms := admin.Group("/regular-programs")
			{
				regularPrograms.GET("", handlers.AdminGetRegularPrograms)
				regularPrograms.POST("", middleware.RequireRoles("superadmin", "admin"), handlers.CreateRegularProgram)
				regularPrograms.PUT("/:id", middleware.RequireRoles("superadmin", "admin"), handlers.UpdateRegularProgram)
				regularPrograms.DELETE("/:id", middleware.RequireRoles("superadmin", "admin"), handlers.DeleteRegularProgram)
			}
		}
	}
}
