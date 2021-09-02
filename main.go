package main

import (
	"fmt"
	"log"
	"os"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	xormadapter "github.com/casbin/xorm-adapter/v2"
	_ "github.com/go-sql-driver/mysql"
)

type Request struct {
	Subject      string
	TargetObject string
	Action       string
}

// AddRules ポリシーへのロールの紐付け
func AddRules(e *casbin.Enforcer, rules [][]string) {
	if ok, err := e.AddNamedPolicies("p", rules); err != nil {
		log.Fatalf("error: AddNamedPolicies: %s", err)
	} else if !ok {
		log.Printf("the rule already exists")
	}
}

// CheckAuthRules アクセス要求判定
func CheckAuthRules(e *casbin.Enforcer, req Request) {
	if ok, err := e.Enforce(req.Subject, req.TargetObject, req.Action); err != nil {
		log.Fatalf("error: Enforce: %s", err)
	} else if ok {
		fmt.Println("permit")
	} else {
		fmt.Println("deny")
	}
}

func main() {
	// Initialize a Xorm adapter with MySQL database.
	const dsn = "root:root@tcp(db:3306)/casbinsample"

	a, err := xormadapter.NewAdapter("mysql", dsn, true)
	if err != nil {
		log.Fatalf("error: adapter: %s", err)
	}

	// モデル定義の読み込み
	f, err := os.Open("model.conf")
	if err != nil {
		log.Fatalf("failed to os.Open(): %s", err)
	}
	data := make([]byte, 1024)
	b, err := f.Read(data)
	if err != nil {
		log.Fatalf("failed to f.Read(): %s", err)
	}

	m, err := model.NewModelFromString(string(data[:b]))
	if err != nil {
		log.Fatalf("error: model: %s", err)
	}

	// Casbinコンストラクタ
	e, err := casbin.NewEnforcer(m, a)
	if err != nil {
		log.Fatalf("error: enforcer: %s", err)
	}

	// アクセスルール設定
	rules := [][]string{
		{"role1", "data1", "read"},
		{"role1", "data2", "write"},
		{"role2", "data2", "write"},
		{"role2", "data3", "read"},
	}

	AddRules(e, rules)

	// ユーザーへのロール紐付け
	if ok, err := e.AddRoleForUser("user1", "role1"); err != nil {
		log.Fatalf("failed to e.AddRoleForUser(): %s", err)
	} else if !ok {
		log.Printf("the user already has the role")
	}

	// アクセス要求判定
	req := Request{
		Subject:      "user1",
		TargetObject: "data3",
		Action:       "read",
	}

	CheckAuthRules(e, req)

	// 対象ユーザーのロール確認
	roles, err := e.GetRolesForUser(req.Subject)
	if err != nil {
		log.Fatalf("faild to e.GetRolesForUser(): %s", err)
	}
	fmt.Printf("%s has rules like this: %v\n", req.Subject, roles)
	fmt.Println("All Done")
}
