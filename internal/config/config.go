package config

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-errors/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

type Config struct {
	HTTP               HTTP         `json:"http"`
	Persistence        Persistence  `json:"persistence"`
	Registration       Registration `json:"registration"`
	Auth               Auth         `json:"auth"`
	JWT                JWT          `json:"jwt"`
	Mapbox             Mapbox       `json:"mapbox"`
	NATS               NATS         `json:"nats"`
	ParallelLogParsers uint         `json:"parallel_log_parsers" yaml:"parallel_log_parsers"`
	LogLevel           LogLevel     `json:"log_level" yaml:"log_level"`
}

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

type NATS struct {
	Enabled bool   `json:"enabled"`
	URL     string `json:"url"`
	Token   string `json:"token"`
}

type Sentinel struct {
	Enabled    bool     `json:"enabled"`
	MasterName string   `json:"master_name" yaml:"master_name"`
	Addresses  []string `json:"addresses"`
	Password   string   `json:"password"`
	Username   string   `json:"username"`
}

type JWT struct {
	Secret string `json:"secret"`
}

type Auth struct {
	Google Google `json:"google"`
	GitHub GitHub `json:"github"`
	Custom Custom `json:"custom"`
}

type Mapbox struct {
	SecretToken string `json:"secret_token" yaml:"secret_token"`
	PublicToken string `json:"public_token" yaml:"public_token"`
}

type Google struct {
	Enabled      bool   `json:"enabled"`
	ClientID     string `json:"client_id" yaml:"client_id"`
	ClientSecret string `json:"client_secret" yaml:"client_secret"`
}

type GitHub struct {
	Enabled      bool   `json:"enabled"`
	ClientID     string `json:"client_id" yaml:"client_id"`
	ClientSecret string `json:"client_secret" yaml:"client_secret"`
}

type Custom struct {
	Enabled      bool   `json:"enabled"`
	ClientID     string `json:"client_id" yaml:"client_id"`
	ClientSecret string `json:"client_secret" yaml:"client_secret"`
	TokenURL     string `json:"token_url" yaml:"token_url"`
	UserURL      string `json:"user_url" yaml:"user_url"`
}

type Registration struct {
	Enabled bool `json:"enabled"`
}

type Persistence struct {
	Database Database `json:"database"`
	Uploads  Uploads  `json:"uploads"`
}

type UploadsDriver string

const (
	UploadsDriverFilesystem UploadsDriver = "filesystem"
	UploadsDriverS3         UploadsDriver = "s3"
)

type Uploads struct {
	Driver            UploadsDriver     `json:"driver"`
	FilesystemOptions FilesystemOptions `json:"filesystem_options" yaml:"filesystem_options"`
	S3Options         S3Options         `json:"s3_options" yaml:"s3_options"`
}

type FilesystemOptions struct {
	Directory string `json:"directory"`
}

type S3Options struct {
	Region   string `json:"region"`
	Bucket   string `json:"bucket"`
	Endpoint string `json:"endpoint"`
}

type DatabaseDriver string

const (
	DatabaseDriverSQLite   DatabaseDriver = "sqlite"
	DatabaseDriverMySQL    DatabaseDriver = "mysql"
	DatabaseDriverPostgres DatabaseDriver = "postgres"
)

type Database struct {
	Driver          DatabaseDriver `json:"driver"`
	Database        string         `json:"database"`
	Username        string         `json:"username"`
	Password        string         `json:"password"`
	Host            string         `json:"host"`
	Port            uint16         `json:"port"`
	ExtraParameters string         `json:"extra_perimeters" yaml:"extra_perimeters"`
}

type HTTPListener struct {
	IPV4Host string `json:"ipv4_host" yaml:"ipv4_host"`
	IPV6Host string `json:"ipv6_host" yaml:"ipv6_host"`
	Port     uint16 `json:"port"`
}

type Tracing struct {
	Enabled      bool   `json:"enabled"`
	OTLPEndpoint string `json:"otlp_endpoint" yaml:"otlp_endpoint"`
}

type PProf struct {
	Enabled bool `json:"enabled"`
}

type Metrics struct {
	HTTPListener
	Enabled bool `json:"enabled"`
}

type HTTP struct {
	HTTPListener
	Tracing
	BackendURL     string   `json:"backend_url" yaml:"backend_url"`
	PProf          PProf    `json:"pprof"`
	TrustedProxies []string `json:"trusted_proxies" yaml:"trusted_proxies"`
	Metrics        Metrics  `json:"metrics"`
	CORSHosts      []string `json:"cors_hosts" yaml:"cors_hosts"`
}

