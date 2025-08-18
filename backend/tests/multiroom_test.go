package tests

import (
	"flag"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var multiroomOpts = godog.Options{
	Output: colors.Colored(os.Stdout),
	Format: "pretty",
}

func init() {
	godog.BindCommandLineFlags("godog.", &multiroomOpts)
}

func TestMultiRoomFeatures(t *testing.T) {
	flag.Parse()
	multiroomOpts.Paths = []string{"../../features/multiroom_chat.feature"}

	status := godog.TestSuite{
		Name:                 "MultiRoom Chat Features",
		ScenarioInitializer:  InitializeMultiRoomScenario,
		TestSuiteInitializer: InitializeMultiRoomTestSuite,
		Options:              &multiroomOpts,
	}.Run()

	if status != 0 {
		t.Fatalf("非零退出狀態: %d", status)
	}
}

func InitializeMultiRoomTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
		// 在整個測試套件開始前執行的代碼
	})
}
