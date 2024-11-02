package main

import (
	"bufio"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

var (
	DBPath             = flag.String("db", "./db/database-auth.sqlite3", "sqlLite database path")
	AllowedIPsFilePath = flag.String("allowed-ips-file", "/var/nginx/allowed-ips.conf", "nginx allowed ips file path")
	Port               = flag.String("port", "82", "web server port running on")
	Host               = flag.String("host", "0.0.0.0", "web server host running on")
	UpstreamServer     = flag.String("upstream", "", "upstream server witch should get new authorized ip. seperated by ','")
	AdminToken         = flag.String("admin-token", "", "admin token which will be used to create users and all upstreams")
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

	if len(*AdminToken) < 2 {
		fatalErrLog(logger, "-admin-token is not set, It's required!!", nil)
	}

	upstreams := slices.DeleteFunc(strings.Split(*UpstreamServer, ","), func(e string) bool {
		return e == ""
	})

	gin.SetMode(gin.ReleaseMode)

	server := gin.Default()

	db, err := sql.Open("sqlite3", *DBPath)
	if err != nil {
		fatalErrLog(logger, "couldn't open db file", err)
	}

	defer db.Close()
	_, err = db.Exec("create table if not exists users (token TEXT UNIQUE NOT NULL , username TEXT, last_ip TEXT, limitation INTEGER DEFAULT 1 NOT NULL, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL, updated_at TIMESTAMP)")
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

		var user User
		isExist, err := getUser(db, logger, &user, token)
		if !isExist {
			if err != nil {
				logger.Error("error in validating token", "Details", err)
				c.String(http.StatusInternalServerError, "server logs")
				return
			}

			c.String(http.StatusUnauthorized, "401")
			return
		}

		// call upstreams
		for _, upstream := range upstreams {
			err = callUpstream(upstream, ip, *AdminToken, &user, logger)
			if err != nil {
				logger.Error("couldn't call upstream!", "Upstream", upstream, "Details", err)
			}
		}

		isAlreadyAllowed, err := isIpInAllowedIpFile(ip, *AllowedIPsFilePath, logger)
		if err != nil {
			c.String(http.StatusInternalServerError, "server logs")
			return
		}

		if isAlreadyAllowed {
			c.String(http.StatusOK, "already added")
			return
		}

		err = allowIp(db, &user, ip, logger)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		c.String(http.StatusOK, "added")

		return
	})

	server.POST("/update-upstreams", func(c *gin.Context) {
		// send users
		users, err := getUsers(db)
		if err != nil {
			logger.Error("couldn't get users!", "Details", err)
			c.String(http.StatusInternalServerError, "server logs")
			return
		}

		for _, user := range users {
			for _, upstream := range upstreams {
				err = callUpstream(upstream, "", *AdminToken, &user, logger)
				if err != nil {
					logger.Error("couldn't update upstream!", "Upstream", upstream, "Details", err)
				}
			}
		}

		c.String(http.StatusOK, "registered all")

		return
	})

	server.GET("/register-user", func(c *gin.Context) {
		ip := c.Request.URL.Query().Get("ip")
		token := c.Request.URL.Query().Get("token")
		username := c.Request.URL.Query().Get("username")
		limitation := c.Request.URL.Query().Get("limitation")
		adminToken := c.Request.URL.Query().Get("adminToken")

		if adminToken == "" || token == "" {
			c.String(http.StatusBadRequest, "Bad Request")
			return
		}

		if adminToken != *AdminToken {
			c.String(http.StatusUnauthorized, "401")
			return
		}

		var user User
		isExist, err := getUser(db, logger, &user, token)
		if err != nil {
			c.String(http.StatusInternalServerError, "server logs")
			return
		}

		if !isExist {
			user.Token = token
			user.Username = sql.NullString{String: username, Valid: true}
			if len(limitation) > 0 {
				if user.Limitation, err = strconv.Atoi(limitation); err != nil {
					logger.Error("limitation must be a number!", "UserInput", limitation)
					c.String(http.StatusInternalServerError, "server logs")
					return
				}
			} else {
				user.Limitation = 1
			}
			err = insertUser(db, logger, &user)
			if err != nil {
				c.String(http.StatusInternalServerError, "server logs")
				return
			}
			logger.Debug("added new user", "Token", user.Token, "Username", user.Username.String)
		}

		if len(ip) > 6 {
			isAlreadyAllowed, err := isIpInAllowedIpFile(ip, *AllowedIPsFilePath, logger)
			if err != nil {
				c.String(http.StatusInternalServerError, "server logs")
				return
			}

			if isAlreadyAllowed {
				c.String(http.StatusOK, "already added")
				return
			}

			err = allowIp(db, &user, ip, logger)
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
		}

		c.String(http.StatusOK, "added")
	})

	logger.Info("listening on => " + *Host + ":" + *Port)
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

