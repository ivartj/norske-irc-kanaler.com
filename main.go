package main

import (
	"github.com/ivartj/norske-irc-kanaler.com/args"
	"github.com/ivartj/norske-irc-kanaler.com/web"
	"github.com/ivartj/norske-irc-kanaler.com/irssilog"
	"github.com/ivartj/norske-irc-kanaler.com/sched"
	"net"
	"net/http"
	"net/http/fcgi"
	"html/template"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path"
	"log"
	"time"
)

var mainConfFilename string = ""

const (
	mainName		= "norske-irc-kanaler.com"
	mainVersion		= "1.0"
)

func mainUsage(out io.Writer) {
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintf(out,  "  %s CONFIGURATION_FILE\n", mainName)
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Description:")
	fmt.Fprintln(out, "  Serves website that inspects and advertises")
	fmt.Fprintln(out, "  IRC channels.")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Options:")
	fmt.Fprintln(out, "  -h, --help  Prints help message")
	fmt.Fprintln(out, "  --version   Prints version")
	fmt.Fprintln(out, "")
}

func mainArgs() {

	tok := args.NewTokenizer(os.Args)

	for tok.Next() {
		arg := tok.Arg()
		isOption := tok.IsOption()

		switch {
		case isOption:
			switch arg {
			case "-h", "--help":
				mainUsage(os.Stdout)
				os.Exit(0)
			case "--version":
				fmt.Printf("%s version %s\n", mainName, mainVersion)
				os.Exit(0)
			}
		case !isOption:
			if mainConfFilename != "" {
				mainUsage(os.Stderr)
				os.Exit(1)
			}
			mainConfFilename = arg
		}
	}

	if tok.Err() != nil {
		fmt.Fprintf(os.Stderr, "Error on processing arguments: %s.\n", tok.Err().Error())
		os.Exit(1)
	}

	if mainConfFilename == "" {
		mainUsage(os.Stderr)
		os.Exit(1)
	}
}

func mainChangeDirectory() {
	dir := path.Dir(mainConfFilename)
	err := os.Chdir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to change to directory '%s': %s.\n", dir, err.Error())
		os.Exit(1)
	}
}

func mainOpenLog(cfg *conf) {
	if(cfg.LogPath() == "") {
		return
	}
	f, err := os.Create(cfg.LogPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file '%s': %s\n", cfg.LogPath(), err.Error())
		os.Exit(1)
	}
	mw := io.MultiWriter(f, os.Stderr)
	log.SetOutput(mw)
}

type mainContext struct {
	auth *auth
	conf *conf
	site *web.Site
	db *sql.DB
}

func mainNewContext(cfg *conf) *mainContext {

	db, err := sql.Open("sqlite3", cfg.DatabasePath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open database: %s.\n", err.Error())
		os.Exit(1)
	}

	err = dbInit(db, cfg.AssetsPath() + "/sql")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s.\n", err.Error())
	}

	tpl, err := web.NewTemplate().ParseFiles(cfg.AssetsPath() + "/templates.html")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse template: %s.\n", err.Error())
		os.Exit(1)
	}
	site := web.NewSite(db, tpl)

	ctx := &mainContext{
		conf: cfg,
		site: site,
		db: db,
		auth: &auth{},
	}

	ctx.site.SetFieldMap(map[string]interface{}{
		"site-title" : cfg.WebsiteTitle(),
		"site-description" : cfg.WebsiteDescription(),
		"page-title" : "",
		"page-messages" : []template.HTML{},
		"admin" : false,
	})

	auth := ctx.auth

	ctx.HandlePage("/", indexPage)
	ctx.HandlePage("/submit", submitPage)
	ctx.HandleDir("/static/", assetsDir)
	ctx.HandleDir("/info/", infoDir)
	ctx.HandlePage("/login", loginPage)
	ctx.HandlePage("/logout", auth.Guard(logoutPage))
	ctx.HandlePage("/admin", auth.Guard(adminPage))
	ctx.HandlePage("/approve", auth.Guard(approvePage))
	ctx.HandlePage("/exclude", auth.Guard(excludePage))
	ctx.HandlePage("/edit", auth.Guard(editPage))
	ctx.HandlePage("/delete", auth.Guard(deletePage))

	return ctx
}

