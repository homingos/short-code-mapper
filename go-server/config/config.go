package config

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleWare "github.com/go-chi/chi/v5/middleware"
	"github.com/homingos/campaign-svc/types/consts"
	"github.com/homingos/flam-go-common/authz"
	appMiddleware "github.com/homingos/flam-go-common/middleware"
	"github.com/homingos/flam-go-common/server"
	"github.com/homingos/flam-go-common/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"

	"os"
	"strconv"

	"github.com/go-chi/metrics"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/rs/cors"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

// Configurations ...
type Configurations struct {
	DB                DBConfig
	Milvus            MilvusConfig
	SERVER            server.ServerConfig
	GCP               GCP
	ENV               string
	JWT               JWT
	REDIS             REDISConfig
	NatsHost          string
	BaseURL           string
	InterServiceToken string
	OpenFGA           authz.OpenFGAConfig
	UserSvcBaseURL    string
	PaymentSvcBaseURL string
	EmbeddingModel    EmbeddingModelConfig
}

// GCP Credential

type PosthogConfig struct {
	PostHogApiKey string
	Endpoint      string
}

type EmbeddingModelConfig struct {
	URL    string
	APIKey string
}

type GCP struct {
	ClientEmail string
	PrivateKey  string
	GCS_BUCKET  string
	ProjectID   string
}

type JWT struct {
	SecretKey string
}

type Config struct {
	CloudFunctionURL string `env:"CLOUD_FUNCTION_URL"`
}

// DBConfig ...
type DBConfig struct {
	URI      string
	HOST     string
	USERNAME string
	PASSWORD string
	NAME     string
	PORT     string
}

// Milvus
type MilvusConfig struct {
	Host           string
	Key            string
	CollectionName string
}
type MilvusClient struct {
	Client client.Client
}

type REDISConfig struct {
	URI string
}

type AspectoConfig struct {
	ServiceName string
	Token       string
}

// GetAppConfig ...
func GetAppConfig() *Configurations {
	appConfig := LoadConfig()
	return &appConfig
}

func NewAspectoConfig() *AspectoConfig {
	return &AspectoConfig{
		ServiceName: consts.SERVICE_NAME,
		Token:       os.Getenv("ASPECTO_AUTH"),
	}
}

func NewPosthogConfig() *PosthogConfig {
	return &PosthogConfig{
		PostHogApiKey: os.Getenv("POSTHOG_API_KEY"),
		Endpoint:      os.Getenv("POSTHOG_ENDPOINT"),
	}
}

// GetDBConn ...
func GetDBConn(ctx context.Context, log *zap.SugaredLogger, conf DBConfig) *mongo.Database {
	dbConn := configureDatabase(ctx, log, conf)
	return dbConn
}

// LoadConfig ...
func LoadConfig() Configurations {
	var env map[string]string
	var conf Configurations
	viper.SetConfigName(".env") // config file name
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv() // read value ENV variable
	// viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("fatal error config file: default \n", err)
		os.Exit(1)
	}

	err = viper.Unmarshal(&env)
	if err != nil {
		fmt.Println("fatal error config file: default \n", err)
		os.Exit(1)
	}
	conf.DB = DBConfig{
		URI:  env["mongodb_url"],
		NAME: env["db_name"],
	}
	conf.ENV = os.Getenv("APP_ENV")
	// Server Config
	port, _ := strconv.Atoi(env["server_port"])
	idleTimeout, _ := strconv.Atoi(env["server_idletimeout"])
	readTimeout, _ := strconv.Atoi(env["server_readtimeout"])
	shutdownTimeout, _ := strconv.Atoi(env["server_shutdownwait"])
	writeTimeout, _ := strconv.Atoi(env["server_writetimeout"])
	conf.SERVER = server.ServerConfig{
		Port:         port,
		IdleTimeout:  idleTimeout,
		ReadTimeout:  readTimeout,
		ShutdownWait: shutdownTimeout,
		WriteTimeout: writeTimeout,
	}
	// GCP Cloud Config
	conf.GCP = GCP{
		PrivateKey:  env["google_private_key"],
		ClientEmail: env["google_client_email"],
		GCS_BUCKET:  env["gcs_bucket"],
		ProjectID:   env["gcp_project_id"],
	}

	conf.JWT = JWT{SecretKey: env["jwt_secret_key"]}
	conf.REDIS = REDISConfig{URI: env["redis_url"]}
	conf.NatsHost = env["nats_host"]
	conf.BaseURL = env["base_url"]
	conf.InterServiceToken = env["interservice_token"]
	conf.OpenFGA = authz.OpenFGAConfig{
		URL:         env["openfga_url"],
		StoreID:     env["openfga_store_id"],
		AuthModelID: env["openfga_auth_model_id"],
	}
	conf.UserSvcBaseURL = env["user_svc_dns_url"]
	conf.PaymentSvcBaseURL = env["payment_svc_dns_url"]
	conf.Milvus = MilvusConfig{
		Host:           env["milvus_host"],
		Key:            env["milvus_api_key"],
		CollectionName: env["milvus_collection"],
	}
	conf.EmbeddingModel = EmbeddingModelConfig{
		URL:    env["embedding_api_url"],
		APIKey: env["embedding_api_key"],
	}
	return conf
}

