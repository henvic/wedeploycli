package integration

import (
	"path/filepath"
	"testing"

	"github.com/wedeploy/cli/tdata"
)

func testAnalyticsReportStatus(status string, t *testing.T) {
	var cmd = &Command{
		Args: []string{"analytics-report", "status"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "mocks/home/bucket/project/container",
	}

	var e = &Expect{
		ExitCode: 0,
		Stdout:   status,
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func testAnalyticsReportEnable(t *testing.T) {
	var cmd = &Command{
		Args: []string{"analytics-report", "enable"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "mocks/home/bucket/project/container",
	}

	var e = &Expect{}
	cmd.Run()
	e.Assert(t, cmd)
}

func testAnalyticsReportDisable(t *testing.T) {
	var cmd = &Command{
		Args: []string{"analytics-report", "disable"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "mocks/home/bucket/project/container",
	}

	var e = &Expect{}
	cmd.Run()
	e.Assert(t, cmd)
}

func testAnalyticsReportReset(t *testing.T) {
	var cmd = &Command{
		Args: []string{"analytics-report", "reset"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "mocks/home/bucket/project/container",
	}

	if len(tdata.FromFile(filepath.Join(GetLoginHome(), ".we_metrics"))) < 200 {
		t.Errorf("Expected .we_metrics file to have at least 200 characters")
	}

	var e = &Expect{}
	cmd.Run()
	e.Assert(t, cmd)

	if len(tdata.FromFile(filepath.Join(GetLoginHome(), ".we_metrics"))) != 0 {
		t.Errorf("Expected .we_metrics file to be truncated")
	}
}

func TestAnalyticsReport(t *testing.T) {
	defer Teardown()
	Setup()

	t.Run("testAnalyticsReportStatus", func(t *testing.T) {
		testAnalyticsReportStatus("disabled", t)
	})

	t.Run("testAnalyticsReportEnable", testAnalyticsReportEnable)

	t.Run("testAnalyticsReportStatus", func(t *testing.T) {
		testAnalyticsReportStatus("enabled", t)
	})

	t.Run("testAnalyticsReportDisable", testAnalyticsReportDisable)

	t.Run("testAnalyticsReportStatus", func(t *testing.T) {
		testAnalyticsReportStatus("disabled", t)
	})

	t.Run("testAnalyticsReportReset", testAnalyticsReportReset)

	t.Run("testAnalyticsReportStatus", func(t *testing.T) {
		testAnalyticsReportStatus("disabled", t)
	})

	t.Run("testAnalyticsReportEnable", testAnalyticsReportEnable)

	if len(tdata.FromFile(filepath.Join(GetLoginHome(), ".we_metrics"))) < 200 {
		t.Errorf("Expected .we_metrics file to have at least 200 characters")
	}
}
