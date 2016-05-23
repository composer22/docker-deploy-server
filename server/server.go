// Package server implements a server for docker-machine deployment.
package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	// Allow dynamic profiling.
	_ "net/http/pprof"

	"github.com/composer22/docker-deploy-server/db"
	"github.com/composer22/docker-deploy-server/logger"
	redis "gopkg.in/redis.v3"
)

// Server is the main structure that represents a server instance.
type Server struct {
	mu      sync.RWMutex   // For locking access to server attributes.
	wg      sync.WaitGroup // Synchronize shutdown pending jobs.
	running bool           // Is the server running?
	opt     *Options       // Original options used to create the server.
	db      *db.DBConnect  // Database connection.
	redis   *redis.Client  // Redis connection.
	stats   *Status        // Server statistics since it started.
	srvr    *http.Server   // HTTP server.
	done    chan bool      // A channel to signal to environments to close down.
	log     *logger.Logger // Log instance for recording error and other messages.
}

// New is a factory function that returns a new server instance.
func New(o *Options, l *logger.Logger) *Server {
	s := &Server{
		running: false,
		opt:     o,
		stats:   NewStatus(),
		done:    make(chan bool),
		log:     l,
	}

	if s.opt.Debug {
		s.log.SetLogLevel(logger.Debug)
	}

	// Setup the routes and server.
	mux := http.NewServeMux()
	mux.HandleFunc(httpRouteV1Health, s.healthHandler)
	mux.HandleFunc(httpRouteV1Info, s.infoHandler)
	mux.HandleFunc(httpRouteV1Metrics, s.metricsHandler)
	mux.HandleFunc(httpRouteV1Deploy, s.deployHandler)
	mux.HandleFunc(httpRouteV1Status, s.statusHandler)
	s.srvr = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.opt.Hostname, s.opt.Port),
		Handler:      &Middleware{serv: s, handler: mux},
		ReadTimeout:  TCPReadTimeout,
		WriteTimeout: TCPWriteTimeout,
	}
	return s
}

// PrintVersionAndExit prints the version of the server then exits.
func PrintVersionAndExit() {
	fmt.Printf("%s version %s\n", applicationName, version)
	os.Exit(0)
}

// Start spins up the server to accept incoming requests.
func (s *Server) Start() error {
	if s.isRunning() {
		return errors.New("Server already started.")
	}
	s.log.Infof("Starting %s version %s\n", applicationName, version)
	s.handleSignals()
	s.mu.Lock()

	// Connect to DB and Redis.
	var err error
	s.db, err = db.NewDBConnect(s.opt.DSN)
	if err != nil {
		s.mu.Unlock()
		return err
	}
	s.redis, err = NewRedisClient(s.opt.RedisHostname, s.opt.RedisPort, s.opt.RedisPassword, s.opt.RedisDatabase)
	if err != nil {
		s.mu.Unlock()
		return err
	}

	// Start the deployment service.
	db, err := db.NewDBConnect(s.opt.DSN)
	if err != nil {
		s.mu.Unlock()
		return err
	}
	r, err := NewRedisClient(s.opt.RedisHostname, s.opt.RedisPort, s.opt.RedisPassword, s.opt.RedisDatabase)
	if err != nil {
		s.mu.Unlock()
		return err
	}
	d := NewDeployService(s.opt, db, r, s.done, s.log, &s.wg)
	go d.Run()

	// Pprof http endpoint for the profiler.
	if s.opt.ProfPort > 0 {
		s.StartProfiler()
	}

	s.stats.Start = time.Now()
	s.running = true
	s.mu.Unlock()
	err = s.srvr.ListenAndServe()
	if err != nil {
		s.log.Emergencyf("Listen and Server Error: %s", err.Error())
	}

	// Done.
	s.mu.Lock()
	s.running = false
	s.mu.Unlock()
	return nil
}

// StartProfiler is called to enable dynamic profiling.
func (s *Server) StartProfiler() {
	s.log.Infof("Starting profiling on http port %d", s.opt.ProfPort)
	hp := fmt.Sprintf("%s:%d", s.opt.Hostname, s.opt.ProfPort)
	go func() {
		err := http.ListenAndServe(hp, nil)
		if err != nil {
			s.log.Emergencyf("Error starting profile monitoring service: %s", err)
		}
	}()
}