//nolint:golint,gochecknoglobals
var (
	ConfigFileKey                                   = "config"
	HTTPIPV4HostKey                                 = "http.ipv4_host"
	HTTPIPV6HostKey                                 = "http.ipv6_host"
	HTTPPortKey                                     = "http.port"
	HTTPTracingEnabledKey                           = "http.tracing.enabled"
	HTTPTracingOTLPEndKey                           = "http.tracing.otlp_endpoint"
	HTTPPProfEnabledKey                             = "http.pprof.enabled"
	HTTPTrustedProxiesKey                           = "http.trusted_proxies"
	HTTPMetricsEnabledKey                           = "http.metrics.enabled"
	HTTPMetricsIPV4HostKey                          = "http.metrics.ipv4_host"
	HTTPMetricsIPV6HostKey                          = "http.metrics.ipv6_host"
	HTTPMetricsPortKey                              = "http.metrics.port"
	HTTPCORSHostsKey                                = "http.cors_hosts"
	HTTPBackendURLKey                               = "http.backend_url"
	PersistenceDatabaseDriverKey                    = "persistence.database.driver"
	PersistenceDatabaseDatabaseKey                  = "persistence.database.database"
	PersistenceDatabaseUsernameKey                  = "persistence.database.username"
	PersistenceDatabasePasswordKey                  = "persistence.database.password"
	PersistenceDatabaseHostKey                      = "persistence.database.host"
	PersistenceDatabasePortKey                      = "persistence.database.port"
	PersistenceDatabaseExtraParametersKey           = "persistence.database.extra_parameters"
	PersistenceUploadsDriverKey                     = "persistence.uploads.driver"
	PersistenceUploadsFilesystemOptionsDirectoryKey = "persistence.uploads.filesystem_options.directory"
	PersistenceUploadsS3OptionsBucketKey            = "persistence.uploads.s3_options.bucket"
	PersistenceUploadsS3OptionsRegionKey            = "persistence.uploads.s3_options.region"
	PersistenceUploadsS3OptionsEndpointKey          = "persistence.uploads.s3_options.endpoint"
	RegistrationEnabledKey                          = "registration.enabled"
	AuthGoogleEnabledKey                            = "auth.google.enabled"
	AuthGoogleClientIDKey                           = "auth.google.client_id"
	//nolint:golint,gosec
	AuthGoogleClientSecretKey = "auth.google.client_secret"
	AuthGitHubEnabledKey      = "auth.github.enabled"
	AuthGitHubClientIDKey     = "auth.github.client_id"
	//nolint:golint,gosec
	AuthGitHubClientSecretKey = "auth.github.client_secret"
	AuthCustomEnabledKey      = "auth.custom.enabled"
	AuthCustomClientIDKey     = "auth.custom.client_id"
	//nolint:golint,gosec
	AuthCustomClientSecretKey = "auth.custom.client_secret"
	//nolint:golint,gosec
	AuthCustomTokenURLKey = "auth.custom.token_url"
	AuthCustomUserURLKey  = "auth.custom.user_url"
	JWTSecretKey          = "jwt.secret"
	MapboxPublicTokenKey  = "mapbox.public_token"
	MapboxSecretTokenKey  = "mapbox.secret_token"
	NATSEnabledKey        = "nats.enabled"
	NATSURLKey            = "nats.url"
	NATSTokenKey          = "nats.token"
	LogLevelKey           = "log_level"
	ParallelLogParsersKey = "parallel_log_parsers"
)

const (
	DefaultConfigPath                                   = "config.yaml"
	DefaultHTTPIPV4Host                                 = "0.0.0.0"
	DefaultHTTPIPV6Host                                 = "::"
	DefaultHTTPPort                                     = 8080
	DefaultHTTPMetricsIPV4Host                          = "127.0.0.1"
	DefaultHTTPMetricsIPV6Host                          = "::1"
	DefaultHTTPMetricsPort                              = 8081
	DefaultPersistenceDatabaseDriver                    = DatabaseDriverSQLite
	DefaultPersistenceDatabaseDatabase                  = "rtz.db"
	DefaultRegistrationEnabled                          = false
	DefaultNATSEnabled                                  = false
	DefaultAuthGitHubEnabled                            = false
	DefaultAuthGoogleEnabled                            = false
	DefaultAuthCustomEnabled                            = false
	DefaultLogLevel                                     = LogLevelInfo
	DefaultParallelLogParsers                           = 4
	DefaultPersistenceUploadsDriver                     = UploadsDriverFilesystem
	DefaultPersistenceUploadsFilesystemOptionsDirectory = "uploads/"
)

func RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringP(ConfigFileKey, "c", DefaultConfigPath, "Config file path")
	cmd.Flags().String(HTTPIPV4HostKey, DefaultHTTPIPV4Host, "HTTP server IPv4 host")
	cmd.Flags().String(HTTPIPV6HostKey, DefaultHTTPIPV6Host, "HTTP server IPv6 host")
	cmd.Flags().Uint16(HTTPPortKey, DefaultHTTPPort, "HTTP server port")
	cmd.Flags().Bool(HTTPTracingEnabledKey, false, "Enable Open Telemetry tracing")
	cmd.Flags().String(HTTPTracingOTLPEndKey, "", "Open Telemetry endpoint")
	cmd.Flags().Bool(HTTPPProfEnabledKey, false, "Enable pprof")
	cmd.Flags().StringSlice(HTTPTrustedProxiesKey, []string{}, "Comma-separated list of trusted proxies")
	cmd.Flags().Bool(HTTPMetricsEnabledKey, false, "Enable metrics server")
	cmd.Flags().String(HTTPMetricsIPV4HostKey, DefaultHTTPMetricsIPV4Host, "Metrics server IPv4 host")
	cmd.Flags().String(HTTPMetricsIPV6HostKey, DefaultHTTPMetricsIPV6Host, "Metrics server IPv6 host")
	cmd.Flags().Uint16(HTTPMetricsPortKey, DefaultHTTPMetricsPort, "Metrics server port")
	cmd.Flags().StringSlice(HTTPCORSHostsKey, []string{}, "Comma-separated list of CORS hosts")
	cmd.Flags().String(HTTPBackendURLKey, "", "Backend URL")
	cmd.Flags().String(PersistenceDatabaseDriverKey, string(DefaultPersistenceDatabaseDriver), "Database driver, one of: sqlite, mysql, postgres")
	cmd.Flags().String(PersistenceDatabaseDatabaseKey, DefaultPersistenceDatabaseDatabase, "Database path")
	cmd.Flags().String(PersistenceDatabaseUsernameKey, "", "Database username")
	cmd.Flags().String(PersistenceDatabasePasswordKey, "", "Database password")
	cmd.Flags().String(PersistenceDatabaseHostKey, "", "Database host")
	cmd.Flags().Uint16(PersistenceDatabasePortKey, 0, "Database port")
	cmd.Flags().String(PersistenceDatabaseExtraParametersKey, "", "Database extra parameters")
	cmd.Flags().String(PersistenceUploadsDriverKey, string(DefaultPersistenceUploadsDriver), "Uploads driver, one of: memory, filesystem, s3")
	cmd.Flags().String(PersistenceUploadsFilesystemOptionsDirectoryKey, DefaultPersistenceUploadsFilesystemOptionsDirectory, "Filesystem uploads directory")
	cmd.Flags().String(PersistenceUploadsS3OptionsBucketKey, "", "S3 bucket")
	cmd.Flags().String(PersistenceUploadsS3OptionsRegionKey, "", "S3 region")
	cmd.Flags().String(PersistenceUploadsS3OptionsEndpointKey, "", "S3 endpoint")
	cmd.Flags().Bool(RegistrationEnabledKey, DefaultRegistrationEnabled, "Enable registration")
	cmd.Flags().Bool(AuthGoogleEnabledKey, DefaultAuthGoogleEnabled, "Enable Google OAuth")
	cmd.Flags().String(AuthGoogleClientIDKey, "", "Google OAuth client ID")
	cmd.Flags().String(AuthGoogleClientSecretKey, "", "Google OAuth client secret")
	cmd.Flags().Bool(AuthGitHubEnabledKey, DefaultAuthGitHubEnabled, "Enable GitHub OAuth")
	cmd.Flags().String(AuthGitHubClientIDKey, "", "GitHub OAuth client ID")
	cmd.Flags().String(AuthGitHubClientSecretKey, "", "GitHub OAuth client secret")
	cmd.Flags().Bool(AuthCustomEnabledKey, DefaultAuthCustomEnabled, "Enable custom OAuth")
	cmd.Flags().String(AuthCustomClientIDKey, "", "Custom OAuth client ID")
	cmd.Flags().String(AuthCustomClientSecretKey, "", "Custom OAuth client secret")
	cmd.Flags().String(AuthCustomTokenURLKey, "", "Custom OAuth token URL")
	cmd.Flags().String(AuthCustomUserURLKey, "", "Custom OAuth user URL")
	cmd.Flags().String(JWTSecretKey, "", "JWT signing secret")
	cmd.Flags().String(MapboxPublicTokenKey, "", "Mapbox public token")
	cmd.Flags().String(MapboxSecretTokenKey, "", "Mapbox secret token")
	cmd.Flags().Bool(NATSEnabledKey, DefaultNATSEnabled, "Enable NATS")
	cmd.Flags().String(NATSURLKey, "", "NATS URL")
	cmd.Flags().String(NATSTokenKey, "", "NATS token")
	cmd.Flags().String(LogLevelKey, string(DefaultLogLevel), "Log level")
	cmd.Flags().Uint(ParallelLogParsersKey, DefaultParallelLogParsers, "Number of parallel log parsers")
}

