package server

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/composer22/coreos-deploy/etcd2"
	"github.com/composer22/docker-deploy-server/db"
	"github.com/composer22/docker-deploy-server/logger"
	"github.com/spf13/viper"
	redis "gopkg.in/redis.v3"
)

// DeployService handles requests for deployoemnt into one or more machines for an environment (dev, qa etc.)
type deployService struct {
	opt   *Options        // Server options.
	db    *db.DBConnect   // Database connection.
	redis *redis.Client   // Redis connection for the queue.
	done  chan bool       // Channel to receive signal to shutdown now.
	log   *logger.Logger  // Application log for events.
	wg    *sync.WaitGroup // Wait group for the run.
}

// NewDeployService is a factory function that returns a new deployment service instance.
func NewDeployService(o *Options, s *db.DBConnect, r *redis.Client, d chan bool, l *logger.Logger, wg *sync.WaitGroup) *deployService {
	return &deployService{
		opt:   o,
		db:    s,
		redis: r,
		done:  d,
		log:   l,
		wg:    wg,
	}
}

// Run is the main event loop that processes deploy requests.
func (d *deployService) Run() {
	d.wg.Add(1)

	defer d.wg.Done()
	for {
		select {
		case <-d.done: // Server signal quit
			d.db.Close()
			d.redis.Close()
			return
		default:
			result, err := d.redis.RPop(d.opt.RedisKeyQueue).Result()
			if err != nil && err.Error() != "redis: nil" {
				d.log.Errorf(err.Error())
				break
			}
			if result == "" {
				break
			}
			b := []byte(result)
			var r DeployRequest
			if err := json.Unmarshal(b, &r); err != nil {
				d.log.Errorf(err.Error())
				break
			}
			d.deploy(&r)
		}
		time.Sleep(time.Duration(d.opt.RedisPollInt) * time.Second)
	}
}

// Deploy will handle details to deploy a request to a machine or cluster of machines.
func (d *deployService) deploy(r *DeployRequest) {
	// Log the start to the DB.
	row, err := d.db.QueryDeploy(r.DeployID)
	if err != nil {
		msg := "Could not get Deploy from MySQL for ID: "
		d.log.Errorf("ERR: %s %s\n%s\n", msg, r.DeployID, err)
		return
	}
	log := row.Log
	msg := "Started Deploy."
	log += fmt.Sprintln(msg)
	d.db.UpdateDeploy(r.DeployID, db.Started, msg, log)

	// Get last deploy image tag for this Docker image in the request.
	lastImageDeployKey := fmt.Sprintf("%s:%s", d.opt.RedisKeyLastDeploy, r.Environment)
	lastImageTag, err := d.redis.HGet(lastImageDeployKey, r.ImageName).Result()
	if err != nil && err.Error() != "redis: nil" {
		msg = "Unable to access redis server for last deploy validation."
		log += fmt.Sprintf("ERR: %s\n%s\n", msg, err)
		d.db.UpdateDeploy(r.DeployID, db.Failed, msg, log)
		return
	}

	// Create working temp directory for this deploy.
	msg = "Creating working temp directory for this deploy."
	log += fmt.Sprintln(msg)
	d.db.UpdateDeploy(r.DeployID, db.Started, msg, log)
	tempDirectory, err := d.createTempDirectory(d.opt.TempPath, r.Environment, r.ImageName)
	if err != nil {
		msg = "Unable to create temporary work directory on server."
		log += fmt.Sprintf("ERR: %s\n%s\n", msg, err)
		d.db.UpdateDeploy(r.DeployID, db.Failed, msg, log)
		return
	}
	defer os.RemoveAll(tempDirectory)

	// Download the image locally and extract out the docker-compose.yml for the new container.
	msg = "Extracting meta-data from Docker image in registry."
	log += fmt.Sprintln(msg)
	d.db.UpdateDeploy(r.DeployID, db.Started, msg, log)
	cmd := exec.Command("./scripts/download-image.sh", r.ImageTag, r.Registry, r.ImageName, tempDirectory)
	log, err = d.executeCommand(cmd, r, msg, log)
	if err != nil {
		return
	}
	// TODO should we be able to override this in some way.
	// What if it doesn't exist? Should be look in meta area?
	// What if we want to override intentionally: based on image tag? (except latest)
	if _, err := os.Stat(fmt.Sprintf("%s/docker-compose.yml", tempDirectory)); os.IsNotExist(err) {
		msg = "docker-compose.yml file doesn't exist for this launch."
		log += fmt.Sprintf("ERR: %s\n%s\n", msg, err)
		d.db.UpdateDeploy(r.DeployID, db.Failed, msg, log)
		return
	}

	// Download the github meta-data locally for provisioning and extraction.
	msg = "Downloading meta-data from git."
	log += fmt.Sprintln(msg)
	d.db.UpdateDeploy(r.DeployID, db.Started, msg, log)
	cmd = exec.Command("./scripts/download-metadata.sh", d.opt.GitRepo, d.opt.GitRoot, tempDirectory)
	log, err = d.executeCommand(cmd, r, msg, log)
	if err != nil {
		return
	}

	// Extract out the number of containers we need for this environment and app.
	msg = "Extracting number of containers to launch."
	log += fmt.Sprintln(msg)
	d.db.UpdateDeploy(r.DeployID, db.Started, msg, log)

	numCont := r.NumCont
	v := viper.New()
	v.SetConfigName("main")
	v.AddConfigPath(fmt.Sprintf("%s/%s/roles/%s/meta", tempDirectory, d.opt.GitRepo, r.ImageName))
	v.AddConfigPath(fmt.Sprintf("%s/%s/roles/common/meta", tempDirectory, d.opt.GitRepo))
	if err := v.ReadInConfig(); err == nil {
		if kc := v.GetInt(fmt.Sprintf("environments.%s.containers", r.Environment)); kc > 0 {
			numCont = kc
		}
	}

	// Deploy metadata to all machines in the environment.
	msg = "Deploying meta-data."
	log += fmt.Sprintln(msg)
	d.db.UpdateDeploy(r.DeployID, db.Started, msg, log)
	cmd = exec.Command("./scripts/deploy-metadata.sh", r.EnvTag, d.opt.GitRepo, r.MetaMount, tempDirectory)
	log, err = d.executeCommand(cmd, r, msg, log)
	if err != nil {
		return
	}

	// Update the etcd keys in the cluster to this new meta-data (common + app specific only).
	if r.EtcdEndpoint != "" {
		msg := "Deploying etcd2 keys."
		log += fmt.Sprintln(msg)
		d.db.UpdateDeploy(r.DeployID, db.Started, msg, log)
		if msg, err := d.updateEtcd(r, tempDirectory); err != nil {
			log += fmt.Sprintf("ERR: %s\n%s\n", msg, err)
			d.db.UpdateDeploy(r.DeployID, db.Failed, msg, log)
			return
		}
	}

	// Deploy containers to the machines.
	service := strings.Replace(r.ImageName, "-", "_", -1)
	sw := strconv.FormatBool(r.Swarm)
	msg = "Starting up containers."
	log += fmt.Sprintln(msg)
	d.db.UpdateDeploy(r.DeployID, db.Started, msg, log)
	nc := strconv.FormatInt(int64(numCont), 10)
	if lastImageTag == "" {
		lastImageTag = r.ImageTag
	}
	cmd = exec.Command("./scripts/deploy-containers.sh", r.ImageName, r.ImageTag, lastImageTag,
		r.Registry, service, r.Machine, nc, d.opt.Project, sw, tempDirectory)
	log, err = d.executeCommand(cmd, r, msg, log)
	if err != nil {
		return
	}

	// Update redis with the repo, image, image tag of the last deploy.
	_, err = d.redis.HSet(lastImageDeployKey, r.ImageName, r.ImageTag).Result()
	if err != nil {
		msg = "Unable to access redis server to set last deploy image tag."
		log += fmt.Sprintf("ERR: %s\n%s\n", msg, err)
		d.db.UpdateDeploy(r.DeployID, db.Failed, msg, log)
		return
	}

	// Update as success.
	msg = "Containers deployed successfully."
	log += fmt.Sprintf("SUCCESS: %s\n", msg)
	d.db.UpdateDeploy(r.DeployID, db.Success, msg, log)
}