// Shutdown takes down the server gracefully back to an initialize state.
func (s *Server) Shutdown() {
	if !s.isRunning() {
		return
	}
	s.log.Infof("BEGIN server service stop.")
	s.mu.Lock()
	s.srvr.SetKeepAlivesEnabled(false)
	close(s.done)
	s.wg.Wait()
	if s.db != nil {
		s.db.Close()
	}
	if s.redis != nil {
		s.redis.Close()
	}
	s.running = false
	s.mu.Unlock()
	s.log.Infof("END server service stop.")
}

// handleSignals responds to operating system interrupts such as application kills.
func (s *Server) handleSignals() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			s.log.Infof("Server received signal: %v\n", sig)
			s.Shutdown()
			s.log.Infof("Server exiting.")
			os.Exit(0)
		}
	}()
}

// The following methods handle server routes.

// healthHandler handles a client "is the server alive?" request.
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	if s.invalidMethod(w, r, httpGet) {
		return
	}
}

// infoHandler handles a client request for server information.
func (s *Server) infoHandler(w http.ResponseWriter, r *http.Request) {
	if s.invalidHeader(w, r) || s.invalidMethod(w, r, httpGet) || s.invalidAuth(w, r) {
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	b, _ := json.Marshal(
		&struct {
			Options *Options `json:"options"`
		}{
			Options: s.opt,
		})
	w.Write(b)
}

// metricsHandler handles a client request for server statistics.
func (s *Server) metricsHandler(w http.ResponseWriter, r *http.Request) {
	if s.invalidHeader(w, r) || s.invalidMethod(w, r, httpGet) || s.invalidAuth(w, r) {
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	mStats := &runtime.MemStats{}
	runtime.ReadMemStats(mStats)
	b, _ := json.Marshal(
		&struct {
			Options *Options          `json:"options"`
			Stats   *Status           `json:"stats"`
			Memory  *runtime.MemStats `json:"memStats"`
		}{
			Options: s.opt,
			Stats:   s.stats,
			Memory:  mStats,
		})
	w.Write(b)
}

// deployHandler handles a client request for deploying a service to the cluster.
func (s *Server) deployHandler(w http.ResponseWriter, r *http.Request) {
	if s.invalidHeader(w, r) || s.invalidMethod(w, r, httpPost) || s.invalidAuth(w, r) {
		return
	}
	reqID := w.Header().Get("X-Request-ID")

	// Get the payload into a request struct.
	var d DeployRequest
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, InvalidBody, http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(b, &d); err != nil {
		http.Error(w, InvalidJSONText, http.StatusBadRequest)
		return
	}
	// Is environment deployable from this server?
	if _, ok := s.opt.Environments[d.Environment]; !ok {
		http.Error(w, InvalidDeployEnv, http.StatusBadRequest)
		return
	}
	// Does the user have auth for deploys against this env?
	if s.authDeployEnvironment(w, r, d.Environment) {
		return
	}
	// Is there an image to deploy in the payload?
	if d.ImageName == "" {
		http.Error(w, InvalidDeployImage, http.StatusBadRequest)
		return
	}
	if d.ImageTag == "" {
		d.ImageTag = DefaultImageTag
	}
	// Format extra meta-data for the deploy.
	var numCont int
	if i, err := strconv.ParseInt(s.opt.Environments[d.Environment]["num_containers"], 10, 32); err == nil {
		numCont = int(i)
	}
	if d.NumCont == 0 {
		numCont = DefaultNumCont
	}
	swarm, _ := strconv.ParseBool(s.opt.Environments[d.Environment]["swarm"])

	// Push the payload into the queue.
	payload := NewDeployRequest(reqID, d.ImageName, d.ImageTag, d.Environment, s.opt.Environments[d.Environment]["env_tag"],
		s.opt.Environments[d.Environment]["etcd_endpoint"], s.opt.Environments[d.Environment]["machine"],
		s.opt.Environments[d.Environment]["metadata_mount"], numCont, s.opt.Environments[d.Environment]["docker_registry"], swarm)
	if _, err := s.redis.RPush(s.opt.RedisKeyQueue, fmt.Sprint(payload)).Result(); err != nil {
		http.Error(w, InvalidDeployCannotQueue, http.StatusServiceUnavailable)
		return
	}
	s.db.QueueDeploy(payload.DeployID, payload.Environment, payload.ImageName, payload.ImageTag)
	w.Write([]byte(fmt.Sprintf(`{"deployID":"%s"}`, reqID)))
}

// statusHandler handles a client request for checking on a previous deploy status.
func (s *Server) statusHandler(w http.ResponseWriter, r *http.Request) {
	if s.invalidHeader(w, r) || s.invalidMethod(w, r, httpGet) || s.invalidAuth(w, r) {
		return
	}

	// Get the ID from the query parameters and perform a lookup.
	_, deployID := filepath.Split(r.URL.Path)
	result, err := s.db.QueryDeploy(deployID)

	// Format the response data.
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`{"id":%s,"error":"%s"}`, deployID, err)))
		return
	}
	b, _ := json.Marshal(result)
	w.Write(b)
}

