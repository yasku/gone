package cli

import (
	"encoding/json"
	"fmt"
	"time"
)

type OsqueryRunner struct {
	runner *Runner
	tool   string
}

func NewOsqueryRunner() *OsqueryRunner {
	return &OsqueryRunner{
		runner: NewRunner(30 * time.Second),
		tool:   "osqueryi",
	}
}

func (o *OsqueryRunner) IsAvailable() bool {
	return IsAvailable("osqueryi")
}

type OsqueryResult struct {
	Columns map[string]string `json:"columns"`
	Types   map[string]string `json:"types"`
}

func (o *OsqueryRunner) Query(sql string) ([]map[string]string, error) {
	if !o.IsAvailable() {
		return nil, fmt.Errorf("osqueryi not available")
	}

	args := []string{
		"--json",
		sql,
	}

	var results []map[string]string
	err := o.runner.ExecStream("osqueryi", args, func(line []byte) bool {
		if len(line) == 0 || string(line) == "[]\n" {
			return true
		}

		// Try to parse as array of objects
		if line[0] == '[' {
			var rows []map[string]string
			if err := json.Unmarshal(line, &rows); err == nil {
				results = append(results, rows...)
				return true
			}
		}

		// Try to parse as single object
		var row map[string]string
		if err := json.Unmarshal(line, &row); err == nil && row != nil {
			results = append(results, row)
		}
		return true
	})

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (o *OsqueryRunner) GetApps() ([]map[string]string, error) {
	return o.Query("SELECT name, path FROM apps")
}

func (o *OsqueryRunner) GetProcesses() ([]map[string]string, error) {
	return o.Query("SELECT name, pid, cpu_percent, memory_percent FROM processes ORDER BY cpu_percent DESC LIMIT 20")
}

func (o *OsqueryRunner) GetStartupItems() ([]map[string]string, error) {
	return o.Query("SELECT name, path, status FROM startup_items WHERE status = 'enabled'")
}

func (o *OsqueryRunner) GetBrowserPlugins() ([]map[string]string, error) {
	return o.Query("SELECT name, version FROM browser_plugins")
}

func (o *OsqueryRunner) GetNetworkConnections() ([]map[string]string, error) {
	return o.Query("SELECT process_name, local_address, remote_address, state FROM process_open_sockets WHERE state = 'ESTABLISHED'")
}

func (o *OsqueryRunner) GetOpenPorts() ([]map[string]string, error) {
	return o.Query("SELECT port, process, path FROM listening_ports WHERE port > 0")
}
