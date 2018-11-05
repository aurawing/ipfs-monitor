package command

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/shirou/gopsutil/disk"
)

var Base_URL string

// ID struct for command `ipfs id`
type ID struct {
	ID              string
	PublicKey       string
	Addresses       []string
	AgentVersion    string
	ProtocolVersion string
}

// PinedList struct for command `ipfs pin ls`
type PinedList struct {
	Keys map[string]interface{}
}

// ObjectItem nested struct for command `ipfs object get`
type ObjectItem struct {
	Name string
	Hash string
	Size uint64
}

// Object struct for command `ipfs object get`
type Object struct {
	Links []ObjectItem
	Data  string
}

// Bandwidth struct for command `ipfs stats bw`
type Bandwidth struct {
	TotalIn  uint64
	Totalout uint64
	RateIn   float64
	RateOut  float64
}

// RepoStat struct for command `ipfs repo stat`
type RepoStat struct {
	RepoSize   uint64
	StorageMax uint64
	NumObjects uint64
	RepoPath   string
	Version    string
}

//Pined result struct for command `ipfs pin add`
type PinedResult struct {
	Pins     []string
	Progress uint64
}

// GetPeerID used for get IPFS peer ID
func GetPeerID() (string, error) {
	resp, err := http.Get(Base_URL + "/api/v0/id")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Get peer id failed: %s", resp.Status)
	}
	var result ID
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.ID, nil
}

// GetPinedList used for get pined file list
func GetPinedList() ([]string, map[string]uint64, error) {
	resp, err := http.Get(Base_URL + "/api/v0/pin/ls?type=recursive")
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("Get pined files failed: %s", resp.Status)
	}
	var result PinedList
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, nil, err
	}
	keys := make([]string, len(result.Keys))
	items := make(map[string]uint64)
	i := 0
	for key := range result.Keys {
		size, err := calculateRootSpace(key)
		if err != nil {
			return nil, nil, err
		}
		//items[i] = reporter.Item{ID: k, Size: size}
		keys[i] = key
		items[key] = size
		i++
	}
	return keys, items, nil
}

// func GetPinedSpace(pinedList []string) (uint64, error) {
// 	var space uint64
// 	for _, hash := range pinedList {
// 		rootSpace, err := calculateRootSpace(hash)
// 		if err != nil {
// 			return 0, err
// 		}
// 		space += rootSpace
// 	}
// 	return space, nil
// }

func calculateRootSpace(hash string) (uint64, error) {
	resp, err := http.Get(Base_URL + "/api/v0/object/get?arg=" + hash)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("Calculate space for file %s failed: %s", hash, resp.Status)
	}
	var result Object
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	var space uint64
	for _, item := range result.Links {
		space += item.Size
	}
	if space == 0 {
		space = uint64(len(result.Data))
	}
	return space, nil
}

// GetThroughput used for get throughput
func GetThroughput() (uint64, error) {
	resp, err := http.Get(Base_URL + "/api/v0/stats/bw")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("Get throughput failed: %s", resp.Status)
	}
	var result Bandwidth
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	return result.Totalout, nil
}

// GetFreeSpace used for get free space of repo disk
func GetFreeSpace() (uint64, error) {
	path, err := GetRepoPath()
	if err != nil {
		return 0, err
	}
	stat, err := disk.Usage(path)
	if err != nil {
		return 0, err
	}
	return stat.Free, nil
	// fs := syscall.Statfs_t{}
	// err2 := syscall.Statfs(path, &fs)
	// if err2 != nil {
	// 	return 0, err2
	// }
	// return fs.Bavail * uint64(fs.Bsize), nil
}

func GetRepoPath() (string, error) {
	resp, err := http.Get(Base_URL + "/api/v0/repo/stat")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Get repo stat failed: %s", resp.Status)
	}
	var result RepoStat
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.RepoPath, nil
}

func GetFile(hash string, dst io.Writer, progress func(int64, int64)) error {
	resp, err := http.Get(Base_URL + "/api/v0/get?arg=" + hash)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	fileSizeStr := resp.Header.Get("X-Content-Length")
	if fileSizeStr == "" {
		fileSizeStr = "-1"
	}
	fileSize, err := strconv.ParseInt(fileSizeStr, 10, 64)
	if err != nil {
		fileSize = -1
	}
	var downloadSize int64
	for {
		written, err := io.CopyN(dst, resp.Body, 128*1024)
		if progress != nil {
			downloadSize += written
			progress(downloadSize, fileSize)
		}
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
	}
	return nil
}

func PinFile(hash string) (*PinedResult, error) {
	resp, err := http.Get(Base_URL + "/api/v0/pin/add?arg=" + hash + "&recursive=true&progress=false")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Pin file stat failed: %s", resp.Status)
	}
	var result PinedResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}
