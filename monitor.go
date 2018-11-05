package main

import (
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"ipfs-monitor/command"
	"ipfs-monitor/pinner"
	"ipfs-monitor/reporter"
	"ipfs-monitor/signer"
	"ipfs-monitor/verifier"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	ci "github.com/libp2p/go-libp2p-crypto"
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
	bytes, _ := base64.StdEncoding.DecodeString("CAASpwkwggSjAgEAAoIBAQC0C9G/KNVPGgK2chBAfrdH6b8DiGfS4WNWbPhP9qKnYBNVW9ufNqIiMyZUBPmganbB6E93BkCqlMSotkQrdDkAlt19sd6Oi6CZ3PhkpdnMsWCmCzgm+GAf8LH6Xia8SqsQnfW0tkC4Xi6FZRjVBsFiCTaeS5a4K0n6dn65L2urv/+V9Ry0+8kqPX2FUOykd00oHY8h4cdg7kbYnilY4/F772xOHx/qKhauO27W8Co3sqwZS671w6Y7cD0oPvyrKwSZ9d1dEbqT26g0th+o9g2oT8jBzN34jqX9U8N4XL2o0Fjb773Z8j27N1SvVGeioUNvDMV7IR9G69eLtiPJGYNZAgMBAAECggEADxGSK2qSd71YjsZ7H7q7QjSI/RW0gszEUJ5sJd0hfdqno5Q9jFS5Ox2Gzq9f6RIgAFieFfsa/GvZDbm7eNuQTcFSpkt1sf5zoY0B6QKMePo7eYok1/YfrWyqqKaqnUWujYR65PX/8q5HPHjanDGli7vzq0nuQlm1JlY2gu86FrSl7gTZ2TnDiW3V4iC6R+jrcT5oTYZDk/xpfNIAqa/585ghNBdrh319xlw4uA8shU21URhUzitBl7mQwSx6MVangoGQIxter4k5e86E/xpWXZGb4pzHCxN6GfDRaDUHu3Qb/2qh0AyYZMnxNSiqLqnj3Du9uspqWn3Tj+13tq5BkQKBgQDEXKg/7ZL98a5GZmhNhBvkGKThEFabw9xSYjqBifiSw56usjQ52/DeOYRS/kMWlGQ7efvTw4bR+Hp520rBGJ8U8ObVKovUL7KvXWEIPHaPxEPB29IcB3CJr2ACrBkcB2fTY18RE/PLJ4npWx/sxlPuE2R62bgZygtnZOCYmV8uPQKBgQDqupgO8tVFulbF/NsRQheUFcd8Fc+ibkIBeSqxNDdSyrNenqRk0qsLAOucpmoe9QUKmtkINkuWv23IZm0w/IA6eCRsvvALIIkR/wz9xXzDTCzREwjSeImi46KqlJwgLzUdFg5mi7ow+9XWiBB6tIiVVzIGRtu8fvf1afK7HMe3TQKBgQCcuzB4RlKrazqlapwaMzZn69u+4OGgVscG4uy79Lp5urZvzkGtZQZu+g3KiLRX033lk0oUDt8PvXtUxFp4ogRoHJjC0MDnTmMbYjdO8aPYwNksZR7hzHZhD6Bcwa0RGEAhFBIPeZOANkD4CAoFXZAkWUo8XId73i8YCgbTBqhYKQKBgEG3M9MbkTIQVTLg88QFKOzAnDTVNmhXTgtzjMxHviVjsowPBh9kC5btDwmwXY2FM4AgkhqGeOsfdPDiHDfWuV+lOy1m13WGCloLeeuiWqeir/loxtOGA6Ki8GupZ3xrOweFoUp8XAhSuS9ysIpp+MM2wj6Xn/VJ5snCe75+9DsRAoGAbyI01Wrq0mpdmo3npVkklBGqnmdi30JXd0k/zQJrkAPHKBaJqToUY1hYClbTknRl4xg1RgousqUSgMnd1R+/pKrL7GWWqCOc8eIkiupwlRplIL91PhfotLMi51JRyflTXdyRA7dP4YJumVczX2aJOFvhwx4nfbeELc0zMt8GPXI=")
	// //bytes1, err := base64.StdEncoding.DecodeString("CAASpgIwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQC0C9G/KNVPGgK2chBAfrdH6b8DiGfS4WNWbPhP9qKnYBNVW9ufNqIiMyZUBPmganbB6E93BkCqlMSotkQrdDkAlt19sd6Oi6CZ3PhkpdnMsWCmCzgm+GAf8LH6Xia8SqsQnfW0tkC4Xi6FZRjVBsFiCTaeS5a4K0n6dn65L2urv/+V9Ry0+8kqPX2FUOykd00oHY8h4cdg7kbYnilY4/F772xOHx/qKhauO27W8Co3sqwZS671w6Y7cD0oPvyrKwSZ9d1dEbqT26g0th+o9g2oT8jBzN34jqX9U8N4XL2o0Fjb773Z8j27N1SvVGeioUNvDMV7IR9G69eLtiPJGYNZAgMBAAE=")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	content := []byte(`{"node_external_id":"QmdaZVgC8HgzT7Q2ocTbkWmMAAazM91Tsv2vFTrR66HA2T","pinned_files":[{"id":"QmUSRKhr6Wj71JEM1ovMmXfcDh3cxpS51S1EuhHiXw1YVV","size":13222096},{"id":"QmcUHdzKgRrcJrD5Ah46HgBHF7urWDhmAnLKYwcHaLgeGP","size":1211315561},{"id":"QmfM9mxxcY4g1hBAzjyQhM3dkHCNWvYk3wKxGy3zP7W3BQ","size":7780275},{"id":"QmNtGdg5XhyLJvtCbYM3VnZ5yn7b9Vpv3xAN1p8sJndgoT","size":116},{"id":"QmPhT7R5ZKbSkMpCDXs1X5V3deKpmRHbsk2iuzJrwRfrWj","size":4997},{"id":"QmQYJVD8vNY6ty2HQpLoABXTn2eu9kydeLDENKxNvbzVr3","size":3972777},{"id":"QmUNLLsPACCz1vLxQVkXqqLX5R1X345qqfHbsf67hvA3Nn","size":2},{"id":"QmUixEthEdxPRLYAVwbBREZtNUqygQtMGLjvCua9o1gx2j","size":454474170},{"id":"Qmefnd7NPbE99Jg4CWgPFuKPi4twe5LJfbLggxMnveMRtY","size":486656119},{"id":"QmcubeFV7aCBTYZmnyjmLZG6kp8hsSeY7SS1G8urSEnKpP","size":124128820},{"id":"QmfZNtstQp9xiVNU96cpwB2mP3f21jDmyNMmj6wpddUZ2N","size":1048632},{"id":"QmQ7ztxthx46iSpBo3AboTUtcYDXydHW1pgNpRndhKjW1L","size":1142},{"id":"QmRbx5SmWP3GGuNiPDfaRWFbq6BogzZpGYCsy3KDUMj83g","size":164405165},{"id":"QmS4ustL54uo8FzR9455qaxZwuMiUhyvMcX9Ba8nUH4uVv","size":6189},{"id":"QmUKaQwN2ppapUEFhbHsKoVXn2yBRM7mLpu5HQv9am7dB7","size":2303562483},{"id":"QmXdnMbbyxtgjGF5JSPhmYb2SCnFY4h64XS3qtFado5V11","size":127046739},{"id":"QmcNq16Up2NsdsDduyBfN3GauawAZtqapiVtERAnHjvnnX","size":3146665},{"id":"QmPtqckU2TjDS5X1jbsDRjhbnUKNq41MtgX6d4c1mH5yoD","size":872},{"id":"QmRP5R2azHxMTkfTQuTexmzccFRS2ETkdcnyere9BPsWvv","size":13264402},{"id":"QmUz5hSB17FFLcgEPrcqR5wFG6fpbvn6VT8nUrqM7wFoNS","size":2932586},{"id":"QmdFTSVSH9wqyCcQzBgDzUCSQwKzgYGZagK9bh22H3QkPu","size":140771}],"pinning_file_size":0,"available_space":445675515904,"throughput":1574934578,"last_timestamp":1541412743}`)
	privatekey, err := ci.UnmarshalPrivateKey(bytes)
	signature, err := privatekey.Sign(content)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(hex.EncodeToString(signature[:]))

	bytes1, _ := base64.StdEncoding.DecodeString("CAASpgIwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQC0C9G/KNVPGgK2chBAfrdH6b8DiGfS4WNWbPhP9qKnYBNVW9ufNqIiMyZUBPmganbB6E93BkCqlMSotkQrdDkAlt19sd6Oi6CZ3PhkpdnMsWCmCzgm+GAf8LH6Xia8SqsQnfW0tkC4Xi6FZRjVBsFiCTaeS5a4K0n6dn65L2urv/+V9Ry0+8kqPX2FUOykd00oHY8h4cdg7kbYnilY4/F772xOHx/qKhauO27W8Co3sqwZS671w6Y7cD0oPvyrKwSZ9d1dEbqT26g0th+o9g2oT8jBzN34jqX9U8N4XL2o0Fjb773Z8j27N1SvVGeioUNvDMV7IR9G69eLtiPJGYNZAgMBAAE=")
	publickey, _ := ci.UnmarshalPublicKey(bytes1)

	sig, _ := hex.DecodeString("4687811659a7621fe1ac7009d6eddabbe2a373941a94d4e149cd1b2c25203b10bbc0434e622853b23e02637a5046475fa778397f762149bf811ea222adbe206a47dc586a43d4d5508361d416c0ea002c378078322d63beb44f49d5474a1727d54f90f9eac9343a24ca2bb2a33d450af87f4fcefec3ffc3c5230bd0279a64aa518843010d7181e721865426bb709b9d2fb5ad1e0def57ce12fcc4f7e961e7c31e1bf0e7c506a472ea24c39a48142c44086fa2b52338f9298c1a9dcf4772f78a0a945b0a0f9cf15a6acfde6e38d7554186d282ae1b51beca34dee43c95798d4c59051fe67adb1c9b2f938b888ca95e83fe596786d04e5fad4d2fb85ba6c913257b")
	b, _ := publickey.Verify([]byte(`{"node_external_id":"QmdaZVgC8HgzT7Q2ocTbkWmMAAazM91Tsv2vFTrR66HA2T","pinned_files":[{"id":"QmUSRKhr6Wj71JEM1ovMmXfcDh3cxpS51S1EuhHiXw1YVV","size":13222096},{"id":"QmcUHdzKgRrcJrD5Ah46HgBHF7urWDhmAnLKYwcHaLgeGP","size":1211315561},{"id":"QmfM9mxxcY4g1hBAzjyQhM3dkHCNWvYk3wKxGy3zP7W3BQ","size":7780275},{"id":"QmNtGdg5XhyLJvtCbYM3VnZ5yn7b9Vpv3xAN1p8sJndgoT","size":116},{"id":"QmPhT7R5ZKbSkMpCDXs1X5V3deKpmRHbsk2iuzJrwRfrWj","size":4997},{"id":"QmQYJVD8vNY6ty2HQpLoABXTn2eu9kydeLDENKxNvbzVr3","size":3972777},{"id":"QmUNLLsPACCz1vLxQVkXqqLX5R1X345qqfHbsf67hvA3Nn","size":2},{"id":"QmUixEthEdxPRLYAVwbBREZtNUqygQtMGLjvCua9o1gx2j","size":454474170},{"id":"Qmefnd7NPbE99Jg4CWgPFuKPi4twe5LJfbLggxMnveMRtY","size":486656119},{"id":"QmcubeFV7aCBTYZmnyjmLZG6kp8hsSeY7SS1G8urSEnKpP","size":124128820},{"id":"QmfZNtstQp9xiVNU96cpwB2mP3f21jDmyNMmj6wpddUZ2N","size":1048632},{"id":"QmQ7ztxthx46iSpBo3AboTUtcYDXydHW1pgNpRndhKjW1L","size":1142},{"id":"QmRbx5SmWP3GGuNiPDfaRWFbq6BogzZpGYCsy3KDUMj83g","size":164405165},{"id":"QmS4ustL54uo8FzR9455qaxZwuMiUhyvMcX9Ba8nUH4uVv","size":6189},{"id":"QmUKaQwN2ppapUEFhbHsKoVXn2yBRM7mLpu5HQv9am7dB7","size":2303562483},{"id":"QmXdnMbbyxtgjGF5JSPhmYb2SCnFY4h64XS3qtFado5V11","size":127046739},{"id":"QmcNq16Up2NsdsDduyBfN3GauawAZtqapiVtERAnHjvnnX","size":3146665},{"id":"QmPtqckU2TjDS5X1jbsDRjhbnUKNq41MtgX6d4c1mH5yoD","size":872},{"id":"QmRP5R2azHxMTkfTQuTexmzccFRS2ETkdcnyere9BPsWvv","size":13264402},{"id":"QmUz5hSB17FFLcgEPrcqR5wFG6fpbvn6VT8nUrqM7wFoNS","size":2932586},{"id":"QmdFTSVSH9wqyCcQzBgDzUCSQwKzgYGZagK9bh22H3QkPu","size":140771}],"pinning_file_size":0,"available_space":445675515904,"throughput":1574934578,"last_timestamp":1541412743}`), sig)

	fmt.Println(b)
	// p, _ := pem.Decode(bytes)
	// fmt.Println(p)
	// k := new(big.Int)
	// k.SetBytes(bytes)
	// priv := new(ecdsa.PrivateKey)
	// curve := elliptic.P256()
	// priv.PublicKey.Curve = curve
	// priv.D = k
	// priv.PublicKey.X, priv.PublicKey.Y = curve.ScalarBaseMult(k.Bytes())
	// bytespub := elliptic.Marshal(elliptic.P256(), priv.PublicKey.X, priv.PublicKey.Y)
	// fmt.Println(base64.StdEncoding.EncodeToString(bytespub))

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
	b = verifier.Verify("CAASpgIwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQC0C9G/KNVPGgK2chBAfrdH6b8DiGfS4WNWbPhP9qKnYBNVW9ufNqIiMyZUBPmganbB6E93BkCqlMSotkQrdDkAlt19sd6Oi6CZ3PhkpdnMsWCmCzgm+GAf8LH6Xia8SqsQnfW0tkC4Xi6FZRjVBsFiCTaeS5a4K0n6dn65L2urv/+V9Ry0+8kqPX2FUOykd00oHY8h4cdg7kbYnilY4/F772xOHx/qKhauO27W8Co3sqwZS671w6Y7cD0oPvyrKwSZ9d1dEbqT26g0th+o9g2oT8jBzN34jqX9U8N4XL2o0Fjb773Z8j27N1SvVGeioUNvDMV7IR9G69eLtiPJGYNZAgMBAAE=", `{"node_external_id":"QmdaZVgC8HgzT7Q2ocTbkWmMAAazM91Tsv2vFTrR66HA2T","pinned_files":[{"id":"QmUSRKhr6Wj71JEM1ovMmXfcDh3cxpS51S1EuhHiXw1YVV","size":13222096},{"id":"QmcUHdzKgRrcJrD5Ah46HgBHF7urWDhmAnLKYwcHaLgeGP","size":1211315561},{"id":"QmfM9mxxcY4g1hBAzjyQhM3dkHCNWvYk3wKxGy3zP7W3BQ","size":7780275},{"id":"QmNtGdg5XhyLJvtCbYM3VnZ5yn7b9Vpv3xAN1p8sJndgoT","size":116},{"id":"QmPhT7R5ZKbSkMpCDXs1X5V3deKpmRHbsk2iuzJrwRfrWj","size":4997},{"id":"QmQYJVD8vNY6ty2HQpLoABXTn2eu9kydeLDENKxNvbzVr3","size":3972777},{"id":"QmUNLLsPACCz1vLxQVkXqqLX5R1X345qqfHbsf67hvA3Nn","size":2},{"id":"QmUixEthEdxPRLYAVwbBREZtNUqygQtMGLjvCua9o1gx2j","size":454474170},{"id":"Qmefnd7NPbE99Jg4CWgPFuKPi4twe5LJfbLggxMnveMRtY","size":486656119},{"id":"QmcubeFV7aCBTYZmnyjmLZG6kp8hsSeY7SS1G8urSEnKpP","size":124128820},{"id":"QmfZNtstQp9xiVNU96cpwB2mP3f21jDmyNMmj6wpddUZ2N","size":1048632},{"id":"QmQ7ztxthx46iSpBo3AboTUtcYDXydHW1pgNpRndhKjW1L","size":1142},{"id":"QmRbx5SmWP3GGuNiPDfaRWFbq6BogzZpGYCsy3KDUMj83g","size":164405165},{"id":"QmS4ustL54uo8FzR9455qaxZwuMiUhyvMcX9Ba8nUH4uVv","size":6189},{"id":"QmUKaQwN2ppapUEFhbHsKoVXn2yBRM7mLpu5HQv9am7dB7","size":2303562483},{"id":"QmXdnMbbyxtgjGF5JSPhmYb2SCnFY4h64XS3qtFado5V11","size":127046739},{"id":"QmcNq16Up2NsdsDduyBfN3GauawAZtqapiVtERAnHjvnnX","size":3146665},{"id":"QmPtqckU2TjDS5X1jbsDRjhbnUKNq41MtgX6d4c1mH5yoD","size":872},{"id":"QmRP5R2azHxMTkfTQuTexmzccFRS2ETkdcnyere9BPsWvv","size":13264402},{"id":"QmUz5hSB17FFLcgEPrcqR5wFG6fpbvn6VT8nUrqM7wFoNS","size":2932586},{"id":"QmdFTSVSH9wqyCcQzBgDzUCSQwKzgYGZagK9bh22H3QkPu","size":140771}],"pinning_file_size":0,"available_space":445675515904,"throughput":1574934578,"last_timestamp":1541412743}`, "4687811659a7621fe1ac7009d6eddabbe2a373941a94d4e149cd1b2c25203b10bbc0434e622853b23e02637a5046475fa778397f762149bf811ea222adbe206a47dc586a43d4d5508361d416c0ea002c378078322d63beb44f49d5474a1727d54f90f9eac9343a24ca2bb2a33d450af87f4fcefec3ffc3c5230bd0279a64aa518843010d7181e721865426bb709b9d2fb5ad1e0def57ce12fcc4f7e961e7c31e1bf0e7c506a472ea24c39a48142c44086fa2b52338f9298c1a9dcf4772f78a0a945b0a0f9cf15a6acfde6e38d7554186d282ae1b51beca34dee43c95798d4c59051fe67adb1c9b2f938b888ca95e83fe596786d04e5fad4d2fb85ba6c913257b")
	fmt.Println(b)
}
