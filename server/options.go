package server

import (
	"encoding/json"
	"strconv"

	"github.com/spf13/viper"
)

// Options represents parameters that are passed to the application for launching the server.
type Options struct {
	ConfigPath         string                       `json:"configPath"`         // Filepath to the config of the server.
	ConfigPrefix       string                       `json:"configPrefix"`       // Prefix of the config file name of the server.
	ServerName         string                       `json:"serverName"`         // Name of the server.
	Domain             string                       `json:"domain"`             // Domain of the server.
	Hostname           string                       `json:"hostName"`           // Hostname of the server.
	Port               int                          `json:"port"`               // HTTP api port of the server.
	ProfPort           int                          `json:"profPort"`           // The profiler port of the server.
	DSN                string                       `json:"-"`                  // DSN login string to the database.
	RedisHostname      string                       `json:"redisHostname"`      // Hostname of a redis server.
	RedisPort          int                          `json:"redisPort"`          // Port of a redis server.
	RedisPassword      string                       `json:"-"`                  // Password of a redis server.
	RedisDatabase      int                          `json:"redisDatabase"`      // Database of a redis server.
	RedisKeyLastDeploy string                       `json:"redisKeyLastDeploy"` // Redis key for the hash that holds the last deploys.
	RedisKeyQueue      string                       `json:"redisKeyQueue"`      // Redis key for the list that acts as a queue.
	RedisPollInt       int                          `json:"redisPollInt"`       // Redis polling interval for teh queue.
	GitRoot            string                       `json:"gitRoot"`            // Prefix for the git command to access account.
	GitRepo            string                       `json:"gitRepo"`            // Repo name on github that contains app config data.
	Project            string                       `json:"project"`            // Docker-compose project param.
	TempPath           string                       `json:"tempPath"`           // Temp directory for work.
	Debug              bool                         `json:"debugEnabled"`       // Is debugging enabled in the application or server.
	Environments       map[string]map[string]string `json:"environments"`       // Environments for deployment.
}

// Fill in the defaults for the viper configuration.
func (o *Options) SetConfigDefaults(v *viper.Viper) {
	v.SetDefault("server_name", DefaultServerName)
	v.SetDefault("domain", DefaultDomain)
	v.SetDefault("hostname", DefaultHostname)
	v.SetDefault("port", DefaultPort)
	v.SetDefault("profiler_port", 0)
	v.SetDefault("dsn", DefaultDSN)
	v.SetDefault("redis", map[string]string{
		"hostname":        DefaultRedisHost,
		"port":            DefaultRedisPort,
		"password":        "",
		"database":        "",
		"key_last_deploy": DefaultRedisKeyLastDeploy,
		"key_queue":       DefaultRedisKeyQueue,
		"poll_interval":   string(DefaultRedisPollInt),
	})
	v.SetDefault("project", DefaultProject)
	v.SetDefault("temp_path", DefaultTempPath)

	// Add the config path.
	v.SetConfigType("yaml")
	v.SetConfigName(o.ConfigPrefix)
	v.AddConfigPath(DefaultConfigPath)
	if o.ConfigPath != "" {
		v.AddConfigPath(o.ConfigPath)
	}
	v.AddConfigPath(".")
}

// Fill in the options from a viper configuration.
func (o *Options) FillConfig(v *viper.Viper) {
	o.ServerName = v.GetString("server_name")
	o.Domain = v.GetString("domain")
	o.Hostname = v.GetString("hostname")
	o.Port = v.GetInt("port")
	o.ProfPort = v.GetInt("profiler_port")
	o.DSN = v.GetString("dsn")
	o.RedisHostname = v.GetString("redis.hostname")
	o.RedisPort = v.GetInt("redis.port")
	o.RedisPassword = v.GetString("redis.password")
	o.RedisDatabase = v.GetInt("redis.database")
	o.RedisKeyLastDeploy = v.GetString("redis.key_last_deploy")
	o.RedisKeyQueue = v.GetString("redis.key_queue")
	o.RedisPollInt = v.GetInt("redis.poll_interval")
	o.GitRoot = v.GetString("git.root")
	o.GitRepo = v.GetString("git.repo")
	o.Project = v.GetString("project")
	o.TempPath = v.GetString("temp_path")

	o.Environments = make(map[string]map[string]string)
	envs := v.GetStringMap("environments")
	for env, tags := range envs {
		for k, v := range tags.(map[interface{}]interface{}) {
			if _, ok := o.Environments[env]; !ok {
				o.Environments[env] = make(map[string]string)
			}
			var value string
			switch v.(type) {
			case int:
				value = strconv.FormatInt(int64(v.(int)), 10)
			case bool:
				value = strconv.FormatBool(v.(bool))
			default:
				value = v.(string)
			}
			o.Environments[env][k.(string)] = value
		}
	}
}

// String is an implentation of the Stringer interface so the structure is returned as a string
// to fmt.Print() etc.
func (o *Options) String() string {
	b, _ := json.Marshal(o)
	return string(b)
}
