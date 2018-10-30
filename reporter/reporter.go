package reporter

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"ipfs-monitor/command"
	"ipfs-monitor/pinner"
	"log"
	"net/http"
	"os"
	"strconv"
)

var Report_URL string

var stdlog, errlog *log.Logger

type Request struct {
	Data      *RequestData `json:"data"`
	Signature string       `json:"signature"`
}

type RequestData struct {
	NodeExternalID string `json:"node_external_id"`
	PinnedFiles    []Item `json:"pinned_files"`
	AvailableSpace uint64 `json:"available_space"`
	Throughput     uint64 `json:"throughput"`
	LastTimestamp  uint64 `json:"last_timestamp"`
}

type Item struct {
	ID   string `json:"id"`
	Size uint64 `json:"size"`
}

type Response struct {
	PinHash          []string `json:"pin_hash"`
	CurrentTimestamp uint64   `json:"current_timestamp"`
}

func init() {
	stdlog = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	errlog = log.New(os.Stderr, "", log.Ldate|log.Ltime)
}

func Report() ([]byte, error) {
	stdlog.Println("Prepare information of IPFS node for reporting status to server...")
	node_external_id, err := command.GetPeerID()
	if err != nil {
		errlog.Println("Get peer ID failed, error: ", err)
		return nil, err
	}
	keys, sizes, err := command.GetPinedList()
	if err != nil {
		errlog.Println("Get pined file list failed, error: ", err)
		return nil, err
	}
	items := make([]Item, len(keys))
	for i, key := range keys {
		size := sizes[key]
		items[i] = Item{ID: key, Size: size}
	}
	available_space, err := command.GetFreeSpace()
	if err != nil {
		errlog.Println("Get free space failed, error: ", err)
		return nil, err
	}
	throughput, err := command.GetThroughput()
	if err != nil {
		errlog.Println("Get throughput failed, error: ", err)
		return nil, err
	}
	timestampstr, err := readTimestamp()
	if err != nil {
		errlog.Println("Read timestamp failed, error: ", err)
		return nil, err
	}
	timestamp, _ := strconv.ParseUint(timestampstr, 10, 64)
	request := &Request{Data: &RequestData{NodeExternalID: node_external_id, PinnedFiles: items, AvailableSpace: available_space, Throughput: throughput, LastTimestamp: timestamp}, Signature: ""}
	requestJson, err := json.Marshal(request)
	stdlog.Println("Ready for report IPFS node status: ", string(requestJson))
	responseJson, err := doBytesPost(Report_URL, requestJson)
	if err != nil {
		errlog.Println("Report status to server failed, error: ", err)
		return nil, err
	}
	stdlog.Println("Sending status successful, retrieving response from server: ", string(responseJson))
	var response Response
	if err := json.NewDecoder(bytes.NewReader(responseJson)).Decode(&response); err != nil {
		errlog.Println("Decode response from server failed, error: ", err)
		return nil, err
	}
	if writeTimestamp(strconv.FormatUint(response.CurrentTimestamp, 10)) != nil {
		errlog.Println("Write timestamp failed, error: ", err)
		return nil, err
	}
	pinner.PinAsync(response.PinHash)
	return requestJson, nil

}

func doBytesPost(url string, data []byte) ([]byte, error) {

	body := bytes.NewReader(data)
	request, err := http.NewRequest("POST", url, body)
	if err != nil {
		//errlog.Println("http.NewRequest,[err=%s][url=%s]", err, url)
		return []byte(""), err
	}
	request.Header.Set("Connection", "Keep-Alive")
	var resp *http.Response
	resp, err = http.DefaultClient.Do(request)
	if err != nil {
		//errlog.Println("http.Do failed,[err=%s][url=%s]", err, url)
		return []byte(""), err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//errlog.Println("http.Do failed,[err=%s][url=%s]", err, url)
		return []byte(""), err
	}
	return b, err
}

func readTimestamp() (string, error) {
	timestamp_path, err := command.GetRepoPath()
	if err != nil {
		return "", err
	}
	content, err := ioutil.ReadFile(timestamp_path + "/monitor_timestamp")
	if err != nil {
		if os.IsNotExist(err) {
			content = []byte("0")
		} else {
			return "", err
		}
	}
	return string(content), nil
}

func writeTimestamp(timestamp string) error {
	timestamp_path, err := command.GetRepoPath()
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(timestamp_path+"/monitor_timestamp", []byte(timestamp), 0644)
	if err != nil {
		return err
	} else {
		return nil
	}
}
