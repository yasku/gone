package cli

import (
	"testing"
)

func TestOsqueryRunnerIsAvailable(t *testing.T) {
	o := NewOsqueryRunner()

	if !o.IsAvailable() {
		t.Skip("osquery not installed, skipping")
	}
}

func TestOsqueryRunnerNotAvailable(t *testing.T) {
	o := NewOsqueryRunner()

	origLookup := whichCache["osqueryi"]
	defer func() {
		if origLookup != "" {
			whichCache["osqueryi"] = origLookup
		} else {
			delete(whichCache, "osqueryi")
		}
	}()

	delete(whichCache, "osqueryi")

	if o.IsAvailable() {
		t.Error("expected osqueryi to not be available after cache clear")
	}
}

func TestOsqueryQueryNotAvailable(t *testing.T) {
	o := NewOsqueryRunner()

	origLookup := whichCache["osqueryi"]
	defer func() {
		if origLookup != "" {
			whichCache["osqueryi"] = origLookup
		} else {
			delete(whichCache, "osqueryi")
		}
	}()

	delete(whichCache, "osqueryi")

	_, err := o.Query("SELECT * FROM processes")
	if err == nil {
		t.Error("expected error when osquery not available")
	}
}

func TestOsqueryQuerySimple(t *testing.T) {
	if !IsAvailable("osqueryi") {
		t.Skip("osquery not installed, skipping")
	}

	o := NewOsqueryRunner()

	results, err := o.Query("SELECT name FROM processes LIMIT 1")
	if err != nil {
		t.Fatalf("OsqueryRunner.Query failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("expected at least one result")
	}

	for _, row := range results {
		if row["name"] == "" {
			t.Error("expected non-empty name in result")
		}
	}
}

func TestOsqueryGetApps(t *testing.T) {
	if !IsAvailable("osqueryi") {
		t.Skip("osquery not installed, skipping")
	}

	o := NewOsqueryRunner()

	results, err := o.GetApps()
	if err != nil {
		t.Fatalf("OsqueryRunner.GetApps failed: %v", err)
	}

	for _, row := range results {
		if row["name"] == "" && row["path"] == "" {
			t.Error("expected name or path in app result")
		}
	}
}

func TestOsqueryGetProcesses(t *testing.T) {
	if !IsAvailable("osqueryi") {
		t.Skip("osquery not installed, skipping")
	}

	o := NewOsqueryRunner()

	results, err := o.GetProcesses()
	if err != nil {
		t.Fatalf("OsqueryRunner.GetProcesses failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("expected at least one process result")
	}

	for _, row := range results {
		if row["name"] == "" {
			t.Error("expected non-empty name in process result")
		}
	}
}

func TestOsqueryGetStartupItems(t *testing.T) {
	if !IsAvailable("osqueryi") {
		t.Skip("osquery not installed, skipping")
	}

	o := NewOsqueryRunner()

	_, err := o.GetStartupItems()
	if err != nil {
		t.Logf("GetStartupItems returned error (may be normal): %v", err)
	}
}

func TestOsqueryGetBrowserPlugins(t *testing.T) {
	if !IsAvailable("osqueryi") {
		t.Skip("osquery not installed, skipping")
	}

	o := NewOsqueryRunner()

	_, err := o.GetBrowserPlugins()
	if err != nil {
		t.Logf("GetBrowserPlugins returned error (may be normal): %v", err)
	}
}

func TestOsqueryGetNetworkConnections(t *testing.T) {
	if !IsAvailable("osqueryi") {
		t.Skip("osquery not installed, skipping")
	}

	o := NewOsqueryRunner()

	results, err := o.GetNetworkConnections()
	if err != nil {
		t.Logf("GetNetworkConnections returned error (may be normal): %v", err)
	}

	for _, row := range results {
		if row["local_address"] == "" && row["remote_address"] == "" {
			t.Error("expected addresses in network connection result")
		}
	}
}

func TestOsqueryGetOpenPorts(t *testing.T) {
	if !IsAvailable("osqueryi") {
		t.Skip("osquery not installed, skipping")
	}

	o := NewOsqueryRunner()

	results, err := o.GetOpenPorts()
	if err != nil {
		t.Logf("GetOpenPorts returned error (may be normal): %v", err)
	}

	for _, row := range results {
		if row["port"] == "" {
			t.Error("expected port in open ports result")
		}
	}
}

func TestOsqueryResultParsing(t *testing.T) {
	if !IsAvailable("osqueryi") {
		t.Skip("osquery not installed, skipping")
	}

	o := NewOsqueryRunner()

	results, err := o.Query("SELECT 1 AS num, 'test' AS str")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}

	row := results[0]
	if row["num"] != "1" {
		t.Errorf("expected num=1, got %s", row["num"])
	}
	if row["str"] != "test" {
		t.Errorf("expected str=test, got %s", row["str"])
	}
}