func callUpstream(upstream string, ip string, adminToken string, user *User, logger *slog.Logger) error {
	base := url.URL{}

	base.Host = upstream
	base.Scheme = "http"
	base.Path += "/register-user"

	params := url.Values{}
	params.Add("ip", ip)
	params.Add("token", user.Token)
	params.Add("username", user.Username.String)
	params.Add("limitation", strconv.Itoa(user.Limitation))
	params.Add("adminToken", adminToken)
	base.RawQuery = params.Encode()

	client := http.Client{
		Timeout: 3 * time.Second,
	}
	res, err := client.Get(base.String())
	if err != nil {
		return errors.New("error making http request: " + err.Error())
	}

	logger.Debug("upstream %s: status code: %d\n", upstream, res.StatusCode)

	return nil
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func writeLines(lines []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

func getUser(db *sql.DB, logger *slog.Logger, user *User, token string) (bool, error) {
	stmt, err := db.Prepare("select * from users where token = ?")
	if err != nil {
		logger.Error("error in getting user", "Details", err)
		return false, errors.New("error in getting user")
	}

	defer stmt.Close()

	err = stmt.QueryRow(token).Scan(&user.Token, &user.Username, &user.LastIp, &user.Limitation, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return false, nil
	}

	return true, nil
}

func getUsers(db *sql.DB) ([]User, error) {
	stmt, err := db.Prepare("select * from users")
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.Token, &user.Username, &user.LastIp, &user.Limitation, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return users, err
		}
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return users, err
	}

	return users, nil
}

func updateUserLastIp(db *sql.DB, logger *slog.Logger, user *User, ip string) error {
	currentTime := time.Now()
	stmt, err := db.Prepare("UPDATE users SET last_ip = ?, updated_at = ? WHERE token = ?")
	if err != nil {
		logger.Error("couldn't update last login!", "Details", err)
		return err
	}

	defer stmt.Close()
	_, err = stmt.Exec(ip, currentTime.Format("2006-01-02 15:04:05"), user.Token)
	if err != nil {
		logger.Error("couldn't update last login!", "Details", err)
		return err
	}

	return nil
}

func insertUser(db *sql.DB, logger *slog.Logger, user *User) error {
	stmt, err := db.Prepare("INSERT INTO users(token,username,last_ip, limitation) VALUES(?, ?, NULL, ?);")
	if err != nil {
		logger.Error("couldn't create user!", "Details", err)
		return err
	}

	defer stmt.Close()
	_, err = stmt.Exec(user.Token, user.Username, user.Limitation)
	if err != nil {
		logger.Error("couldn't create user!", "Details", err)
		return err
	}

	return nil
}

func addIpToAllowedList(logger *slog.Logger, allowedIpFilePath string, ip string, user *User) error {
	allowedIpsFileLines, err := readLines(allowedIpFilePath)
	if err != nil {
		logger.Error("couldn't read allowed ips file!", "Details", err)
		return err
	}

	currentTime := time.Now()
	newLine := "allow " + ip + "; #(" + user.Token + ") ------ " + currentTime.String()

	allowedIpsFileLines = slices.Insert(allowedIpsFileLines, 0, newLine)

	var resultLines []string
	var userAccessTimes = 0
	for _, line := range allowedIpsFileLines {
		if !strings.Contains(line, "("+user.Token+")") {
			resultLines = append(resultLines, line)
			continue
		}

		userAccessTimes += 1
		if user.Limitation >= userAccessTimes {
			resultLines = append(resultLines, line)
		}
	}

	err = writeLines(resultLines, allowedIpFilePath)
	if err != nil {
		logger.Error("couldn't save allowed ips file!", "Details", err)
		return err
	}
	return nil
}

func reloadNginx(logger *slog.Logger) error {
	cmd := exec.Command("nginx", "-s", "reload")
	_, err := cmd.Output()

	if err != nil {
		logger.Error("couldn't reload nginx!", "Details", err)
		return err
	}

	return nil
}

func allowIp(db *sql.DB, user *User, ip string, logger *slog.Logger) error {
	var err error

	err = updateUserLastIp(db, logger, user, ip)
	if err != nil {
		return errors.New("server logs")
	}

	err = addIpToAllowedList(logger, *AllowedIPsFilePath, ip, user)
	if err != nil {
		return errors.New("server logs")
	}

	err = reloadNginx(logger)
	if err != nil {
		return errors.New("server logs")
	}

	logger.Debug("allow " + ip + " with token: " + user.Token)

	return nil
}

func isIpInAllowedIpFile(ip string, allowedIpFilePath string, logger *slog.Logger) (bool, error) {
	allowedIpsFileLines, err := readLines(allowedIpFilePath)
	if err != nil {
		logger.Error("couldn't read allowed ips file!", "Details", err)
		return false, err
	}

	for _, line := range allowedIpsFileLines {
		if strings.Contains(line, ip) {
			return true, nil
		}
	}
	return false, nil
}
