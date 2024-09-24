package main

import (
	"database/sql"
	"flag"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	DBPath             = flag.String("db", "./db/database-auth.sqlite3", "sqlLite database path")
	AllowedIPsFilePath = flag.String("allowed-ips-file", "/var/nginx/allowed-ips.conf", "nginx allowed ips file path")
	Port               = flag.String("port", "82", "web server port running on")
	Host               = flag.String("host", "0.0.0.0", "web server host running on")
	help               = flag.Bool("help", false, "Display help message")
)

func main() {
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	logLevel := &slog.LevelVar{}
	logLevel.Set(slog.LevelDebug)

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)

	ex, err := os.Executable()
	if err != nil {
		fatalErrLog(logger, "couldn't get current path", err)
	}
	exPath := filepath.Dir(ex)

	if len(*DBPath) < 2 {
		*DBPath = exPath + "/db/database-auth.sqlite3"
	}

	if len(*AllowedIPsFilePath) < 2 {
		fatalErrLog(logger, "-allowed-ips-file path is not set!!", nil)
	}

	server := gin.Default()

	db, err := sql.Open("sqlite3", *DBPath)
	if err != nil {
		fatalErrLog(logger, "couldn't open db file", err)
	}

	defer db.Close()
	_, err = db.Exec("create table if not exists users (token text, username text, last_ip text, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL, updated_at TIMESTAMP)")
	if err != nil {
		fatalErrLog(logger, "couldn't create users table", err)
	}

	server.LoadHTMLGlob(exPath + "/templates/**/*.html")

	server.GET("/", Index())

	server.GET("/tap-in", func(c *gin.Context) {
		ip := c.Request.URL.Query().Get("ip")
		token := c.Request.URL.Query().Get("token")

		if ip == "" || token == "" {
			c.String(http.StatusBadRequest, "Bad Request")
			return
		}

		// Check if token is valid
		stmt, err := db.Prepare("select last_ip from users where token = ?")
		if err != nil {
			logger.Error("error in validating token", "Details", err)
			return
		}

		defer stmt.Close()

		var lastIp string
		err = stmt.QueryRow(token).Scan(&lastIp)
		if err != nil {
			logger.Warn("", "Details", err)
		}

		if err != nil {
			c.String(http.StatusUnauthorized, "401")
			return
		}

		input, err := ioutil.ReadFile(*AllowedIPsFilePath)
		if err != nil {
			logger.Error("couldn't read allowed ips file!", "Details", err)
			return
		}

		var ipAlreadyExist bool = false
		if strings.Contains(string(input), ip) {
			ipAlreadyExist = true
		}
		if ipAlreadyExist {
			c.String(http.StatusOK, "already added")
			return
		}

		// Update db, last_ip and updated_at
		currentTime := time.Now()
		stmtUpdateLogin, err := db.Prepare("UPDATE users SET last_ip = ?, updated_at = ? WHERE token = ?")
		if err != nil {
			logger.Error("couldn't update last login!", "Details", err)
			return
		}

		defer stmtUpdateLogin.Close()
		_, err = stmtUpdateLogin.Exec(ip, currentTime.Format("2006-01-02 15:04:05"), token)
		if err != nil {
			logger.Error("couldn't update last login!", "Details", err)
			return
		}

		// Add ip to allowed-ips.conf
		output := "allow " + ip + "; #" + token + " ------ " + currentTime.String() + "\n" + string(input)
		err = ioutil.WriteFile(*AllowedIPsFilePath, []byte(output), 0644)
		if err != nil {
			logger.Error("couldn't save allowed ips file!", "Details", err)
			return
		}

		// Reload nginx
		cmd := exec.Command("nginx", "-s", "reload")
		cmdOutput, err := cmd.Output()
		logger.Debug(string(cmdOutput))

		if err != nil {
			logger.Error("couldn't reload nginx!", "Details", err)
			return
		}

		logger.Debug("added new ip =>")
		logger.Debug("allow " + ip + " with token: " + token)

		c.String(http.StatusOK, "added")

		return

	})

	err = server.Run(*Host + ":" + *Port)
	fatalErrLog(logger, "Err in starting server!", err)
}

func Index() gin.HandlerFunc {
	return func(context *gin.Context) {
		context.HTML(http.StatusOK, "index.html", nil)
	}
}

func fatalErrLog(logger *slog.Logger, msg string, err error) {
	if err != nil {
		logger.Error(msg, "Details", err)

	} else {
		logger.Error(msg)
	}
	os.Exit(1)
}
