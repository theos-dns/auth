package main

import (
	"flag"
	"github.com/tufanbarisyildirim/gonginx/config"
	"github.com/tufanbarisyildirim/gonginx/dumper"
	"github.com/tufanbarisyildirim/gonginx/parser"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"strings"
)

var (
	ForwardTo          = flag.String("to", "", "host address where authorized requests froward to (port will not be changed!)")
	Port               = flag.String("port", "443,80", "port listen to, seperated by ',' like: 80,443,1080 also can be range like 8080-8090, or combination of both")
	AllowedIPsFilePath = flag.String("allowed-ips-file", "/var/nginx/allowed-ips.conf", "host listen to")
	NginxConfFilePath  = flag.String("nginx-conf-file", "/etc/nginx/nginx.conf", "host listen to")
	help               = flag.Bool("help", false, "Display help message")
)

func main() {
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	slog.SetLogLoggerLevel(slog.LevelDebug)

	if len(*AllowedIPsFilePath) < 2 {
		slog.Error("-allowed-ips-file path is not set!")
		os.Exit(1)
	}

	if len(*NginxConfFilePath) < 2 {
		slog.Error("-nginx-conf-file path is not set!")
		os.Exit(1)
	}

	if len(*ForwardTo) < 2 {
		slog.Error("-to must be an ip address or host!")
		os.Exit(1)
	}

	ports := slices.DeleteFunc(strings.Split(*Port, ","), func(e string) bool {
		return e == ""
	})

	p, err := parser.NewParser(*NginxConfFilePath)
	if err != nil {
		slog.Error("nginx parser failed!", "Details", err)
		os.Exit(1)
	}
	conf, err := p.Parse()
	if err != nil {
		slog.Error("nginx config file is not valid!")
		os.Exit(1)
	}

	for _, port := range ports {
		if strings.Contains(port, "-") {
			pRange := slices.DeleteFunc(strings.Split(port, "-"), func(e string) bool {
				return e == ""
			})
			if !isNumber(pRange[0]) || !isNumber(pRange[1]) {
				slog.Error("defined port in not correct, please check your input!", "Your input", pRange)
				os.Exit(1)
			}
			start := pRange[0]
			end := pRange[1]

			appendServerBlock(conf, *AllowedIPsFilePath, start+"-"+end, *ForwardTo)

		} else {
			if !isNumber(port) {
				slog.Error("defined port in not correct, please check your input!", "Your input", port)
				os.Exit(1)
			}
			appendServerBlock(conf, *AllowedIPsFilePath, port, *ForwardTo)

		}

	}

	slog.Debug("Generated config ", "content", dumper.DumpConfig(conf, dumper.IndentedStyle))

	err = dumper.WriteConfig(conf, dumper.IndentedStyle, false)
	if err != nil {
		slog.Error("couldn't save config file!")
		os.Exit(1)
	}

}

func appendServerBlock(conf *config.Config, includePath string, port string, proxyPass string) {
	for i := 0; i < len(conf.Directives); i++ {
		if conf.Directives[i].GetName() == "stream" {
			block := conf.Directives[i].GetBlock()
			realBlock := block.(*config.Block)

			includeDirective := &config.Directive{
				Name:       "include",
				Parameters: []string{includePath},
			}
			listenDirective := &config.Directive{
				Name:       "listen",
				Parameters: []string{port},
			}
			proxyPassDirective := &config.Directive{
				Name:       "proxy_pass",
				Parameters: []string{proxyPass + ":$server_port"},
			}

			newBlock := &config.Block{
				Directives: []config.IDirective{
					includeDirective,
					listenDirective,
					proxyPassDirective,
				},
			}
			newDirective := &config.Directive{
				Name:  "server",
				Block: newBlock,
			}

			realBlock.Directives = append(realBlock.Directives, newDirective)

		}

	}
}

func isNumber(inp string) bool {
	if _, err := strconv.Atoi(inp); err == nil {
		return true
	}
	return false
}