var (
	ErrJWTSecretRequired          = errors.New("JWT secret is required")
	ErrBackendURLRequired         = errors.New("Backend URL is required")
	ErrOTLPEndpointRequired       = errors.New("OTLP endpoint is required when tracing is enabled")
	ErrMapboxPublicTokenRequired  = errors.New("Mapbox public token is required")
	ErrMapboxSecretTokenRequired  = errors.New("Mapbox secret token is required")
	ErrDBHostRequired             = errors.New("Database host is required")
	ErrDBDatabaseRequired         = errors.New("Database name is required")
	ErrDatabaseDriverRequired     = errors.New("Database driver is required")
	ErrNATSURLRequired            = errors.New("NATS URL is required")
	ErrGitHubOAuthRequired        = errors.New("GitHub OAuth client ID and secret are required")
	ErrGoogleOAuthRequired        = errors.New("Google OAuth client ID and secret are required")
	ErrCustomOAuthRequired        = errors.New("Custom OAuth client ID and secret are required")
	ErrCustomTokenURLRequired     = errors.New("Custom OAuth token URL is required")
	ErrCustomUserURLRequired      = errors.New("Custom OAuth user URL is required")
	ErrParallelLogParsersNotZero  = errors.New("Number of parallel log parsers must be greater than zero")
	ErrInvalidLogLevel            = errors.New("Invalid log level")
	ErrUploadsFSDirectoryRequired = errors.New("Filesystem uploads directory is required")
	ErrUploadsS3BucketRequired    = errors.New("S3 bucket is required")
	ErrUploadsS3RegionRequired    = errors.New("S3 region is required")
)

func (c *Config) Validate() error {
	if c.JWT.Secret == "" {
		return ErrJWTSecretRequired
	}
	if c.HTTP.BackendURL == "" {
		return ErrBackendURLRequired
	}
	if c.HTTP.Tracing.Enabled && c.HTTP.Tracing.OTLPEndpoint == "" {
		return ErrOTLPEndpointRequired
	}
	if c.Mapbox.PublicToken == "" {
		return ErrMapboxPublicTokenRequired
	}
	if c.Mapbox.SecretToken == "" {
		return ErrMapboxSecretTokenRequired
	}
	if c.Persistence.Database.Driver != DatabaseDriverSQLite && c.Persistence.Database.Host == "" {
		return ErrDBHostRequired
	}
	if c.Persistence.Database.Driver == "" {
		return ErrDatabaseDriverRequired
	}
	if c.Persistence.Database.Database == "" {
		return ErrDBDatabaseRequired
	}
	if c.Persistence.Uploads.Driver == UploadsDriverFilesystem && c.Persistence.Uploads.FilesystemOptions.Directory == "" {
		return ErrUploadsFSDirectoryRequired
	}
	if c.Persistence.Uploads.Driver == UploadsDriverS3 && c.Persistence.Uploads.S3Options.Bucket == "" {
		return ErrUploadsS3BucketRequired
	}
	if c.Persistence.Uploads.Driver == UploadsDriverS3 && c.Persistence.Uploads.S3Options.Region == "" {
		return ErrUploadsS3RegionRequired
	}
	if c.NATS.Enabled && c.NATS.URL == "" {
		return ErrNATSURLRequired
	}
	if c.Auth.GitHub.Enabled && (c.Auth.GitHub.ClientID == "" || c.Auth.GitHub.ClientSecret == "") {
		return ErrGitHubOAuthRequired
	}
	if c.Auth.Google.Enabled && (c.Auth.Google.ClientID == "" || c.Auth.Google.ClientSecret == "") {
		return ErrGoogleOAuthRequired
	}
	if c.Auth.Custom.Enabled && (c.Auth.Custom.ClientID == "" || c.Auth.Custom.ClientSecret == "") {
		return ErrCustomOAuthRequired
	}
	if c.Auth.Custom.Enabled && c.Auth.Custom.TokenURL == "" {
		return ErrCustomTokenURLRequired
	}
	if c.Auth.Custom.Enabled && c.Auth.Custom.UserURL == "" {
		return ErrCustomUserURLRequired
	}
	if c.ParallelLogParsers == 0 {
		return ErrParallelLogParsersNotZero
	}
	switch c.LogLevel {
	case LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError:
	default:
		return ErrInvalidLogLevel
	}

	return nil
}