func (ctx *mainContext) HandlePage(path string, pageFn func(*page, *http.Request)) {
	ctx.site.HandlePage(path, pageHandler(ctx, pageFn))
}

func (ctx *mainContext) HandleDir(path string, pageFn func(*page, *http.Request)) {
	ctx.site.HandleDir(path, pageHandler(ctx, pageFn))
}

func mainServeSite(ctx *mainContext) {
	var err error
	switch ctx.conf.ServeMethod() {
	case "http":
		err = http.ListenAndServe(":" + fmt.Sprint(ctx.conf.HttpPort()), ctx.site)
	case "fcgi":
		l, err := net.Listen("unix", ctx.conf.FastcgiPath())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create unix socket for FastCGI: %s.\n", err.Error())
			os.Exit(1)
		}
		err = fcgi.Serve(l, ctx.site)
	}
	log.Fatalf("Error serving site: %s.\n", err.Error())
}

func mainGatherChannelStatuses(ctx *mainContext) {

	scheduler := sched.New()

	for _, method := range ctx.conf.ChannelStatusGatheringMethods() {

		interval, err := time.ParseDuration(method.Interval)
		if err != nil {
			log.Fatalf("Failed to parse the interval given for method '%s': %s", method.Method, err.Error())
		}

		initialTime := time.Now()
		if method.InitialTime != "" {
			initialTime, err = sched.Next(method.InitialTime)
			if err != nil {
				log.Fatalf("Failed to parse the initial time given for method '%s': %s", method.Method, err.Error())
			}
			
		}

		var do func() = nil

		switch method.Method {

		case "irssi-logs":
			do = func() {
				mainIrssiLogs(ctx)
			}

		}
		if do == nil {
			log.Fatalf("Unrecognized method, '%s'", method.Method)
		}

		scheduler.Schedule(do, initialTime, interval)
	}

	go scheduler.Run()
}

func mainIrssiLogs(ctx *mainContext) {

	tx, err := ctx.db.Begin()
	if err != nil {
		log.Fatalf("Failed to initiate transaction: %s", err.Error())
	}
	defer tx.Rollback()

	networks, err := dbGetNetworks(tx)
	if err != nil {
		log.Fatalf("Database error on retrieving networks: %s", err.Error())
	}

	networknames := map[string][]string{}
	for _, network := range networks {
		networknames[network.network] = network.servers
	}
	for network, _ := range ctx.conf.IrssiLogsNetworkNames() {
		names, ok := networknames[network]
		if !ok {
			names = []string{}
		}
		networknames[network] = append(names, ctx.conf.IrssiLogsNetworkNames()[network]...)
	}

	chs, err := dbGetApprovedChannels(tx, 0, 9999)
	if err != nil {
		log.Fatalf("Database error on retrieving channels: %s", err.Error())
	}

	logctx := irssilog.New(ctx.conf.IrssiLogsPath(), networknames)

	for _, ch := range chs {

		cs, err := logctx.ChannelStatus(ch.Name(), ch.Network())
		if err != nil {
			log.Printf("Error retrieving '%s@%s' status from log reading: %s", ch.Name(), ch.Network(), err.Error())
			continue
		}

		if cs.Time.After(ch.CheckTime()) {
			err  = dbUpdateStatus(tx, ch.Name(), ch.Network(), cs.NumUsers, cs.Topic, "irssi-logs", "", cs.Time)
			if err != nil {
				log.Fatalf("Error updating status for '%s@%s': %s", ch.Name(), ch.Network(), err.Error())
			}
		}

	}

	err = tx.Commit()
	if err != nil {
		log.Fatalf("Error on committing status updates from Irssi log reading: %s", err.Error())
	}
}

func main() {
	cfg := confNew()
	mainArgs()
	mainChangeDirectory()
	err := cfg.ParseFile(mainConfFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse configuration file: %s.\n", err.Error())
		os.Exit(1)
	}
	mainOpenLog(cfg)
	ctx := mainNewContext(cfg)

	mainGatherChannelStatuses(ctx)

	if cfg.IRCBotEnable() {
		go channelCheckLoop(ctx)
	}
	mainServeSite(ctx)
}

