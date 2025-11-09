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
	AllowedIPsFilePath = flag.String("allowed-ips-file", "/var/nginx/allowed-ips.conf", "nginx allowed ips file path")
	NginxConfFilePath  = flag.String("nginx-conf-file", "/etc/nginx/nginx.conf", "nginx config file path ")
	ServicesToProtect  = flag.String("protect", "dns-server:53@53,coap:85@5688", "other services that should be protected. Seperated by ','. Structure: {SERVICE_OR_IP}:{SOURCE_PORT}@{DESTINATION_PORT}")
	Resolver           = flag.String("resolver", "127.0.0.53:53", "dns server that resolves protected-services and forward-to hosts")
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

	servicesToProtect := slices.DeleteFunc(strings.Split(*ServicesToProtect, ","), func(e string) bool {
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

	if len(*Resolver) < 2 {
		*Resolver = "127.0.0.53:53"
	}

	setResolverInStream(conf, *Resolver, "3s", "5s")

	for _, port := range ports {
		if strings.Contains(port, "-") {
			pRange := slices.DeleteFunc(strings.Split(port, "-"), func(e string) bool {
				return e == ""
			})
			if !isNumber(pRange[0]) || !isNumber(pRange[1]) {
				slog.Error("defined port in not correct, please check your input!", "Your input", pRange)
				os.Exit(1)
			}
			start, _ := strconv.Atoi(pRange[0])
			end, _ := strconv.Atoi(pRange[1])

			for i := start; i <= end; i++ {
				appendServerBlock(conf, *AllowedIPsFilePath, strconv.Itoa(i), *ForwardTo, strconv.Itoa(i), "tcp,udp")
			}

		} else {
			if !isNumber(port) {
				slog.Error("defined port in not correct, please check your input!", "Your input", port)
				os.Exit(1)
			}
			appendServerBlock(conf, *AllowedIPsFilePath, port, *ForwardTo, port, "tcp,udp")

		}

	}

	for _, service := range servicesToProtect {
		s := strings.Split(service, ":")
		if len(s) != 2 {
			slog.Error("please check your defined protected services", "Your input", service)
			os.Exit(1)
		}
		serviceName, pPorts := s[0], s[1]

		pP := strings.Split(pPorts, "@")

		if len(s) != 2 {
			slog.Error("please check your defined protected services, ports must fallow this structure: {SOURCE_PORT}@{DESTINATION_PORT}", "Your input", pPorts)
			os.Exit(1)
		}

		sourcePort, destPort := pP[0], pP[1]

		if !isNumber(sourcePort) || !isNumber(destPort) {
			slog.Error("please check your defined protected services, ports must be numbers!", "Your input", pPorts)
			os.Exit(1)
		}

		appendServerBlock(conf, *AllowedIPsFilePath, sourcePort, serviceName, destPort, "tcp,udp")
	}

	slog.Debug("Generated config ", "content", dumper.DumpConfig(conf, dumper.IndentedStyle))

	err = dumper.WriteConfig(conf, dumper.IndentedStyle, false)
	if err != nil {
		slog.Error("couldn't save config file!")
		os.Exit(1)
	}

}

func setResolverInStream(conf *config.Config, server string, valid string, timeout string) {
	for i := 0; i < len(conf.Directives); i++ {
		if conf.Directives[i].GetName() == "stream" {
			block := conf.Directives[i].GetBlock()
			realBlock := block.(*config.Block)

			resolver := &config.Directive{
				Name: "resolver",
				Parameters: []config.Parameter{
					{Value: server},
					{Value: "valid=" + valid},
				},
			}
			resolverTimeout := &config.Directive{
				Name:       "resolver_timeout",
				Parameters: []config.Parameter{{Value: timeout}},
			}

			realBlock.Directives = append(realBlock.Directives, resolver, resolverTimeout)
		}
	}
}

func appendServerBlock(conf *config.Config, includePath string, sourcePort string, proxyPass string, destPort string, protocols string) {
	if len(protocols) == 0 {
		protocols = "tcp"
	}

	pros := slices.DeleteFunc(strings.Split(protocols, ","), func(e string) bool {
		return e == ""
	})

	for i := 0; i < len(conf.Directives); i++ {
		if conf.Directives[i].GetName() == "stream" {
			block := conf.Directives[i].GetBlock()
			realBlock := block.(*config.Block)

			for _, p := range pros {
				includeDirective := &config.Directive{
					Name:       "include",
					Parameters: []config.Parameter{{Value: includePath}},
				}

				listenDirective := &config.Directive{
					Name:       "listen",
					Parameters: []config.Parameter{{Value: sourcePort}},
				}

				proxyConnectTimeoutDirective := &config.Directive{
					Name:       "proxy_connect_timeout",
					Parameters: []config.Parameter{{Value: "5s"}},
				}

				proxyTimeoutDirective := &config.Directive{
					Name:       "proxy_timeout",
					Parameters: []config.Parameter{{Value: "60s"}},
				}

				if p == "udp" {
					listenDirective.Parameters = append(listenDirective.Parameters, config.Parameter{Value: "udp"})
				}

				var proxyPassDirective *config.Directive

				if destPort == "" {
					proxyPassDirective = &config.Directive{
						Name:       "proxy_pass",
						Parameters: []config.Parameter{{Value: proxyPass + ":$server_port"}},
					}
				} else {
					proxyPassDirective = &config.Directive{
						Name:       "proxy_pass",
						Parameters: []config.Parameter{{Value: proxyPass + ":" + destPort}},
					}
				}

				newBlock := &config.Block{
					Directives: []config.IDirective{
						includeDirective,
						listenDirective,
						proxyPassDirective,
						proxyConnectTimeoutDirective,
						proxyTimeoutDirective,
					},
				}

				if p == "udp" {
					proxyResponsesDirective := &config.Directive{
						Name:       "proxy_responses",
						Parameters: []config.Parameter{{Value: "0"}},
					}
					newBlock.Directives = append(newBlock.Directives, proxyResponsesDirective)
				}

				newDirective := &config.Directive{
					Name:  "server",
					Block: newBlock,
				}

				realBlock.Directives = append(realBlock.Directives, newDirective)
			}

		}

	}
}

func isNumber(inp string) bool {
	if _, err := strconv.Atoi(inp); err == nil {
		return true
	}
	return false
}