func LoadConfig(cmd *cobra.Command) (*Config, error) {
	var config Config

	// Load flags from envs
	ctx, cancel := context.WithCancelCause(cmd.Context())
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if ctx.Err() != nil {
			return
		}
		optName := strings.ReplaceAll(strings.ReplaceAll(strings.ToUpper(f.Name), "-", "_"), ".", "__")
		if val, ok := os.LookupEnv(optName); !f.Changed && ok {
			if err := f.Value.Set(val); err != nil {
				cancel(err)
			}
			f.Changed = true
		}
	})
	if ctx.Err() != nil {
		return &config, fmt.Errorf("failed to load env: %w", context.Cause(ctx))
	}

	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return &config, fmt.Errorf("failed to get config path: %w", err)
	}
	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return &config, fmt.Errorf("failed to read config: %w", err)
		} else if err == nil {
			if err := yaml.Unmarshal(data, &config); err != nil {
				return &config, fmt.Errorf("failed to unmarshal config: %w", err)
			}
		}
	}

	err = overrideFlags(&config, cmd)
	if err != nil {
		return &config, fmt.Errorf("failed to override flags: %w", err)
	}

	// Defaults
	if config.HTTP.IPV4Host == "" {
		config.HTTP.IPV4Host = DefaultHTTPIPV4Host
	}
	if config.HTTP.IPV6Host == "" {
		config.HTTP.IPV6Host = DefaultHTTPIPV6Host
	}
	if config.HTTP.Port == 0 {
		config.HTTP.Port = DefaultHTTPPort
	}
	if config.HTTP.Metrics.IPV4Host == "" {
		config.HTTP.Metrics.IPV4Host = DefaultHTTPMetricsIPV4Host
	}
	if config.HTTP.Metrics.IPV6Host == "" {
		config.HTTP.Metrics.IPV6Host = DefaultHTTPMetricsIPV6Host
	}
	if config.HTTP.Metrics.Port == 0 {
		config.HTTP.Metrics.Port = DefaultHTTPMetricsPort
	}
	if config.Persistence.Database.Driver == "" {
		config.Persistence.Database.Driver = DefaultPersistenceDatabaseDriver
	}
	if config.Persistence.Database.Database == "" {
		config.Persistence.Database.Database = DefaultPersistenceDatabaseDatabase
	}
	if config.Persistence.Uploads.Driver == "" {
		config.Persistence.Uploads.Driver = DefaultPersistenceUploadsDriver
	}
	if config.Persistence.Uploads.FilesystemOptions.Directory == "" {
		config.Persistence.Uploads.FilesystemOptions.Directory = DefaultPersistenceUploadsFilesystemOptionsDirectory
	}
	if config.ParallelLogParsers == 0 {
		config.ParallelLogParsers = DefaultParallelLogParsers
	}
	if config.LogLevel == "" {
		config.LogLevel = DefaultLogLevel
	}

	return &config, nil
}

