package main

import (
	"flag"
	"fmt"
	"ipfs-monitor/command"
	"ipfs-monitor/config"
	"ipfs-monitor/pinner"
	"ipfs-monitor/reporter"
	"ipfs-monitor/signer"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/robfig/cron"
	"github.com/takama/daemon"
)

const (
	name        = "ipfs-monitor"
	description = "Monitor IPFS node and report status to IPHash server."
)

var stdlog, errlog *log.Logger

var ipfs_base_url = flag.String("ipfs_base_url", config.GetCurrentConfig().BaseUrl, "Base URL of IPFS API")
var server_url = &config.GetCurrentConfig().ServerUrl
var cron_expr = &config.GetCurrentConfig().CronExpr
var job_count = &config.GetCurrentConfig().JobCount
var httpTimeout = config.GetHTTPTimeout()

// Service is the daemon service struct
type Service struct {
	daemon.Daemon
}

func init() {
	stdlog = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	errlog = log.New(os.Stderr, "", log.Ldate|log.Ltime)
}

// Manage by daemon commands or run the daemon
func (service *Service) Manage() (string, error) {
	usage := "Usage: ipfs_monitor install | remove | start | stop | status"
	if len(os.Args) == 2 {
		command := os.Args[1]
		switch command {
		case "install":
			return service.Install()
		case "remove":
			return service.Remove()
		case "start":
			return service.Start()
		case "stop":
			return service.Stop()
		case "status":
			return service.Status()
		default:
			return usage, nil
		}
	}
	stdlog.Println("IPFS monitor starting...")
	stdlog.Printf("Use IPFS base URL: %s\n", *ipfs_base_url)
	stdlog.Printf("Use server URL: %s\n", *server_url)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	c := cron.New()
	c.AddFunc(*cron_expr, func() {
		_, err := reporter.Report()
		if err != nil {
			errlog.Println("Abort reporting, waiting for next turn.")
		}
	})
	c.Start()
	pinner.PinService()
	killSignal := <-interrupt
	stdlog.Println("Got signal:", killSignal)
	return "Service exited", nil
}

func main() {
	flag.Parse()
	command.Base_URL = *ipfs_base_url
	reporter.Report_URL = *server_url
	pinner.JobCount = *job_count
	if pinner.JobCount < 1 {
		pinner.JobCount = 1
	}
	if pinner.JobCount > 20 {
		pinner.JobCount = 20
	}
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 100
	http.DefaultTransport.(*http.Transport).ResponseHeaderTimeout = httpTimeout
	signer.Initialize()
	srv, err := daemon.New(name, description)
	if err != nil {
		errlog.Println("Error: ", err)
		os.Exit(1)
	}
	service := &Service{srv}
	status, err := service.Manage()
	if err != nil {
		errlog.Println(status, "\nError: ", err)
		os.Exit(1)
	}

	fmt.Println(status)
}
