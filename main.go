package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"github.com/chai2010/winsvc"
	_ "github.com/denisenkom/go-mssqldb"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Test1C struct {
	CorrelationId      string `xml:"corID" json:"CorrelationId"`
	PaymentDocumentID  string `xml:"paymentDoc" json:"PaymentDocumentID"`
	PaymentOrderNumber string `xml:"paymentOrder" json:"PaymentOrderNumber"`
}

type App struct {
	Mode                   string
	Server                 string
	Database               string
	InputPort              string
	UseUsernameAndPassword bool
	Username               string
	Password               string
}

var app = App{}
var db *sql.DB
var connString string
var (
	appPath string

	flagServiceName = flag.String("service-name", "golang-test-winsvc", "Set service name")
	flagServiceDesc = flag.String("service-desc", "golang-test windows service", "Set service description")

	flagServiceInstall   = flag.Bool("service-install", false, "Install service")
	flagServiceUninstall = flag.Bool("service-remove", false, "Remove service")
	flagServiceStart     = flag.Bool("service-start", false, "Start service")
	flagServiceStop      = flag.Bool("service-stop", false, "Stop service")
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage:
  go_exec_sp [options]...

Options:
`)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "%s\n", `
Example:
  # run go_exec_sp server
  $ go build -o go_exec_sp.exe main.go
  $ go_exec_sp.exe

  # install go_exec_sp as windows service
  $ go_exec_sp.exe -service-install

  # start/stop go_exec_sp service
  $ go_exec_sp.exe -service-start
  $ go_exec_sp.exe -service-stop

  # remove go_exec_sp service
  $ go_exec_sp.exe -service-remove

  # help
  $ go_exec_sp.exe -h
`)
	}

	var err error
	if appPath, err = winsvc.GetAppPath(); err != nil {
		log.Fatal(err)
	}
	if err := os.Chdir(filepath.Dir(appPath)); err != nil {
		log.Fatal(err)
	}
}

func main() {
	Log(2, "App started", nil)
	flag.Parse()

	// install service
	if *flagServiceInstall {
		if err := winsvc.InstallService(appPath, *flagServiceName, *flagServiceDesc); err != nil {
			Log(2, fmt.Sprintf("installService(%s, %s): ", *flagServiceName, *flagServiceDesc), err)
		}
		fmt.Printf("Done\n")
		return
	}

	// remove service
	if *flagServiceUninstall {
		if err := winsvc.RemoveService(*flagServiceName); err != nil {
			Log(2, "removeService ", err)
		}
		fmt.Printf("Done\n")
		return
	}

	// start service
	if *flagServiceStart {
		if err := winsvc.StartService(*flagServiceName); err != nil {
			Log(2, "startService ", err)
		}
		fmt.Printf("Done\n")
		return
	}

	// stop service
	if *flagServiceStop {
		if err := winsvc.StopService(*flagServiceName); err != nil {
			Log(2, "stopService ", err)
		}
		fmt.Printf("Done\n")
		return
	}

	// run as service
	if !winsvc.InServiceMode() {
		Log(2, "main: runService ", nil)
		if err := winsvc.RunAsService(*flagServiceName, StartServer, StopServer, false); err != nil {
			//log.Fatalf("svc.Run: %v\n", err)
			Log(1, "svc.Run:", err)
		}
		return
	}
	StartServer()
}

func StartServer() {
	Log(2, "StartServer", nil)
	initApp(&app)
	var errdb error
	db, errdb = sql.Open("mssql", connString)
	if errdb != nil {
		Log(0, " Error open db:", errdb)
	}
	ctx := context.Background()
	errdb = db.PingContext(ctx)
	if errdb != nil {
		Log(0, " Error open db:", errdb)
	}
	_, err := db.Query("select @@version")
	if err != nil {
		Log(0, " Error db query:", err)
	}
	db.Close()
	r := http.NewServeMux()
	r.HandleFunc("/goexec", createObj)
	Log(0, "Error open listening", (http.ListenAndServe(":"+app.InputPort, r)))
}

func StopServer() {
	Log(2, "StopServer", nil)
}

func initApp(a *App) {
	dir1, _ := os.Getwd()
	file, err := os.Open(filepath.Join(dir1, "config.json"))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(a)
	if err != nil {
		log.Fatal(err)
	}
	connString = "server=" + a.Server + ";database=" + a.Database
	if app.UseUsernameAndPassword {
		connString = connString + ";user id=" + a.Username + ";password=" + a.Password
	}
	connString = connString + ";encrypt=disable;"
}

func createObj(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		Log(2, "createObj", nil)
		w.Header().Set("Content-Type", "application/json")
		var reqObj Test1C
		var err error
		result := "ok"
		ctx := context.Background()
		db, err = sql.Open("mssql", connString)
		if err != nil {
			Log(1, "sql.Open error", err)
			result = "error"
		}
		defer db.Close()
		if db == nil {
			Log(1, "createObj: db is nill", errors.New("createObj: db is nill"))
			result = "error"
		}
		err = db.PingContext(ctx)
		if err != nil {
			Log(1, "createObj PingContext error", err)
			result = "error"
		}
		err = json.NewDecoder(r.Body).Decode(&reqObj)
		if err != nil {
			Log(1, "Decode json error", err)
			result = "error"
		}
		var enc []byte
		enc, err = xml.MarshalIndent(reqObj, " ", "  ")
		if err != nil {
			Log(1, "Encode XML error", err)
			result = "error"
		}
		_, err = db.ExecContext(ctx, "_tee_testProc"+"'"+string(enc)+"'")
		if err != nil {
			Log(1, "", err)
			result = "error"
		}
		err = json.NewEncoder(w).Encode(result)
		if err != nil {
			Log(1, "Create response error", err)
		}
	}

}
func Log(level int, message string, inputErr error) {
	t := time.Now().Format("20060102")
	dir1, _ := os.Getwd()
	f, err := os.OpenFile(filepath.Join(dir1, t+".log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
	var pref string
	switch level {
	case 0:
		pref = "PANIC"
	case 1:
		pref = "ERROR"
	default:
		pref = "INFO"
	}
	if app.Mode == "dev" {
		if level <= 1 {
			log.Fatal(pref, message, inputErr)
		} else {
			log.Println(pref, message)
		}
	} else {
		log.Println(pref, message, inputErr)
	}
}
