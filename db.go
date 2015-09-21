package main

import (
	"database/sql"
	_ "code.google.com/p/go-sqlite/go1/sqlite3"
	"github.com/rubenv/sql-migrate"
	"fmt"
	"time"
	"io"
	"sync"
	"log"
)

var dbLock *sync.Mutex = &sync.Mutex{}

const dbTimeFmt string = "2006-01-02 15:04:05"

func dbInit() {
	migrations := &migrate.FileMigrationSource{
		Dir: conf.AssetsPath + "/sql",
	}

	c := dbOpen()
	defer c.Close()

	n, err := migrate.Exec(c, "sqlite3", migrations, migrate.Up)
	log.Printf("Applied %d migrations.\n", n);
	if err != nil {
		log.Fatalf("Error on applying migrations:\n%s\n", err.Error())
	}
}

func dbOpen() *sql.DB {
	dbLock.Lock()
	c, err := sql.Open("sqlite3", conf.DatabasePath)
	if err != nil {
		panic(fmt.Errorf("Failed to open database file '%s': %s.\n", conf.DatabasePath, err.Error()))
	}
	dbLock.Unlock()

	return c
}

type dbChannel struct {
	name string
	server string
	weblink string
	description string
	numusers int
	approved bool
	checked bool
	lastcheck time.Time
	submitdate time.Time
	approvedate time.Time
	errmsg string
}

func dbGetChannel(c *sql.DB, name, server string) (*dbChannel, error) {
	row := c.QueryRow(`
		select
			weblink,
			description,
			numusers,
			approved,
			checked,
			lastcheck,
			errmsg,
			submitdate,
			approvedate
		from
			channels
		where
			name is ? and server is ?;
	`, name, server)

	var (
		weblink string
		description string
		numusers int
		approved bool
		checked bool
		lastcheck string
		errmsg string
		submitdate string
		approvedate string
	)

	err := row.Scan(&weblink, &description, &numusers, &approved, &checked, &lastcheck, &errmsg, &submitdate, &approvedate)
	if err != nil {
		return nil, err
	}
	tLastcheck, _ := time.Parse(dbTimeFmt, lastcheck)
	tSubmitdate, _ := time.Parse(dbTimeFmt, submitdate)
	tApprovedate, _ := time.Parse(dbTimeFmt, approvedate)
	ch := &dbChannel{
		name: name,
		server: server,
		weblink: weblink,
		description: description,
		numusers: numusers,
		approved: approved,
		checked: checked,
		lastcheck: tLastcheck,
		errmsg: errmsg,
		submitdate: tSubmitdate,
		approvedate: tApprovedate }
	return ch, nil
}

func dbEditChannel(c *sql.DB, originalName, originalServer string, name, server, weblink, description string) {
	_, err := c.Exec(`
		update channels
		set
			name = ?,
			server = ?,
			weblink = ?,
			description = ?
		where
			name is ? and server is ?;
	`, name, server, weblink, description, originalName, originalServer)
	if err != nil {
		panic(fmt.Errorf("Failed to update channel '%s@%s': %s.\n", originalName, originalServer, err.Error()))
	}
}

func dbUpdateStatus(c *sql.DB, name, server string, numusers int, errmsg string) {
	_, err := c.Exec(`
		update channels
		set
			numusers = ?,
			errmsg = ?,
			checked = 1,
			lastcheck = datetime()
		where
			name is ? and server is ?;
	`, numusers, errmsg, name, server)
	if err != nil {
		panic(fmt.Errorf("Failed to update channel status: %s", err.Error()))
	}
}

func dbUncheck(c *sql.DB, name, server string) {
	_, err := c.Exec(`
		update channels
		set
			checked = 0
		where
			name is ? and server is ?;
	`, name, server)
	if err != nil {
		panic(fmt.Errorf("Failed to uncheck channel: %s", err.Error()))
	}
}

