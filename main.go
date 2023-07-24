package main

import (
	"crypto/sha1"
	"crypto/tls"
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/ivartj/norske-irc-kanaler.com/args"
	"github.com/ivartj/norske-irc-kanaler.com/irc"
	"github.com/ivartj/norske-irc-kanaler.com/irssilog"
	"github.com/ivartj/norske-irc-kanaler.com/sched"
	"github.com/ivartj/norske-irc-kanaler.com/util"
	"github.com/ivartj/norske-irc-kanaler.com/web"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

var mainConfFilename string = ""

const (
	mainName    = "norske-irc-kanaler.com"
	mainVersion = "1.0"
)

func mainUsage(out io.Writer) {
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintf(out, "  %s CONFIGURATION_FILE\n", mainName)
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
	if cfg.LogPath() == "" {
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
	db   *sql.DB
}

func mainNewContext(cfg *conf) *mainContext {

	db, err := sql.Open("sqlite3", cfg.DatabasePath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open database: %s.\n", err.Error())
		os.Exit(1)
	}

	err = dbInit(db, cfg.AssetsPath()+"/sql")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s.\n", err.Error())
		os.Exit(1)
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
		db:   db,
		auth: &auth{},
	}

	ctx.site.SetFieldMap(map[string]interface{}{
		"site-title":       cfg.WebsiteTitle(),
		"site-description": cfg.WebsiteDescription(),
		"page-title":       "",
		"page-messages":    []template.HTML{},
		"admin":            false,
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
		err = http.ListenAndServe(":"+fmt.Sprint(ctx.conf.HttpPort()), ctx.site)
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
				err := mainIrssiLogs(ctx)
				if err != nil {
					log.Print(err)
				}
			}

		case "znc":
			do = func() {
				err := mainZncMethod(ctx)
				if err != nil {
					log.Print(err)
				}
			}

		}
		if do == nil {
			log.Fatalf("Unrecognized method, '%s'", method.Method)
		}

		scheduler.Schedule(do, initialTime, interval)
	}

	go scheduler.Run()
}

func mainIrssiLogs(ctx *mainContext) error {

	tx, err := ctx.db.Begin()
	if err != nil {
		return fmt.Errorf("Failed to initiate transaction: %w", err)
	}
	defer tx.Rollback()

	networks, err := dbGetNetworks(tx)
	if err != nil {
		return fmt.Errorf("Database error on retrieving networks: %w", err)
	}

	// networknames is a mapping from main network addresses (like
	// irc.libera.chat) to names that may be used for the network in Irssi
	// (like Libera)
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
		return fmt.Errorf("Database error on retrieving channels: %w", err)
	}

	logctx := irssilog.New(ctx.conf.IrssiLogsPath(), networknames)

	for _, ch := range chs {

		cs, err := logctx.ChannelStatus(ch.Name(), ch.Network())
		if err != nil {
			log.Printf("Error retrieving '%s@%s' status from log reading: %s", ch.Name(), ch.Network(), err.Error())
			continue
		}

		if cs.Time.After(ch.CheckTime()) {
			err = dbUpdateStatus(tx, ch.Name(), ch.Network(), cs.NumUsers, cs.Topic, "irssi-logs", "", cs.Time)
			if err != nil {
				return fmt.Errorf("Error updating status for '%s@%s': %w", ch.Name(), ch.Network(), err)
			}
		}

	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("Error on committing status updates from Irssi log reading: %w", err)
	}

	return nil
}

