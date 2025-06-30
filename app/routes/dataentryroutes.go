package routes

import (
	"RAAS/core/config"
	"RAAS/core/middlewares"

	"RAAS/internal/handlers/features/appuser"
	"RAAS/internal/handlers/preference"


	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	// "go.mongodb.org/mongo-driver/mongo/options"
)

func SetupDataEntryRoutes(r *gin.Engine, client *mongo.Client, cfg *config.Config) {
	r.Use(middleware.InjectDB(client))
	// TIMELINE
	timeline := r.Group("/b1/user/entry-progress/check")
	timeline.Use(middleware.AuthMiddleware()) // Middleware to authenticate JWT

	// Define the route for getting the next entry step
	timeline.GET("", appuser.GetNextEntryStep())

	// PERSONAL INFO routes
	personalInfoHandler := preference.NewPersonalInfoHandler()
	personalInfoRoutes := r.Group("/b1/personal-info")
	personalInfoRoutes.Use(middleware.AuthMiddleware())
	{
		personalInfoRoutes.POST("", personalInfoHandler.CreatePersonalInfo)
		personalInfoRoutes.GET("", personalInfoHandler.GetPersonalInfo)    
	}

	// // PROFESSIONAL SUMMARY routes
	// professionalSummaryHandler := preference.NewProfessionalSummaryHandler()
	// professionalSummaryRoutes := r.Group("/b1/professional-summary")
	// professionalSummaryRoutes.Use(middleware.AuthMiddleware())
	// {
	// 	professionalSummaryRoutes.POST("", professionalSummaryHandler.CreateProfessionalSummary)
	// 	professionalSummaryRoutes.GET("", professionalSummaryHandler.GetProfessionalSummary)
	// 	professionalSummaryRoutes.PUT("", professionalSummaryHandler.UpdateProfessionalSummary)
	// }


	workExperienceHandler := preference.NewWorkExperienceHandler()
	workExperienceRoutes := r.Group("/b1/work-experience")
	workExperienceRoutes.Use(middleware.AuthMiddleware())
	{
		workExperienceRoutes.POST("", workExperienceHandler.CreateWorkExperience)
		workExperienceRoutes.GET("", workExperienceHandler.GetWorkExperience)
		workExperienceRoutes.PUT("/:id", workExperienceHandler.UpdateWorkExperience)
		workExperienceRoutes.DELETE("/:id", workExperienceHandler.DeleteWorkExperience)

	}


	academicsHandler := preference.NewAcademicsHandler()
	academicsRoutes := r.Group("/b1/academics")
	academicsRoutes.Use(middleware.AuthMiddleware())
	{
		academicsRoutes.POST("", academicsHandler.CreateAcademics)
		academicsRoutes.GET("", academicsHandler.GetAcademics)
		academicsRoutes.PUT("/:id",academicsHandler.UpdateAcademics)
		academicsRoutes.DELETE("/:id",academicsHandler.DeleteAcademics)
	}

	pastProjectHandler:= preference.NewPastProjectHandler()
	pastProjectRoutes:=r.Group("/b1/pastprojects")
	pastProjectRoutes.Use(middleware.AuthMiddleware())
	{
		pastProjectRoutes.POST("",pastProjectHandler.CreatePastProject)
		pastProjectRoutes.GET("",pastProjectHandler.GetPastProjects)
		pastProjectRoutes.PUT("/:id",pastProjectHandler.UpdatePastProject)
		pastProjectRoutes.DELETE("/:id",pastProjectHandler.DeletePastProject)

	}


	// CERTIFICATES routes
	certificateHandler := preference.NewCertificateHandler()
	certificateRoutes := r.Group("/b1/certificates")
	certificateRoutes.Use(middleware.AuthMiddleware())
	{
		certificateRoutes.POST("", certificateHandler.CreateCertificate)
		certificateRoutes.GET("", certificateHandler.GetCertificates)
		certificateRoutes.PUT("/:id", certificateHandler.UpdateCertificate)
		certificateRoutes.DELETE("/:id", certificateHandler.DeleteCertificate)
	}

	// LANGUAGES routes	
	languageHandler := preference.NewLanguageHandler()
	languageRoutes := r.Group("/b1/languages")
	languageRoutes.Use(middleware.AuthMiddleware())
	{
		languageRoutes.POST("", languageHandler.CreateLanguage)
		languageRoutes.GET("", languageHandler.GetLanguages)
		languageRoutes.PUT("/:id", languageHandler.UpdateLanguage)
		languageRoutes.DELETE("/:id", languageHandler.DeleteLanguage)
	}

	// JOB TITLES routes
	jobTitleHandler := preference.NewJobTitleHandler()
	jobTitleRoutes := r.Group("/b1/jobtitles")
	jobTitleRoutes.Use(middleware.AuthMiddleware())
	{
		jobTitleRoutes.POST("", jobTitleHandler.CreateJobTitleOnce)
		jobTitleRoutes.GET("", jobTitleHandler.GetJobTitle)
	}
	keySkillsHandler := preference.NewKeySkillsHandler()

	keySkillsRoutes:= r.Group("/b1/keyskills", middleware.AuthMiddleware()) // or your actual auth middleware
	{
		keySkillsRoutes.GET("", keySkillsHandler.GetKeySkills)
		keySkillsRoutes.POST("", keySkillsHandler.SetKeySkills)
	}
	
	//CVNCL FORMAT routes

	cvClHandler := preference.NewFormatHandler()
	cvClRoutes := r.Group("/b1/formats")
	cvClRoutes.Use(middleware.AuthMiddleware())
	{
		cvClRoutes.PUT("/cv", cvClHandler.UpdateCvFormat)
		cvClRoutes.PUT("/cl", cvClHandler.UpdateClFormat)
	}

	//PHOTO UPLOAD routes

	photoHandler := preference.NewPhotoHandler()
	photoRoutes := r.Group("/b1/photo")
	photoRoutes.Use(middleware.AuthMiddleware()) // Only for upload & self
	{
		photoRoutes.POST("/upload", photoHandler.UploadProfilePhoto)
		photoRoutes.GET("", photoHandler.GetProfilePhoto) // current user
	}
	r.GET("/b1/photo/view/:user_id", photoHandler.PublicGetProfilePhoto) 

}