// updateEtcd updates etcd2 keys in the environment.
func (d *deployService) updateEtcd(r *DeployRequest, tempDirectory string) (msg string, err error) {
	etcd2Keys := make(map[string]string)

	// Read in common etcd2 keys from the metadata.
	v := viper.New()
	v.SetConfigName(fmt.Sprintf("%s.etcd2", r.Environment))
	v.AddConfigPath(fmt.Sprintf("%s/%s/roles/common/meta", tempDirectory, d.opt.GitRepo))
	d.readEtcdKeys(v, etcd2Keys)

	// Read in the application etcd2 keys from the metadata.
	v = viper.New()
	v.SetConfigName(fmt.Sprintf("%s.etcd2", r.Environment))
	v.AddConfigPath(fmt.Sprintf("%s/%s/roles/%s/meta", tempDirectory, d.opt.GitRepo, r.ImageName))
	d.readEtcdKeys(v, etcd2Keys)

	// No keys to update. Return.
	if len(etcd2Keys) <= 0 {
		return "", nil
	}

	// Get a connection.
	conn, err := etcd2.NewEtcd2Connect(r.EtcdEndpoint)
	if err != nil {
		msg := "Unable to connect to etcd2 server."
		return msg, err
	}

	// Update the server.
	err = conn.Set(etcd2Keys)
	if err != nil {
		msg := "Unable to perform updates to etcd2 server."
		return msg, err
	}

	return "", nil
}

// readEtcdKeys reads the etcd2 key values from viper into an array of key/values.
func (d *deployService) readEtcdKeys(v *viper.Viper, etcd2Keys map[string]string) {
	if err := v.ReadInConfig(); err == nil {
		var k []map[string]string
		if err := v.Unmarshal(k); err == nil {
			for _, rec := range k {
				etcd2Keys[rec["key"]] = rec["value"]
			}
		}
	}
	return
}

// createTempDirectory creates a unique temporary directory for the deploy process.
func (d *deployService) createTempDirectory(tempPath string, environment string, imageName string) (tempDirectory string, err error) {
	tempDirectory = fmt.Sprintf("%s/%s/%s-%s", tempPath, environment, imageName, randomString(maxRandom))
	if err := os.MkdirAll(tempDirectory, os.ModePerm); err != nil {
		return "", err
	}
	return tempDirectory, nil
}

// executeCommand executes a shell command and log an error if fail.
func (d *deployService) executeCommand(cmd *exec.Cmd, r *DeployRequest, msg string, log string) (string, error) {
	if _, err := execCmd(cmd); err != nil {
		if err.Error() == "Desired container number already achieved\n" {
			return log, nil
		}
		log += fmt.Sprintf("ERR: %s\n%s\n", msg, err)
		d.db.UpdateDeploy(r.DeployID, db.Failed, msg, log)
		return log, err
	}
	return log, nil
}
