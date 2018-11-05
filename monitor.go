package main

import (
	"flag"
	"fmt"
	"ipfs-monitor/command"
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

var ipfs_base_url = flag.String("ipfs_base_url", "http://127.0.0.1:5001", "Base URL of IPFS API")
var server_url = flag.String("server_url", "http://newtest.mboxone.com/ipfs/public/index.php/index/Call/index", "Server URL for reporting status")
var cron_expr = flag.String("cron_expr", "0 0/30 * * * *", "Cron expression for reporting IPFS node status regularly")
var job_count = flag.Int("job_count", 5, "Thread count of concurrent pinning job, before 1 and 20")

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

func main1() {
	// command.Base_URL = "http://127.0.0.1:5001"

	// sig, err := signer.Sign("12345")
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Println(sig)
	// }
	// bytes, err := base64.StdEncoding.DecodeString("CAASpwkwggSjAgEAAoIBAQC0C9G/KNVPGgK2chBAfrdH6b8DiGfS4WNWbPhP9qKnYBNVW9ufNqIiMyZUBPmganbB6E93BkCqlMSotkQrdDkAlt19sd6Oi6CZ3PhkpdnMsWCmCzgm+GAf8LH6Xia8SqsQnfW0tkC4Xi6FZRjVBsFiCTaeS5a4K0n6dn65L2urv/+V9Ry0+8kqPX2FUOykd00oHY8h4cdg7kbYnilY4/F772xOHx/qKhauO27W8Co3sqwZS671w6Y7cD0oPvyrKwSZ9d1dEbqT26g0th+o9g2oT8jBzN34jqX9U8N4XL2o0Fjb773Z8j27N1SvVGeioUNvDMV7IR9G69eLtiPJGYNZAgMBAAECggEADxGSK2qSd71YjsZ7H7q7QjSI/RW0gszEUJ5sJd0hfdqno5Q9jFS5Ox2Gzq9f6RIgAFieFfsa/GvZDbm7eNuQTcFSpkt1sf5zoY0B6QKMePo7eYok1/YfrWyqqKaqnUWujYR65PX/8q5HPHjanDGli7vzq0nuQlm1JlY2gu86FrSl7gTZ2TnDiW3V4iC6R+jrcT5oTYZDk/xpfNIAqa/585ghNBdrh319xlw4uA8shU21URhUzitBl7mQwSx6MVangoGQIxter4k5e86E/xpWXZGb4pzHCxN6GfDRaDUHu3Qb/2qh0AyYZMnxNSiqLqnj3Du9uspqWn3Tj+13tq5BkQKBgQDEXKg/7ZL98a5GZmhNhBvkGKThEFabw9xSYjqBifiSw56usjQ52/DeOYRS/kMWlGQ7efvTw4bR+Hp520rBGJ8U8ObVKovUL7KvXWEIPHaPxEPB29IcB3CJr2ACrBkcB2fTY18RE/PLJ4npWx/sxlPuE2R62bgZygtnZOCYmV8uPQKBgQDqupgO8tVFulbF/NsRQheUFcd8Fc+ibkIBeSqxNDdSyrNenqRk0qsLAOucpmoe9QUKmtkINkuWv23IZm0w/IA6eCRsvvALIIkR/wz9xXzDTCzREwjSeImi46KqlJwgLzUdFg5mi7ow+9XWiBB6tIiVVzIGRtu8fvf1afK7HMe3TQKBgQCcuzB4RlKrazqlapwaMzZn69u+4OGgVscG4uy79Lp5urZvzkGtZQZu+g3KiLRX033lk0oUDt8PvXtUxFp4ogRoHJjC0MDnTmMbYjdO8aPYwNksZR7hzHZhD6Bcwa0RGEAhFBIPeZOANkD4CAoFXZAkWUo8XId73i8YCgbTBqhYKQKBgEG3M9MbkTIQVTLg88QFKOzAnDTVNmhXTgtzjMxHviVjsowPBh9kC5btDwmwXY2FM4AgkhqGeOsfdPDiHDfWuV+lOy1m13WGCloLeeuiWqeir/loxtOGA6Ki8GupZ3xrOweFoUp8XAhSuS9ysIpp+MM2wj6Xn/VJ5snCe75+9DsRAoGAbyI01Wrq0mpdmo3npVkklBGqnmdi30JXd0k/zQJrkAPHKBaJqToUY1hYClbTknRl4xg1RgousqUSgMnd1R+/pKrL7GWWqCOc8eIkiupwlRplIL91PhfotLMi51JRyflTXdyRA7dP4YJumVczX2aJOFvhwx4nfbeELc0zMt8GPXI=")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// k := new(big.Int)
	// k.SetBytes(bytes)
	// priv := new(ecdsa.PrivateKey)
	// curve := elliptic.P256()
	// priv.PublicKey.Curve = curve
	// priv.D = k
	// priv.PublicKey.X, priv.PublicKey.Y = curve.ScalarBaseMult(k.Bytes())
	// fmt.Printf("X: %d, Y: %d\n", priv.PublicKey.X, priv.PublicKey.Y)

	// h := md5.New()
	// io.WriteString(h, "test123")
	// signhash := h.Sum(nil)
	// fmt.Println(signhash)

	// r, s, err := ecdsa.Sign(rand.Reader, priv, signhash)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// signature := r.Bytes()
	// signature = append(signature, s.Bytes()...)
	// fmt.Printf("Signature: %x\n", signature)

	// verifystatus := ecdsa.Verify(&priv.PublicKey, signhash, r, s)
	// fmt.Println(verifystatus)
}
