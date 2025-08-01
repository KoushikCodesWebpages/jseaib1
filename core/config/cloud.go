package config

import (
    "fmt"
    "github.com/spf13/viper"
)

type CloudConfig struct {
    AzureStorageAccount        string
    AzureStorageKey            string
    AzureCertificatesContainer string
    AzureLanguagesContainer    string
    AzureCoverlettersContainer string
    AzureResumesContainer      string
    AzureProfilePicContainer   string

    GoogleClientId             string
    GoogleClientSecret         string
    GoogleRedirectURL          string

    EmailBackend               string
    EmailHost                  string
    EmailPort                  int
    EmailUseTLS                bool
    EmailHostUser              string
    EmailHostPassword          string
    DefaultFromEmail           string
    StaticURL                  string

    MongoDBUri                 string
    MongoDBName                string

    CL_Url                     string
    CV_Url                     string

    DataExtractionAPI          string

    GEN_API_KEY                string

    BasicPlanMonthly           string
    BasicPlanQuarterly          string
    AdvancedPlanMonthly        string
    AdvancedPlanQuarterly       string
    PremiumPlanMonthly         string
    PremiumPlanQuarterly        string


    // DevBasicPlanMonthly         string
    // DevStripeSecretKey          string
    // DevStripePublishableKey       string
    // DevStripeWebHookKey         string

    StripeSecretKey            string
    StripePublishableKey       string
    StripeWebHookKey           string

    TaxRateID                   string
}

func LoadCloudConfig() (*CloudConfig, error) {
    dbConfig := &CloudConfig{
        AzureStorageAccount:        viper.GetString("AZURE_STORAGE_ACCOUNT"),
        AzureStorageKey:             viper.GetString("AZURE_STORAGE_KEY"),
        AzureCertificatesContainer: viper.GetString("AZURE_CERTIFICATES_CONTAINER"),
        AzureLanguagesContainer:    viper.GetString("AZURE_LANGUAGES_CONTAINER"),
        AzureCoverlettersContainer: viper.GetString("AZURE_COVERLETTERS_CONTAINER"),
        AzureResumesContainer:      viper.GetString("AZURE_RESUMES_CONTAINER"),
        AzureProfilePicContainer:   viper.GetString("AZURE_PROFILEPICS_CONTAINER"),

        GoogleClientId:             viper.GetString("GOOGLE_CLIENT_ID"),
        GoogleClientSecret:         viper.GetString("GOOGLE_CLIENT_SECRET"),
        GoogleRedirectURL:          viper.GetString("GOOGLE_REDIRECT_URL"),

        EmailBackend:               viper.GetString("EMAIL_BACKEND"),
        EmailHost:                  viper.GetString("EMAIL_HOST"),
        EmailPort:                  viper.GetInt("EMAIL_PORT"),
        EmailUseTLS:                viper.GetBool("EMAIL_USE_TLS"),
        EmailHostUser:              viper.GetString("EMAIL_HOST_USER"),
        EmailHostPassword:          viper.GetString("EMAIL_HOST_PASSWORD"),
        DefaultFromEmail:           viper.GetString("DEFAULT_FROM_EMAIL"),
        StaticURL:                  viper.GetString("STATIC_URL"),

        MongoDBUri:                 viper.GetString("MONGO_DB_URI"),
        MongoDBName:                viper.GetString("MONGO_DB_NAME"),

        CL_Url:                     viper.GetString("COVER_LETTER_API_URL"),
        CV_Url:                     viper.GetString("CV_RESUME_API_URL"),
        GEN_API_KEY:                viper.GetString("COVER_CV_API_KEY"),

        DataExtractionAPI:          viper.GetString("DATA_EXTRACTION_API"),    

        // DevBasicPlanMonthly:        viper.GetString("DEV_PRICE_BASIC_MONTH"),
        // DevStripeSecretKey:         viper.GetString("DEV_STRIPE_SECRET_KEY"),
        // DevStripePublishableKey:       viper.GetString("DEV_STRIPE_PUBLISHABLE_KEY"),
        // DevStripeWebHookKey:           viper.GetString("DEV_STRIPE_WEBHOOK_SECRET"),


        
        BasicPlanMonthly:           viper.GetString("PRICE_BASIC_MONTH"),
        BasicPlanQuarterly:          viper.GetString("PRICE_BASIC_QUAD"),
        AdvancedPlanMonthly:        viper.GetString("PRICE_ADVANCED_MONTH"),
        AdvancedPlanQuarterly:       viper.GetString("PRICE_ADVANCED_QUAD"),
        PremiumPlanMonthly:         viper.GetString("PRICE_PREMIUM_MONTH"),
        PremiumPlanQuarterly:        viper.GetString("PRICE_PREMIUM_QUAD"),

        StripeSecretKey:            viper.GetString("STRIPE_SECRET_KEY"),
        StripePublishableKey:       viper.GetString("STRIPE_PUBLISHABLE_KEY"),
        StripeWebHookKey:           viper.GetString("STRIPE_WEBHOOK_SECRET"),

        TaxRateID:                  viper.GetString("TAX_RATE_ID"),


    }

    // Validate required fields
    if dbConfig.AzureStorageAccount == "" {
        return nil, fmt.Errorf("AzureStorageAccount is required but not set")
    }
    if dbConfig.GoogleClientId == "" {
        return nil, fmt.Errorf("GoogleClientId is required but not set")
    }

    return dbConfig, nil
}
