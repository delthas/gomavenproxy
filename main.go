package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/textproto"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
	"gopkg.in/yaml.v2"
)

type config struct {
	Port int `yaml:"port"`
	Ftp  struct {
		Host      string `yaml:"host"`
		Port      int    `yaml:"port"`
		Timeout   int    `yaml:"timeout"`
		MavenPath string `yaml:"maven_path"`
	} `yaml:"ftp"`
}

func fatal(f string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, f, args...)
	os.Exit(1)
}

func e(f string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, f+"\n", args...)
}

func main() {
	configPath := flag.String("config", "gomavenproxy.yml", "config file path")
	flag.Parse()

	f, err := os.Open(*configPath)
	if err != nil {
		fatal("error: open config file: %s", err.Error())
		return
	}

	var config config
	err = yaml.NewDecoder(f).Decode(&config)
	if err != nil {
		fatal("error: decode config file: %s", err.Error())
		return
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		user, pass, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", "Basic realm=\"FTP Maven Repository Proxy\"")
			w.WriteHeader(401)
			return
		}

		c, err := ftp.Dial(net.JoinHostPort(config.Ftp.Host, strconv.Itoa(config.Ftp.Port)), ftp.DialWithTimeout(time.Duration(config.Ftp.Timeout)*time.Second))
		if err != nil {
			e("error: connect to ftp: %s", err.Error())
			w.WriteHeader(500)
			return
		}

		err = c.Login(user, pass)
		if err != nil {
			e("error: login to ftp: %s", err.Error())
			w.WriteHeader(403)
			return
		}
		defer c.Quit()

		path := strings.TrimLeft(r.URL.Path, "/")
		if path == "" {
			w.WriteHeader(400)
			return
		}
		path = config.Ftp.MavenPath + path

		if r.Method == "PUT" {
			if r.Body == http.NoBody {
				w.WriteHeader(400)
				return
			}
			if err := c.Stor(path, r.Body); err != nil {
				if ee, ok := err.(*textproto.Error); ok && ee.Code == 550 && len(path) > 1 {
					// no such file or folder, try to create parent directories
					// reader has not been read yet
					i := 1
					for {
						j := strings.IndexRune(path[i:], '/')
						if j == -1 {
							break
						}
						i = j + i + 1
						err = c.MakeDir(path[:i])
						if err != nil {
							if ee, ok := err.(*textproto.Error); !ok || ee.Code != 550 {
								w.WriteHeader(400)
								return
							}
						}
					}
					if err := c.Stor(path, r.Body); err != nil {
						e("error: stor file to ftp at path %s: %s", path, err)
						w.WriteHeader(500)
						return
					}
				} else {
					e("error: stor file to ftp at path %s: %s", path, err)
					w.WriteHeader(500)
					return
				}
			}
			w.WriteHeader(200)
			return
		}
		if r.Method == "GET" {
			r, err := c.Retr(path)
			if err != nil {
				if ee, ok := err.(*textproto.Error); ok && ee.Code == 550 {
					w.WriteHeader(404)
					return
				}
				w.WriteHeader(400)
				return
			}

			w.WriteHeader(200)
			_, err = io.Copy(w, r)
			if err != nil {
				e("copying FTP file to HTTP response: %v", err)
			}
			return
		}

		e("unhandled HTTP method: %v", r.Method)
		w.WriteHeader(400)
		return
	})
	fmt.Println("listening on localhost:" + strconv.Itoa(config.Port))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.Port), nil))
}