// initResponseHeader sets up the common http response headers for the return of all json calls.
func (s *Server) initResponseHeader(w http.ResponseWriter) {
	h := w.Header()
	h.Add("Content-Type", "application/json;charset=utf-8")
	h.Add("Date", time.Now().UTC().Format(time.RFC1123Z))
	if s.opt.ServerName != "" {
		h.Add("Server", s.opt.ServerName)
	}
	h.Add("X-Request-ID", createV4UUID())
}

// incrementStats increments the statistics for the request being handled by the server.
func (s *Server) incrementStats(r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stats.IncrRequestStats(r.ContentLength)
	s.stats.IncrRouteStats(r.URL.Path, r.ContentLength)
}

// invalidHeader validates that the header information is acceptable for processing the
// request from the client.
func (s *Server) invalidHeader(w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("Content-Type") != "application/json" ||
		r.Header.Get("Accept") != "application/json" {
		http.Error(w, InvalidMediaType, http.StatusUnsupportedMediaType)
		return true
	}
	return false
}

// invalidMethod validates that the http method is acceptable for processing this route.
func (s *Server) invalidMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		http.Error(w, InvalidMethod, http.StatusMethodNotAllowed)
		return true
	}
	return false
}

// invalidAuth validates that the Authorization token is valid for using the API
func (s *Server) invalidAuth(w http.ResponseWriter, r *http.Request) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.db.ValidAuth(strings.Replace(r.Header.Get("Authorization"), "Bearer ", "", -1)) {
		http.Error(w, InvalidAuthorization, http.StatusUnauthorized)
		return true
	}
	return false
}

// invalidAuth validates that the Authorization token can deploy to the environment.
func (s *Server) authDeployEnvironment(w http.ResponseWriter, r *http.Request, env string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.db.AuthDeployEnv(strings.Replace(r.Header.Get("Authorization"), "Bearer ", "", -1), env) {
		http.Error(w, InvalidEnvAuthorization, http.StatusUnauthorized)
		return true
	}
	return false
}

// isRunning returns a boolean representing whether the server is running or not.
func (s *Server) isRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// requestLogEntry is a datastructure of a log entry for recording server access requests.
type requestLogEntry struct {
	Method        string      `json:"method"`
	URL           *url.URL    `json:"url"`
	Proto         string      `json:"proto"`
	Header        http.Header `json:"header"`
	Body          string      `json:"body"`
	ContentLength int64       `json:"contentLength"`
	Host          string      `json:"host"`
	RemoteAddr    string      `json:"remoteAddr"`
	RequestURI    string      `json:"requestURI"`
	Trailer       http.Header `json:"trailer"`
}

// LogRequest logs the http request information into the logger.
func (s *Server) LogRequest(r *http.Request) {
	var cl int64

	if r.ContentLength > 0 {
		cl = r.ContentLength
	}

	bd, err := ioutil.ReadAll(r.Body)
	if err != nil {
		bd = []byte("Could not parse body")
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(bd)) // We need to set the body back after we read it.

	b, _ := json.Marshal(&requestLogEntry{
		Method:        r.Method,
		URL:           r.URL,
		Proto:         r.Proto,
		Header:        r.Header,
		Body:          string(bd),
		ContentLength: cl,
		Host:          r.Host,
		RemoteAddr:    r.RemoteAddr,
		RequestURI:    r.RequestURI,
		Trailer:       r.Trailer,
	})
	s.log.Infof(`{"request":%s}`, string(b))
}
