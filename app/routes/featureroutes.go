package routes

import (
	"RAAS/core/config"
	"RAAS/core/middlewares"

	// "RAAS/internal/handlers/features/generation"
	"RAAS/internal/handlers"
	"RAAS/internal/handlers/features/appuser"
	"RAAS/internal/handlers/features/generation"
	"RAAS/internal/handlers/features/jobs"
	"RAAS/internal/handlers/features/payment"
	"RAAS/internal/handlers/features/settings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupFeatureRoutes(r *gin.Engine, client *mongo.Client, cfg *config.Config) {
	// Inject MongoDB into context
	r.Use(middleware.InjectDB(client))

	// Auth Middleware + Pagination helpers
	auth := middleware.AuthMiddleware()
	paginate := middleware.PaginationMiddleware

	// === USER ===

	seekerHandler := settings.NewSeekerHandler()
	r.Group("/b1/jobprofile", auth).
	GET("",seekerHandler.GetSeekerProfile)

	seekerProfileHandler := appuser.NewSeekerProfileHandler()
	dashBoardRoute := r.Group("/b1/dashboard", auth)
	dashBoardRoute.GET("", seekerProfileHandler.GetDashboard)


	// savedJobsHandler := appuser.NewSavedJobsHandler()
	// r.Group("/b1/saved-jobs", auth, paginate).
	// 	POST("", savedJobsHandler.SaveJob).
	// 	GET("", savedJobsHandler.GetSavedJobs)

	selectedJobsHandler := appuser.NewSelectedJobHandler()
	r.Group("/b1/api/selected-jobs", auth, paginate).
		GET("", selectedJobsHandler.GetSelectedJobApplications)

	// myApplicationsHandler := appuser.NewMyApplicationsHandler()
	// r.Group("/b1/api/my-applications", auth, paginate).
	// 	GET("", myApplicationsHandler.GetMyApplications)


	// === JOBS ===

	r.Group("/b1/api/jobs", auth, paginate).
		GET("", jobs.JobRetrievalHandler)

	linkProviderHandler := jobs.NewLinkProviderHandler()
	r.Group("/b1/provide-link", auth).
		POST("", linkProviderHandler.PostAndGetLink)

	applicationTrackerHandler := appuser.NewApplicationTrackerHandler()
	r.Group("/b1/api/application-tracker",auth,paginate).
	GET("", applicationTrackerHandler.GetApplicationTracker).
	PUT("/:job_id/status", applicationTrackerHandler.UpdateApplicationStatus).
	POST("/download-all", applicationTrackerHandler.GetCVAndCL)

	r.GET("/b1/test/academics/dates", handlers.TestAcademicDatesHandler)


	// // === GENERATION ===

    // Group route under /b1/generate-cover-letter
	coverLetterHandler := generation.NewInternalCoverLetterHandler()
    generateCLRoute := r.Group("/b1/internal/generate-cover-letter", auth)
    generateCLRoute.POST("", coverLetterHandler.PostCoverLetter)
	generateCLRoute.PUT("",coverLetterHandler.PutCoverLetter)
	generateCLRoute.GET("",coverLetterHandler.GetCoverLetter)


	resumeHandler := generation.NewInternalCVHandler()

	resumeRoute := r.Group("/b1/internal/generate-resume", auth)
	resumeRoute.POST("", resumeHandler.PostCV)
	// resumeRoute.PUT("", resumeHandler.PutCV)
	resumeRoute.GET("", resumeHandler.GetCV)
	resumeRoute.PUT("",resumeHandler.PutCV)


	extGenHandler := generation.NewExternalJobCVNCLGenerator()
	route := r.Group("/b1/external/generate", auth)
	route.POST("", extGenHandler.PostExternalCVNCL)
	route.GET("", extGenHandler.GetExternalCVNCL)
	route.PUT("/cv",extGenHandler.PutCV)
	route.PUT("/cl",extGenHandler.PutCoverLetter)

	
	//PAYMENT routes

	paymentHandler := payment.NewPaymentHandler()
	payRoutes := r.Group("/b1/payment", auth)
	{
		payRoutes.POST("/checkout", paymentHandler.CreateCheckout)
		payRoutes.GET("/billing/portal", paymentHandler.CustomerPortal)
	}
	r.POST("/b1/payment/webhook", paymentHandler.Webhook)

	// // JOB METADATA routes

	matchHandler := jobs.NewMatchScoreHandler()
	r.Group("/b1/matchscores", auth).
    GET("", matchHandler.GetMatchScores)

	// SETTINGS routes
	settingsHandler := settings.NewSettingsHandler()
	settingsRoutes := r.Group("/b1/settings", auth)
	{
		settingsRoutes.GET("/general", settingsHandler.GetGeneralSettings)
		settingsRoutes.POST("/change-password",settings.RequestPasswordChangeHandler)
		settingsRoutes.GET("/getpreferences", settingsHandler.GetPreferences)
		settingsRoutes.PUT("/editpreferences", settingsHandler.UpdatePreferences)
		settingsRoutes.GET("/getnotification", settingsHandler.GetNotificationSettings)
		settingsRoutes.PUT("/editnotification", settingsHandler.UpdateNotificationSettings)
		settingsRoutes.GET("/view/billing", settingsHandler.GetBillingInfo)
		settingsRoutes.GET("/billing/payment-method",settingsHandler.PortalPaymentMethod)
		settingsRoutes.GET("/explore-plans",settingsHandler.GetExplorePlans)

		settingsRoutes.POST("/givefeedback", settingsHandler.RequestFeedbackEmail)

		settingsRoutes.POST("/change-email-request", settingsHandler.SendEmailChangeRequest)
		settingsRoutes.POST("/change-job-title-request", settingsHandler.SendJobTitleChangeRequest)
	}
	r.DELETE("b1/user/account", auth, settings.DeleteMyAccountHandler)

}
