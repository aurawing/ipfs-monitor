# ipfs-monitor

### Command-Line arguments of ipfs-monitor:

### 1. -cron_expr string
Cron expression for reporting IPFS node status regularly, please refer to [https://godoc.org/github.com/robfig/cron](https://godoc.org/github.com/robfig/cron) for expression details.
### 2. -ipfs_base_url string
Base URL of IPFS API, default is http://127.0.0.1:5001.
### 3. -server_url string
Server URL for reporting status.