func dbDeleteChannel(c *sql.DB, name, server string) {
	_, err := c.Exec(`
		delete from channels
		where
			name = ? and server = ?;
	`, name, server)
	if err != nil {
		panic(fmt.Errorf("Failed to delete channel '%s@%s': %s", name, server, err.Error()))
	}
}

func dbApproveChannel(c *sql.DB, name, server string) {
	_, err := c.Exec(`
		update channels
		set
			approved = 1,
			approvedate = datetime()
		where
			name = ? and server = ?;
	`, name, server)
	if err != nil {
		panic(fmt.Errorf("Failed to approve channel '%s@%s': %s", name, server, err.Error()))
	}
}

func dbGetApprovedChannels(c *sql.DB, off, len int) ([]dbChannel, int) {
	return dbGetChannels(c, off, len, "approved")
}

func dbGetUnapprovedChannels(c *sql.DB, off, len int) ([]dbChannel, int) {
	return dbGetChannels(c, off, len, "unapproved")
}

func dbGetLatestChannels(c *sql.DB, off, len int) ([]dbChannel, int) {
	return dbGetChannels(c, off, len, "latest")
}

func dbGetChannels(c *sql.DB, off, len int, tablename string) ([]dbChannel, int) {

	table := "channels_" + tablename

	rows, err := c.Query(`
		select
			name,
			server,
			weblink, 
			description,
			numusers,
			approved,
			checked,
			lastcheck,
			errmsg,
			submitdate,
			approvedate,
			(select count(*) from ` + table + `)
		from
			` + table + `
		limit
			?
		offset
			?;
	`, len, off);

	if err == io.EOF {
		return []dbChannel{}, 0
	}

	if err != nil {
		panic(fmt.Errorf("Failed to query channels from database: %s.\n", err.Error()))
	}

	defer rows.Close()

	channels := make([]dbChannel, 0, len)

	var (
		name string
		server string
		weblink string
		description string
		numusers int
		approved bool
		checked bool
		lastcheck string
		errmsg string
		submitdate string
		approvedate string
		numchannels int
	)

	for rows.Next() {
		rows.Scan(&name, &server, &weblink, &description, &numusers, &approved, &checked, &lastcheck, &errmsg, &submitdate, &approvedate, &numchannels)
		tLastcheck, _ := time.Parse(dbTimeFmt, lastcheck)
		tSubmitdate, _ := time.Parse(dbTimeFmt, submitdate)
		tApprovedate, _ := time.Parse(dbTimeFmt, approvedate)
		ch := dbChannel{
			name: name,
			server: server,
			weblink: weblink,
			description: description,
			numusers: numusers,
			approved: approved,
			checked: checked,
			lastcheck: tLastcheck,
			errmsg: errmsg,
			submitdate: tSubmitdate,
			approvedate: tApprovedate,
		 }
		channels = append(channels, ch)
	}

	return channels, numchannels
}

func dbGetServers(c *sql.DB) []string {
	rows, err := c.Query(`
		select
			server
		from
			servers_all;
	`);

	if err == io.EOF {
		return []string{}
	}

	if err != nil {
		panic(fmt.Errorf("Failed to query servers from database: %s.\n", err.Error()))
	}

	defer rows.Close()

	servers := []string{}

	var (
		server string
	)

	for rows.Next() {
		rows.Scan(&server)
		servers = append(servers, server)
	}

	return servers
}


func dbAddChannel(c *sql.DB, name, server, weblink, description string, numusers int) {
	_, err := c.Exec(`
		insert into channels
			(name, server, weblink, description, numusers, approved, checked, lastcheck, errmsg, submitdate, approvedate)
		values
			(?, ?, ?, ?, ?, ?, 0, datetime(), '', datetime(), datetime());
	`, name, server, weblink, description, numusers, !conf.Approval)
	if err != nil {
		panic(fmt.Errorf("Failed to add channel to database: %s.\n", err.Error()))
	}
}