func overrideFlags(config *Config, cmd *cobra.Command) error {
	var err error
	if cmd.Flags().Changed(HTTPIPV4HostKey) {
		config.HTTP.IPV4Host, err = cmd.Flags().GetString(HTTPIPV4HostKey)
		if err != nil {
			return fmt.Errorf("failed to get HTTP IPv4 host: %w", err)
		}
	}

	if cmd.Flags().Changed(HTTPIPV6HostKey) {
		config.HTTP.IPV6Host, err = cmd.Flags().GetString(HTTPIPV6HostKey)
		if err != nil {
			return fmt.Errorf("failed to get HTTP IPv6 host: %w", err)
		}
	}

	if cmd.Flags().Changed(HTTPPortKey) {
		config.HTTP.Port, err = cmd.Flags().GetUint16(HTTPPortKey)
		if err != nil {
			return fmt.Errorf("failed to get HTTP port: %w", err)
		}
	}

	if cmd.Flags().Changed(HTTPPProfEnabledKey) {
		config.HTTP.PProf.Enabled, err = cmd.Flags().GetBool(HTTPPProfEnabledKey)
		if err != nil {
			return fmt.Errorf("failed to get pprof enabled: %w", err)
		}
	}

	if cmd.Flags().Changed(HTTPTrustedProxiesKey) {
		config.HTTP.TrustedProxies, err = cmd.Flags().GetStringSlice(HTTPTrustedProxiesKey)
		if err != nil {
			return fmt.Errorf("failed to get trusted proxies: %w", err)
		}
	}

	if cmd.Flags().Changed(HTTPMetricsEnabledKey) {
		config.HTTP.Metrics.Enabled, err = cmd.Flags().GetBool(HTTPMetricsEnabledKey)
		if err != nil {
			return fmt.Errorf("failed to get metrics enabled: %w", err)
		}
	}

	if cmd.Flags().Changed(HTTPMetricsIPV4HostKey) {
		config.HTTP.Metrics.IPV4Host, err = cmd.Flags().GetString(HTTPMetricsIPV4HostKey)
		if err != nil {
			return fmt.Errorf("failed to get metrics IPv4 host: %w", err)
		}
	}

	if cmd.Flags().Changed(HTTPMetricsIPV6HostKey) {
		config.HTTP.Metrics.IPV6Host, err = cmd.Flags().GetString(HTTPMetricsIPV6HostKey)
		if err != nil {
			return fmt.Errorf("failed to get metrics IPv6 host: %w", err)
		}
	}

	if cmd.Flags().Changed(HTTPMetricsPortKey) {
		config.HTTP.Metrics.Port, err = cmd.Flags().GetUint16(HTTPMetricsPortKey)
		if err != nil {
			return fmt.Errorf("failed to get metrics port: %w", err)
		}
	}

	if cmd.Flags().Changed(HTTPTracingEnabledKey) {
		config.HTTP.Tracing.Enabled, err = cmd.Flags().GetBool(HTTPTracingEnabledKey)
		if err != nil {
			return fmt.Errorf("failed to get tracing enabled: %w", err)
		}
	}

	if cmd.Flags().Changed(HTTPTracingOTLPEndKey) {
		config.HTTP.Tracing.OTLPEndpoint, err = cmd.Flags().GetString(HTTPTracingOTLPEndKey)
		if err != nil {
			return fmt.Errorf("failed to get tracing OTLP endpoint: %w", err)
		}
	}

	if cmd.Flags().Changed(HTTPCORSHostsKey) {
		config.HTTP.CORSHosts, err = cmd.Flags().GetStringSlice(HTTPCORSHostsKey)
		if err != nil {
			return fmt.Errorf("failed to get CORS hosts: %w", err)
		}
	}

	if cmd.Flags().Changed(HTTPBackendURLKey) {
		config.HTTP.BackendURL, err = cmd.Flags().GetString(HTTPBackendURLKey)
		if err != nil {
			return fmt.Errorf("failed to get backend URL: %w", err)
		}
	}

	if cmd.Flags().Changed(PersistenceDatabaseDriverKey) {
		drvr, err := cmd.Flags().GetString(PersistenceDatabaseDriverKey)
		if err != nil {
			return fmt.Errorf("failed to get database driver: %w", err)
		}
		config.Persistence.Database.Driver = DatabaseDriver(strings.ToLower(drvr))
	}

	if cmd.Flags().Changed(PersistenceDatabaseDatabaseKey) {
		config.Persistence.Database.Database, err = cmd.Flags().GetString(PersistenceDatabaseDatabaseKey)
		if err != nil {
			return fmt.Errorf("failed to get database name: %w", err)
		}
	}

	if cmd.Flags().Changed(PersistenceDatabaseUsernameKey) {
		config.Persistence.Database.Username, err = cmd.Flags().GetString(PersistenceDatabaseUsernameKey)
		if err != nil {
			return fmt.Errorf("failed to get database username: %w", err)
		}
	}

	if cmd.Flags().Changed(PersistenceDatabasePasswordKey) {
		config.Persistence.Database.Password, err = cmd.Flags().GetString(PersistenceDatabasePasswordKey)
		if err != nil {
			return fmt.Errorf("failed to get database password: %w", err)
		}
	}

	if cmd.Flags().Changed(PersistenceDatabaseHostKey) {
		config.Persistence.Database.Host, err = cmd.Flags().GetString(PersistenceDatabaseHostKey)
		if err != nil {
			return fmt.Errorf("failed to get database host: %w", err)
		}
	}

	if cmd.Flags().Changed(PersistenceDatabasePortKey) {
		config.Persistence.Database.Port, err = cmd.Flags().GetUint16(PersistenceDatabasePortKey)
		if err != nil {
			return fmt.Errorf("failed to get database port: %w", err)
		}
	}

	if cmd.Flags().Changed(PersistenceDatabaseExtraParametersKey) {
		config.Persistence.Database.ExtraParameters, err = cmd.Flags().GetString(PersistenceDatabaseExtraParametersKey)
		if err != nil {
			return fmt.Errorf("failed to get database extra parameters: %w", err)
		}
	}

	if cmd.Flags().Changed(PersistenceUploadsDriverKey) {
		drvr, err := cmd.Flags().GetString(PersistenceUploadsDriverKey)
		if err != nil {
			return fmt.Errorf("failed to get uploads driver: %w", err)
		}
		config.Persistence.Uploads.Driver = UploadsDriver(strings.ToLower(drvr))
	}

	if cmd.Flags().Changed(PersistenceUploadsFilesystemOptionsDirectoryKey) {
		config.Persistence.Uploads.FilesystemOptions.Directory, err = cmd.Flags().GetString(PersistenceUploadsFilesystemOptionsDirectoryKey)
		if err != nil {
			return fmt.Errorf("failed to get filesystem uploads directory: %w", err)
		}
	}

	if cmd.Flags().Changed(PersistenceUploadsS3OptionsBucketKey) {
		config.Persistence.Uploads.S3Options.Bucket, err = cmd.Flags().GetString(PersistenceUploadsS3OptionsBucketKey)
		if err != nil {
			return fmt.Errorf("failed to get S3 bucket: %w", err)
		}
	}

	if cmd.Flags().Changed(PersistenceUploadsS3OptionsRegionKey) {
		config.Persistence.Uploads.S3Options.Region, err = cmd.Flags().GetString(PersistenceUploadsS3OptionsRegionKey)
		if err != nil {
			return fmt.Errorf("failed to get S3 region: %w", err)
		}
	}

	if cmd.Flags().Changed(PersistenceUploadsS3OptionsEndpointKey) {
		config.Persistence.Uploads.S3Options.Endpoint, err = cmd.Flags().GetString(PersistenceUploadsS3OptionsEndpointKey)
		if err != nil {
			return fmt.Errorf("failed to get S3 endpoint: %w", err)
		}
	}

	if cmd.Flags().Changed(RegistrationEnabledKey) {
		config.Registration.Enabled, err = cmd.Flags().GetBool(RegistrationEnabledKey)
		if err != nil {
			return fmt.Errorf("failed to get registration enabled: %w", err)
		}
	}

	if cmd.Flags().Changed(AuthGoogleClientIDKey) {
		config.Auth.Google.ClientID, err = cmd.Flags().GetString(AuthGoogleClientIDKey)
		if err != nil {
			return fmt.Errorf("failed to get Google OAuth client ID: %w", err)
		}
	}

	if cmd.Flags().Changed(AuthGoogleClientSecretKey) {
		config.Auth.Google.ClientSecret, err = cmd.Flags().GetString(AuthGoogleClientSecretKey)
		if err != nil {
			return fmt.Errorf("failed to get Google OAuth client secret: %w", err)
		}
	}

	if cmd.Flags().Changed(AuthGoogleEnabledKey) {
		config.Auth.Google.Enabled, err = cmd.Flags().GetBool(AuthGoogleEnabledKey)
		if err != nil {
			return fmt.Errorf("failed to get Google OAuth enabled: %w", err)
		}
	}

	if cmd.Flags().Changed(AuthGitHubEnabledKey) {
		config.Auth.GitHub.Enabled, err = cmd.Flags().GetBool(AuthGitHubEnabledKey)
		if err != nil {
			return fmt.Errorf("failed to get GitHub OAuth enabled: %w", err)
		}
	}

	if cmd.Flags().Changed(AuthGitHubClientIDKey) {
		config.Auth.GitHub.ClientID, err = cmd.Flags().GetString(AuthGitHubClientIDKey)
		if err != nil {
			return fmt.Errorf("failed to get GitHub OAuth client ID: %w", err)
		}
	}

	if cmd.Flags().Changed(AuthGitHubClientSecretKey) {
		config.Auth.GitHub.ClientSecret, err = cmd.Flags().GetString(AuthGitHubClientSecretKey)
		if err != nil {
			return fmt.Errorf("failed to get GitHub OAuth client secret: %w", err)
		}
	}

	if cmd.Flags().Changed(AuthCustomEnabledKey) {
		config.Auth.Custom.Enabled, err = cmd.Flags().GetBool(AuthCustomEnabledKey)
		if err != nil {
			return fmt.Errorf("failed to get custom OAuth enabled: %w", err)
		}
	}

	if cmd.Flags().Changed(AuthCustomClientIDKey) {
		config.Auth.Custom.ClientID, err = cmd.Flags().GetString(AuthCustomClientIDKey)
		if err != nil {
			return fmt.Errorf("failed to get custom OAuth client ID: %w", err)
		}
	}

	if cmd.Flags().Changed(AuthCustomClientSecretKey) {
		config.Auth.Custom.ClientSecret, err = cmd.Flags().GetString(AuthCustomClientSecretKey)
		if err != nil {
			return fmt.Errorf("failed to get custom OAuth client secret: %w", err)
		}
	}

	if cmd.Flags().Changed(AuthCustomTokenURLKey) {
		config.Auth.Custom.TokenURL, err = cmd.Flags().GetString(AuthCustomTokenURLKey)
		if err != nil {
			return fmt.Errorf("failed to get custom OAuth token URL: %w", err)
		}
	}

	if cmd.Flags().Changed(AuthCustomUserURLKey) {
		config.Auth.Custom.UserURL, err = cmd.Flags().GetString(AuthCustomUserURLKey)
		if err != nil {
			return fmt.Errorf("failed to get custom OAuth user URL: %w", err)
		}
	}

	if cmd.Flags().Changed(JWTSecretKey) {
		config.JWT.Secret, err = cmd.Flags().GetString(JWTSecretKey)
		if err != nil {
			return fmt.Errorf("failed to get JWT secret: %w", err)
		}
	}

	if cmd.Flags().Changed(MapboxPublicTokenKey) {
		config.Mapbox.PublicToken, err = cmd.Flags().GetString(MapboxPublicTokenKey)
		if err != nil {
			return fmt.Errorf("failed to get Mapbox public token: %w", err)
		}
	}

	if cmd.Flags().Changed(MapboxSecretTokenKey) {
		config.Mapbox.SecretToken, err = cmd.Flags().GetString(MapboxSecretTokenKey)
		if err != nil {
			return fmt.Errorf("failed to get Mapbox secret token: %w", err)
		}
	}

	if cmd.Flags().Changed(NATSEnabledKey) {
		config.NATS.Enabled, err = cmd.Flags().GetBool(NATSEnabledKey)
		if err != nil {
			return fmt.Errorf("failed to get NATS enabled: %w", err)
		}
	}

	if cmd.Flags().Changed(NATSURLKey) {
		config.NATS.URL, err = cmd.Flags().GetString(NATSURLKey)
		if err != nil {
			return fmt.Errorf("failed to get NATS URL: %w", err)
		}
	}

	if cmd.Flags().Changed(NATSTokenKey) {
		config.NATS.Token, err = cmd.Flags().GetString(NATSTokenKey)
		if err != nil {
			return fmt.Errorf("failed to get NATS token: %w", err)
		}
	}

	if cmd.Flags().Changed(LogLevelKey) {
		ll, err := cmd.Flags().GetString(LogLevelKey)
		if err != nil {
			return fmt.Errorf("failed to get log level: %w", err)
		}
		config.LogLevel = LogLevel(ll)
	}

	if cmd.Flags().Changed(ParallelLogParsersKey) {
		config.ParallelLogParsers, err = cmd.Flags().GetUint(ParallelLogParsersKey)
		if err != nil {
			return fmt.Errorf("failed to get number of parallel log parsers: %w", err)
		}
	}

	return nil
}
