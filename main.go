package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/micromdm/dep"
	"github.com/micromdm/micromdm/device"
	"github.com/micromdm/micromdm/management"
	"github.com/micromdm/micromdm/workflow"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
)

var (
	// Version info
	Version = "unreleased"
	gitHash = "unknown"
)

func main() {
	ctx := context.Background()
	logger := log.NewLogfmtLogger(os.Stderr)

	//flags
	var (
		flPort      = flag.String("port", envString("MICROMDM_HTTP_LISTEN_PORT", ""), "port to listen on")
		flTLS       = flag.Bool("tls", envBool("MICROMDM_USE_TLS"), "use https")
		flTLSCert   = flag.String("tls-cert", envString("MICROMDM_TLS_CERT", ""), "path to TLS certificate")
		flTLSKey    = flag.String("tls-key", envString("MICROMDM_TLS_KEY", ""), "path to TLS private key")
		flPGconn    = flag.String("postgres", envString("MICROMDM_POSTGRES_CONN_URL", ""), "postgres connection url")
		flRedisconn = flag.String("redis", envString("MICROMDM_REDIS_CONN_URL", ""), "redis connection url")
		flVersion   = flag.Bool("version", false, "print version information")
		// flPushCert   = flag.String("push-cert", envString("MICROMDM_PUSH_CERT", ""), "path to push certificate")
		// flPushPass   = flag.String("push-pass", envString("MICROMDM_PUSH_PASS", ""), "push certificate password")
		flEnrollment = flag.String("profile", envString("MICROMDM_ENROLL_PROFILE", ""), "path to enrollment profile")
	)

	// set tls to true by default. let user set it to false
	*flTLS = true
	flag.Parse()

	// -version flag
	if *flVersion {
		fmt.Printf("micromdm - Version %s\n", Version)
		fmt.Printf("Git Hash - %s\n", gitHash)
		os.Exit(0)
	}

	// check port flag
	// if none is provided, default to 80 or 443
	if *flPort == "" {
		port := defaultPort(*flTLS)
		logger.Log("msg", fmt.Sprintf("No port flag specified. Using %v by default", port))
		*flPort = port
	}

	if *flEnrollment == "" {
		logger.Log("err", "must set path to enrollment profile")
		os.Exit(1)
	}

	// check cert and key if -tls=true
	if *flTLS {
		if err := checkTLSFlags(*flTLSKey, *flTLSCert); err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}
	}

	pgHostAddr := os.Getenv("POSTGRES_PORT_5432_TCP_ADDR")
	if *flPGconn == "" && pgHostAddr != "" {
		*flPGconn = getPGConnFromENV(logger, pgHostAddr)
	}

	// check database connection
	if *flPGconn == "" {
		logger.Log("err", "database connection url not specified")
		os.Exit(1)
	}

	workflowDB, err := workflow.NewDB(
		"postgres",
		*flPGconn,
		logger,
	)
	if err != nil {
		logger.Log("err", err)
		os.Exit(1)
	}

	deviceDB, err := device.NewDB(
		"postgres",
		*flPGconn,
		logger,
	)
	if err != nil {
		logger.Log("err", err)
		os.Exit(1)
	}

	redisHostAddr := os.Getenv("REDIS_PORT_6379_TCP_ADDR")
	if *flRedisconn == "" && redisHostAddr != "" {
		*flRedisconn = getRedisConnFromENV(redisHostAddr)
	}

	// check database connection
	if *flRedisconn == "" {
		logger.Log("err", "database connection url not specified")
		os.Exit(1)
	}

	dc := depClient(logger)
	mgmtSvc := management.NewService(deviceDB, workflowDB, dc)

	httpLogger := log.NewContext(logger).With("component", "http")
	mux := http.NewServeMux()
	mux.Handle("/management/v1/", management.ServiceHandler(ctx, mgmtSvc, httpLogger))

	http.Handle("/", mux)
	http.Handle("/metrics", stdprometheus.Handler())

	serve(logger, *flTLS, *flPort, *flTLSKey, *flTLSCert)
}

func depClient(logger log.Logger) dep.Client {
	config := &dep.Config{
		ConsumerKey:    "CK_48dd68d198350f51258e885ce9a5c37ab7f98543c4a697323d75682a6c10a32501cb247e3db08105db868f73f2c972bdb6ae77112aea803b9219eb52689d42e6",
		ConsumerSecret: "CS_34c7b2b531a600d99a0e4edcf4a78ded79b86ef318118c2f5bcfee1b011108c32d5302df801adbe29d446eb78f02b13144e323eb9aad51c79f01e50cb45c3a68",
		AccessToken:    "AT_927696831c59ba510cfe4ec1a69e5267c19881257d4bca2906a99d0785b785a6f6fdeb09774954fdd5e2d0ad952e3af52c6d8d2f21c924ba0caf4a031c158b89",
		AccessSecret:   "AS_c31afd7a09691d83548489336e8ff1cb11b82b6bca13f793344496a556b1f4972eaff4dde6deb5ac9cf076fdfa97ec97699c34d515947b9cf9ed31c99dded6ba",
	}
	dc, err := dep.NewClient(config, dep.ServerURL("http://localhost:9000"))
	if err != nil {
		logger.Log("err", err)
		os.Exit(1)

	}
	return dc
}

// choose http or https
func serve(logger log.Logger, tls bool, port, key, cert string) {
	portStr := fmt.Sprintf(":%v", port)
	if tls {
		logger.Log("msg", "HTTPs", "addr", port)
		logger.Log("err", http.ListenAndServeTLS(portStr, cert, key, nil))
	} else {
		logger.Log("msg", "HTTP", "addr", port)
		logger.Log("err", http.ListenAndServe(portStr, nil))
	}
}

func envString(key, def string) string {
	if env := os.Getenv(key); env != "" {
		return env
	}
	return def
}

func envBool(key string) bool {
	if env := os.Getenv(key); env == "true" {
		return true
	}
	return false
}

func checkTLSFlags(key, cert string) error {
	if key == "" || cert == "" {
		return errors.New("You must provide a valid path to a TLS cert and key")
	}
	return nil
}

func defaultPort(tls bool) string {
	if tls {
		return "443"
	}
	return "80"
}

// use this in docker container
func getPGConnFromENV(logger log.Logger, host string) string {
	user := os.Getenv("POSTGRES_ENV_POSTGRES_USER")
	if user == "" {
		user = "postgres"
	}
	dbname := os.Getenv("POSTGRES_ENV_POSTGRES_DB")
	if dbname == "" {
		dbname = user //same defaults as the docker pgcontainer
	}
	password := os.Getenv("POSTGRES_ENV_POSTGRES_PASSWORD")
	if password == "" {
		password = "postgres"
	}
	sslmode := os.Getenv("POSTGRES_ENV_SSLMODE")
	if sslmode == "" {
		logger.Log("msg", "POSTGRES_ENV_SSLMODE not specified, using 'require' by default")
		sslmode = "require"
	}
	conn := fmt.Sprintf("user=%v password=%v dbname=%v sslmode=%v host=%v", user, password, dbname, sslmode, host)
	return conn
}

func getRedisConnFromENV(host string) string {
	port := os.Getenv("REDIS_PORT_6379_TCP_PORT")
	conn := fmt.Sprintf("%v:%v", host, port)
	return conn
}
