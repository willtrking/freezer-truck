package main

import (
	"log"

	"github.com/dgraph-io/badger/v2"
	"github.com/kelseyhightower/envconfig"
	"github.com/melbahja/goph"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

type Config struct {
	ScheduleCron string
	Host string
	Username string
	Password string
	PrivateKey string
	InsecureKnownHosts bool
	BadgerPath string
	ConcurrencyLimit int
	RemoteFileRoot string
	LocalFileRoot string
	TempFileRoot string
}

func (c Config) Validate()  {

	if c.ScheduleCron == "" {
		log.Fatal("Missing cron schedule, set TRUCK_SCHEDULECRON")
	}

	if c.Host == "" {
		log.Fatal("Missing SSH host, set TRUCK_HOST")
	}

	if c.Username == "" {
		log.Fatal("Missing SSH username, set TRUCK_USERNAME")
	}

	if c.Password == "" && c.PrivateKey == "" {
		log.Fatal("Missing SSH password / private key, set TRUCK_USERNAME or TRUCK_PRIVATEKEY")
	}

	if c.BadgerPath == "" {
		log.Fatal("Missing badger path, set TRUCK_BADGERPATH")
	}

	if c.ConcurrencyLimit <= 0 {
		log.Fatal("Missing concurrency limit, set TRUCK_CONCURRENCYLIMIT")
	}

	if c.RemoteFileRoot == "" {
		log.Fatal("Missing remote file root, set TRUCK_REMOTEFILEROOT")
	}

	if c.LocalFileRoot == "" {
		log.Fatal("Missing local file root, set TRUCK_LOCALFILEROOT")
	}

	if c.TempFileRoot == "" {
		log.Fatal("Missing temp file root, set TRUCK_TEMPFILEROOT")
	}
}

func (c Config) Auth() goph.Auth  {

	if c.PrivateKey != "" {
		return goph.Key(c.PrivateKey, c.Password)
	}

	return goph.Password(c.Password)
}

func (c Config) HostKeyCallback() (ssh.HostKeyCallback, error) {

	if c.InsecureKnownHosts {
		return ssh.InsecureIgnoreHostKey(), nil
	}

	return goph.DefaultKnownHosts()
}



func main() {


	var config Config

	err := envconfig.Process("truck", &config)

	if err != nil {
		log.Fatal(err)
	}

	config.Validate()

	logger, err := configZap()
	if err != nil {
		log.Fatal(err)
	}

	hostKeyCallback, err := config.HostKeyCallback()
	if err != nil {
		log.Fatal(err)
	}

	client, err := goph.NewConn(config.Username, config.Host, config.Auth(), hostKeyCallback)
	if err != nil {
		log.Fatal(err)
	}

	defer client.Close()


	db, err := badger.Open(badger.DefaultOptions(config.BadgerPath))
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	opts := RetrievalOptions{
		RemoteFileRoot:config.RemoteFileRoot,
		LocalFileRoot:config.LocalFileRoot,
		TempFileRoot:config.TempFileRoot,
		SshClient: client.Conn,
		BadgerDb: db,
		ConcurrencyLimiter:NewLimiter(config.ConcurrencyLimit),
	}

	scheduler := cron.New()

	_, err = scheduler.AddFunc(config.ScheduleCron, func() {
		err := RetrieveNewFiles(logger, opts)
		if err != nil {
			logger.Error("Failed to retrieve new files", zap.Error(err))
		}
	})

	if err != nil {
		log.Fatal(err)
	}


	scheduler.Run()

}
