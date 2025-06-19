package config

import "github.com/spf13/viper"

type ProjectConfig struct {
	// General Settings
	AuthUserModel                string
	FrontendBaseUrl              string

	// CORS and Auth Settings
	CORSAllowedOrigins           string
	AuthHeaderTypes              string

	// JWT Settings
	JWTSecretKey                 string
	JWTExpirationTime            int
	AccessTokenLifetime          int
	RefreshTokenLifetime         int
	RotateRefreshTokens          bool
	BlacklistAfterRotation       bool

	// Static and Media Settings
	SecretKey                    string
	StaticURL                    string
	MediaURL                     string
	MediaRoot                    string
	StaticRoot                   string

	// REST Framework Settings
	RestAuthClasses              string
	RestPermissionClasses        string
	RestPaginationClass          string
	RestFilterBackends           string
	RestRendererClasses          string
	RestThrottleClasses          string
	RestThrottleRatesAnon        string
	RestThrottleRatesUser        string
}

func LoadProjectConfig() (*ProjectConfig, error) {
	// Load values into the config struct
	ProjectConfig := &ProjectConfig{
		AuthUserModel:              viper.GetString("AUTH_USER_MODEL"),
		FrontendBaseUrl:            viper.GetString("FRONTEND_BASE_URL"),

		CORSAllowedOrigins:         viper.GetString("CORS_ALLOWED_ORIGINS"),
		AuthHeaderTypes:            viper.GetString("AUTH_HEADER_TYPES"),

		JWTSecretKey:               viper.GetString("JWT_SECRET_KEY"),
		JWTExpirationTime:          viper.GetInt("JWT_EXPIRATION_TIME"),
		AccessTokenLifetime:        viper.GetInt("ACCESS_TOKEN_LIFETIME"),
		RefreshTokenLifetime:       viper.GetInt("REFRESH_TOKEN_LIFETIME"),
		RotateRefreshTokens:        viper.GetBool("ROTATE_REFRESH_TOKENS"),
		BlacklistAfterRotation:     viper.GetBool("BLACKLIST_AFTER_ROTATION"),

		SecretKey:                  viper.GetString("SECRET_KEY"),
		StaticURL:                  viper.GetString("STATIC_URL"),
		MediaURL:                   viper.GetString("MEDIA_URL"),
		MediaRoot:                  viper.GetString("MEDIA_ROOT"),
		StaticRoot:                 viper.GetString("STATIC_ROOT"),

		RestAuthClasses:            viper.GetString("REST_FRAMEWORK_DEFAULT_AUTHENTICATION_CLASSES"),
		RestPermissionClasses:      viper.GetString("REST_FRAMEWORK_DEFAULT_PERMISSION_CLASSES"),
		RestPaginationClass:        viper.GetString("REST_FRAMEWORK_DEFAULT_PAGINATION_CLASS"),
		RestFilterBackends:         viper.GetString("REST_FRAMEWORK_DEFAULT_FILTER_BACKENDS"),
		RestRendererClasses:        viper.GetString("REST_FRAMEWORK_DEFAULT_RENDERER_CLASSES"),
		RestThrottleClasses:        viper.GetString("REST_FRAMEWORK_DEFAULT_THROTTLE_CLASSES"),
		RestThrottleRatesAnon:      viper.GetString("REST_FRAMEWORK_DEFAULT_THROTTLE_RATES_ANON"),
		RestThrottleRatesUser:      viper.GetString("REST_FRAMEWORK_DEFAULT_THROTTLE_RATES_USER"),
	}

	return ProjectConfig, nil
}