func configureDatabase(ctx context.Context, lgr *zap.SugaredLogger, conf DBConfig) *mongo.Database {
	if conf.URI == "" {
		lgr.Fatal("Set MongoDB URI in your config.yaml file")
	}

	client, err := mongo.Connect(ctx, options.Client().SetMonitor(otelmongo.NewMonitor()), options.Client().ApplyURI(conf.URI))
	if err != nil {
		lgr.Fatal("Error occured connecting to the Database")
	}

	database := client.Database(conf.NAME)
	return database
}

func ConfigureMilvusDatabase(address string, key string) MilvusClient {
	milvusClient, err := client.NewClient(context.Background(), client.Config{
		Address: address,
		APIKey:  key,
		DBName:  "default",
	})
	if err != nil {
		log.Fatal(err)
	}
	return MilvusClient{Client: milvusClient}
}

func InitRouter(logger *zap.SugaredLogger, appConfig *Configurations, tp *sdktrace.TracerProvider, serviceName string) *chi.Mux {
	router := chi.NewRouter()
	cors := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "OPTIONS", "DELETE"},
		AllowedHeaders: []string{
			"Origin", "Authorization", "Access-Control-Allow-Origin",
			"Access-Control-Allow-Header", "Accept",
			"Content-Type", "X-CSRF-Token", "Last-Event-ID",
			"Cache-Control",
			"Access-Control-Allow-Headers", "Content-Length, Accept-Encoding, Priority",
		},
		ExposedHeaders: []string{
			"Content-Length", "Access-Control-Allow-Origin", "Origin",
			"Cache-Control", "Content-Type",
		},
		AllowCredentials: true,
		MaxAge:           300,
	})

	healthRoute := fmt.Sprintf("%s%s%s", "/", consts.RoutePrefix, "/health")

	// // Setup metrics
	// metricConfig, err := utils.NewMetricConfig(serviceName)
	// if err != nil {
	// 	log.Fatalf("unable to initialize metrics due: %v", err)
	// }

	// Create default capture configuration
	captureConfig := telemetry.DefaultCaptureConfig()
	// Add specific paths to skip
	captureConfig.SkipPaths = append(captureConfig.SkipPaths,
		"/health",
		"google.devtools.cloudtrace",
		"BatchWriteSpans",
		"/progress",
	)

	router.Use(
		cors.Handler,
		metrics.Collector(metrics.CollectorOpts{
			Host:  false,
			Proto: true,
			Skip: func(r *http.Request) bool {
				return r.URL.Path == healthRoute || r.URL.Path == "/metrics"
			},
		}),
		chiMiddleWare.RequestID,
		chiMiddleWare.Logger,
		appMiddleware.Recoverer,
		// Basic tracing middleware to ensure correct path naming
		telemetry.TracingMiddleware(tp),
		// Enhanced capture middleware for detailed request/response data
		telemetry.CaptureDetailsMiddleware(captureConfig),
		// Keep metrics middleware
		chiMiddleWare.Heartbeat(healthRoute),
	)
	router.Handle(fmt.Sprintf("/%s/metrics", consts.RoutePrefix), metrics.Handler())

	return router
}