func mainZncGatherChannelStatus(ctx *mainContext, conn *irc.Conn, channel channel) error {
	conn.SendRawf("LIST %s", channel.Name())
	ev, err := conn.WaitUntil("322", "323") // RPL_LIST, RPL_LISTEND
	if err != nil {
		return fmt.Errorf("Error while executing the IRC LIST command: %w", err)
	}
	if ev.Code == "323" {
		return fmt.Errorf("Empty data in response to LIST command")
	} else {
		_, err = conn.WaitUntil("323") // RPL_LISTEND
		if err != nil {
			log.Printf("Error while while waiting for 323 RPL_LISTEND: %s\n", err)
		}
	}
	var numUsers int64 = 0
	if len(ev.Args) >= 3 {
		numUsers, err = strconv.ParseInt(ev.Args[2], 10, 32)
		if err != nil {
			return fmt.Errorf("Failed to parse the number of users for channel %s: %w", channel.Name(), err)
		}
	} else {
		return fmt.Errorf("Number of users was not included in RPL_LIST response")
	}
	topic := ""
	if len(ev.Args) == 4 {
		topic = ev.Args[3]
	}
	err = dbUpdateStatus(ctx.db, channel.Name(), channel.Network(), int(numUsers), topic, "znc", "", time.Now())
	if err != nil {
		return fmt.Errorf("Error updating status for '%s@%s': %w", channel.Name(), channel.Network(), err)
	}
	return nil
}

func mainZncNewConn(ctx *mainContext, network string) (*irc.Conn, error) {
	var netConn net.Conn
	var err error
	zncAddress := ctx.conf.ZncHost()
	if ctx.conf.ZncPort() != 0 { // TODO: allow use of 0 port
		zncAddress += fmt.Sprintf(":%d", ctx.conf.ZncPort())
	} else {
		zncAddress += fmt.Sprintf(":1025") // default znc port
	}
	tlsFingerprint := ctx.conf.ZncTlsFingerprint()
	if tlsFingerprint == "" {
		netConn, err = net.Dial("tcp", zncAddress)
	} else {
		config := &tls.Config{InsecureSkipVerify: true}
		config.VerifyConnection = func(cs tls.ConnectionState) error {
			sha1bytes := sha1.Sum(cs.PeerCertificates[0].Raw)
			sha1string := hex.EncodeToString(sha1bytes[:])
			sha1string = strings.ToLower(sha1string)
			tlsFingerprint = strings.ToLower(tlsFingerprint)
			if sha1string != tlsFingerprint {
				return fmt.Errorf("Expected SHA1 fingerprint '%s' does not match actual fingerprint '%s'", tlsFingerprint, sha1string)
			}
			return nil
		}
		netConn, err = tls.Dial("tcp", zncAddress, config)
	}
	if err != nil {
		return nil, err
	}

	zncNetworkName, ok := ctx.conf.ZncNetworkNames()[network]
	if !ok {
		return nil, fmt.Errorf("No ZNC network name configured for network '%s'", network)
	}
	nick := ctx.conf.ZncUser()
	user := fmt.Sprintf("%s/%s", ctx.conf.ZncUser(), zncNetworkName)
	ircConn, err := irc.New(netConn, &irc.Config{
		Nick:     nick,
		User:     user,
		Password: ctx.conf.ZncPassword(),
		Log:      log.Writer(),
	})
	if err != nil {
		return nil, fmt.Errorf("Error connecting to ZNC for network %s: %w", network, err)
	}

	return ircConn, nil
}

func mainZncGatherNetworkStatus(ctx *mainContext, network string, channels []channel) error {
	conn, err := mainZncNewConn(ctx, network)
	if err != nil {
		return err
	}
	defer conn.Disconnect()

	for _, channel := range channels {
		err = mainZncGatherChannelStatus(ctx, conn, channel)
		if err != nil {
			log.Printf("Failed to get channel status for %s/%s: %s\n", channel.Network(), channel.Name(), err)
		}
	}

	return nil
}

func mainZncMethod(ctx *mainContext) error {
	// TODO: This method is in large part copied from mainIrssiLogs. More code sharing is possible.

	chs, err := dbGetApprovedChannels(ctx.db, 0, 9999)
	if err != nil {
		return fmt.Errorf("Database error on retrieving channels: %w", err)
	}

	var network2channels map[string][]channel = util.GroupBy(chs, func(ch channel) string {
		return ch.Network()
	})

	for network, channels := range network2channels {
		err = mainZncGatherNetworkStatus(ctx, network, channels)
		if err != nil {
			log.Printf("Failed to gather status for network %s: %s\n", network, err)
		}
	}

	return nil
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
