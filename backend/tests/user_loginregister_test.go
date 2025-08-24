package tests

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/spf13/pflag"
)

var userLoginRegisterOpts = godog.Options{
	Output: colors.Colored(os.Stdout),
	Format: "pretty",
}

// 注意：移除 godog.BindCommandLineFlags 以避免與其他測試的 flag 衝突

func TestUserLoginRegisterFeatures(t *testing.T) {
	pflag.Parse()
	userLoginRegisterOpts.Paths = []string{"../../features/user_loginregister.feature"}

	status := godog.TestSuite{
		Name:                 "User Login and Register Features",
		ScenarioInitializer:  InitializeUserLoginRegisterScenario,
		TestSuiteInitializer: InitializeUserLoginRegisterTestSuite,
		Options:              &userLoginRegisterOpts,
	}.Run()

	if status != 0 {
		t.Fatalf("非零退出狀態: %d", status)
	}
}

func InitializeUserLoginRegisterTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
		// 在整個測試套件開始前執行的代碼
	})

	ctx.AfterSuite(func() {
		// 在整個測試套件結束後執行的代碼
	})
}
