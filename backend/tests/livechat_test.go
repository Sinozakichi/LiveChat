package tests

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/spf13/pflag"
)

var livechatOpts = godog.Options{
	Output: colors.Colored(os.Stdout),
	Format: "pretty",
}

// 注意：不再使用 godog.BindCommandLineFlags，避免與 multiroom_test.go 衝突

func TestFeatures(t *testing.T) {
	pflag.Parse()
	livechatOpts.Paths = []string{"../../features"}

	status := godog.TestSuite{
		Name:                 "LiveChat",
		ScenarioInitializer:  InitializeScenario,
		TestSuiteInitializer: InitializeTestSuite,
		Options:              &livechatOpts,
	}.Run()

	if status != 0 {
		t.Fatalf("非零狀態碼: %d", status)
	}
}

func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
		// 在所有測試前執行的設置
	})

	ctx.AfterSuite(func() {
		// 在所有測試後執行的清理
	})
}
