package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"github.com/lumoslabs/dns-healthz/healthz"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	version = "0.1.0"

	app    = kingpin.New("dns-healthz", "A DNS healthz server.").DefaultEnvars()
	config = app.Flag("config", "Path to config file.").Short('C').ExistingFile()
	vlog   = app.Flag("v", "Enable V-leveled logging at the specified level.").Default("0").Int()
	vmod   = app.Flag("vmodule", "glog vmodule settings.").Default("").String()
	probes = app.Flag("probe", "DNS probe definition as json string").Strings()
	prefix = app.Flag("route-prefix", "Route prefix. Routes will be constructed as /<prefix>/<probe name>.").Default("healthz").String()
	listen = app.Flag("listen", "Listen address.").Short('l').Default(":8080").String()
	grace  = app.Flag("grace-timeout", "Graceful shutdown timeout.").Default("30s").Duration()
)

func main() {
	app.Version(version)
	kingpin.MustParse(app.Parse(os.Args[1:]))
	os.Args = []string{os.Args[0],
		"-logtostderr=true",
		"-v=" + fmt.Sprint(*vlog),
		"-vmodule=" + *vmod,
	}
	flag.Parse()

	list := make([]*healthz.Probe, 0, 0)
	for _, p := range *probes {
		pr := &healthz.Probe{}
		if er := json.Unmarshal([]byte(p), pr); er == nil {
			list = append(list, pr)
		} else {
			glog.V(4).Info(er.Error())
		}
	}

	health := new(healthz.Healthz)
	var er error
	if *config != "" {
		health, er = healthz.NewFromConfig(*config)
		if er != nil {
			glog.Fatalf(`Failed to read config %s: %v`, *config, er)
		}
	}
	health.AddProbe(list...)
	if glog.V(4) {
		glog.Info(health.String())
	}

	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())
	e.Logger.SetOutput(os.Stderr)
	e.Logger.SetLevel(log.Lvl(*vlog))

	route := fmt.Sprintf(`/%s/:probe`, *prefix)
	e.GET(route, func(c echo.Context) error {
		p := c.Param("probe")
		status := health.Status(p)
		if status.Error() != nil {
			return c.JSON(http.StatusInternalServerError, status)
		}
		return c.JSON(http.StatusOK, status)
	})

	hctx, hcancel := context.WithCancel(context.Background())
	go func() {
		health.Start(hctx)
		if err := e.Start(*listen); err != nil {
			e.Logger.Info(`Shutting down the server {"error": "%v"}`, err)
		} else {
			e.Logger.Info("Shutting down the server")
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit
	hcancel()
	ctx, cancel := context.WithTimeout(context.Background(), *grace)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
