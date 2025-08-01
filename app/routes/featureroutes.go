package routes

import (
    "RAAS/core/config"
    "RAAS/core/middlewares"
    // "RAAS/internal/handlers/features/generation"
    "RAAS/internal/handlers"
    "RAAS/internal/handlers/features/appuser"
    "RAAS/internal/handlers/features/generation"
    "RAAS/internal/handlers/features/jobs"
    "RAAS/internal/handlers/features/exam"
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


    newDashboardHandler := appuser.NewDashboardHandler()
    newDashboardRoutes := r.Group("/b1/new-dashboard",auth)
    {
        newDashboardRoutes.GET("/mini-status", newDashboardHandler.GetStatus)
        newDashboardRoutes.GET("/mini-info", newDashboardHandler.GetInfoBlock)
        newDashboardRoutes.GET("/mini-profile", newDashboardHandler.GetProfile)
        newDashboardRoutes.GET("/mini-checklist", newDashboardHandler.GetChecklist)
        newDashboardRoutes.GET("/mini-jobs", newDashboardHandler.GetMiniNewJobs)
        newDashboardRoutes.GET("/mini-tests", newDashboardHandler.GetMiniTestSummary)
    }

    savedJobsHandler := appuser.NewSavedJobsHandler()
    r.Group("/b1/saved-jobs", auth, paginate).
        POST("", savedJobsHandler.SaveJob).
        GET("", savedJobsHandler.GetSavedJobs).
        DELETE("/:job_id",savedJobsHandler.DeleteSavedJob)

    selectedJobsHandler := appuser.NewSelectedJobHandler()
    r.Group("/b1/api/selected-jobs", auth, paginate).
        GET("", selectedJobsHandler.GetSelectedJobApplications)
    // myApplicationsHandler := appuser.NewMyApplicationsHandler()
    // r.Group("/b1/api/my-applications", auth, paginate).
    //  GET("", myApplicationsHandler.GetMyApplications)

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
    GET("/download-all/:job_id", applicationTrackerHandler.GetCVAndCL)
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

    //PAYMENT Routes
    paymentHandler := payment.NewPaymentHandler()
    payRoutes := r.Group("/b1/payment", auth)
    {
        payRoutes.POST("/checkout", paymentHandler.CreateCheckout)
        payRoutes.GET("/billing/portal", paymentHandler.CustomerPortal)
    }
    r.POST("/b1/payment/webhook", paymentHandler.Webhook)

    // // JOB METADATA Routes
    matchHandler := jobs.NewMatchScoreHandler()
    r.Group("/b1/matchscores", auth).
    GET("", matchHandler.GetMatchScores)

     // // === EXAMS===
    //EXAMS Routes
    examGroup := r.Group("/b2/exam")

    questionsHandler := exam.NewQuestionsHandler()
    questionsRoutes := examGroup.Group("/questions",auth,paginate)
    {
        questionsRoutes.POST("",questionsHandler.PostQuestion)
        questionsRoutes.GET("", questionsHandler.GetQuestions)
        questionsRoutes.PUT("/:question_id", questionsHandler.UpdateQuestion)
        questionsRoutes.PATCH("/:question_id", questionsHandler.PatchQuestion)
        questionsRoutes.DELETE("/:question_id", questionsHandler.DeleteQuestion)
    }

    examPortalHandler :=exam.NewExamPortalHandler()
    examPortalRoutes := examGroup.Group("/portal",auth)
    {
        examPortalRoutes.POST("/input",examPortalHandler.GenerateRandomExam)
        examPortalRoutes.POST("/results/submit",examPortalHandler.ProcessExamResults)
        examPortalRoutes.GET("/results/list",examPortalHandler.GetFilteredExamResults)
        examPortalRoutes.GET("/results/recent",examPortalHandler.GetRecentExamResults)
    }


    // SETTINGS Routes
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
        settingsRoutes.GET("/cancel/active-plan",settingsHandler.PortalCancelSubscription)
        settingsRoutes.POST("/givefeedback", settingsHandler.RequestFeedbackEmail)
        settingsRoutes.POST("/change-email-request", settingsHandler.SendEmailChangeRequest)
        settingsRoutes.POST("/change-job-title-request", settingsHandler.SendJobTitleChangeRequest)
    }
    r.DELETE("b1/user/account", auth, settings.DeleteMyAccountHandler)
}