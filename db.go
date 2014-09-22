package main

import (
	sql "code.google.com/p/go-sqlite/go1/sqlite3"
	"fmt"
	"time"
	"io"
	"sync"
)

var dbLock *sync.Mutex = &sync.Mutex{}

func dbInit() {
	c := dbOpen()
	defer c.Close()
	err := c.Exec(`
		create table if not exists channels (
			name text not null,
			server text not null,
			weblink text not null,
			description text not null,
			numusers integer not null,
			approved integer not null,
			checked integer not null,
			lastcheck text not null,
			errmsg text not null,
			primary key (name, server)
		);

		create view if not exists channels_approved as
		select
			name,
			server,
			weblink, 
			description,
			numusers,
			approved,
			checked,
			lastcheck,
			errmsg
		from
			channels
		where
			approved is not 0
		order by
			numusers desc, checked desc;

		create view if not exists channels_unapproved as
		select
			name,
			server,
			weblink,
			description,
			numusers,
			approved,
			checked,
			lastcheck,
			errmsg
		from
			channels
		where
			approved is 0
		order by
			lastcheck desc;
	`)
	if err != nil {
		panic(fmt.Errorf("Failed to create schema: %s", err.Error()))
	}
}

func dbOpen() *sql.Conn {
	dbLock.Lock()
	c, err := sql.Open(conf.DatabasePath)
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
	errmsg string
}

func dbGetChannel(c *sql.Conn, name, server string) *dbChannel {
	stmt, err := c.Query(`
		select
			weblink,
			description,
			numusers,
			approved,
			checked,
			lastcheck,
			errmsg
		from
			channels
		where
			name is ? and server is ?;
	`, name, server)

	if err == io.EOF {
		return nil
	}

	if err != nil {
		panic(fmt.Errorf("Unable to retrieve channel '%s@%s'.\n", name, server))
	}
	defer stmt.Close()

	var (
		weblink string
		description string
		numusers int
		approved bool
		checked bool
		lastcheck string
		errmsg string
	)

	err = stmt.Scan(&weblink, &description, &numusers, &approved, &checked, &lastcheck, &errmsg)
	if err != nil {
		panic(err)
	}
	t, _ := time.Parse("2006-01-02 15:04:05", lastcheck)
	ch := &dbChannel{ name, server, weblink, description, numusers, approved, checked, t, errmsg }
	return ch
}

func dbEditChannel(c *sql.Conn, originalName, originalServer string, name, server, weblink, description string) {
	err := c.Exec(`
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

func dbUpdateStatus(c *sql.Conn, name, server string, numusers int, errmsg string) {
	err := c.Exec(`
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

func dbUncheck(c *sql.Conn, name, server string) {
	err := c.Exec(`
		update channels
		set
			checked = 0,
		where
			name is ? and server is ?;
	`, name, server)
	if err != nil {
		panic(fmt.Errorf("Failed to uncheck channel: %s", err.Error()))
	}
}

func dbGetApprovedChannels(c *sql.Conn, off, len int) ([]dbChannel, int) {
	return dbGetChannels(c, off, len, true)
}

func dbGetUnapprovedChannels(c *sql.Conn, off, len int) ([]dbChannel, int) {
	return dbGetChannels(c, off, len, false)
}

func dbDeleteChannel(c *sql.Conn, name, server string) {
	err := c.Exec(`
		delete from channels
		where
			name = ? and server = ?;
	`, name, server)
	if err != nil {
		panic(fmt.Errorf("Failed to delete channel '%s@%s': %s", name, server, err.Error()))
	}
}

func dbApproveChannel(c *sql.Conn, name, server string) {
	err := c.Exec(`
		update channels
		set
			approved = 1
		where
			name = ? and server = ?;
	`, name, server)
	if err != nil {
		panic(fmt.Errorf("Failed to approve channel '%s@%s': %s", name, server, err.Error()))
	}
}

func dbGetChannels(c *sql.Conn, off, len int, approvedTable bool) ([]dbChannel, int) {

	table := ""
	switch approvedTable {
	case true:
		table = "channels_approved"
	case false:
		table = "channels_unapproved"
	}

	stmt, err := c.Query(`
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

	defer stmt.Close()

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
		numchannels int
	)

	for err = nil; err == nil; err = stmt.Next() {
		stmt.Scan(&name, &server, &weblink, &description, &numusers, &approved, &checked, &lastcheck, &errmsg, &numchannels)
		t, _ := time.Parse("2006-01-02 15:04:05", lastcheck)
		ch := dbChannel{
			name: name,
			server: server,
			weblink: weblink,
			description: description,
			numusers: numusers,
			approved: approved,
			checked: checked,
			lastcheck: t,
			errmsg: errmsg }
		channels = append(channels, ch)
	}

	return channels, numchannels
}

func dbAddChannel(c *sql.Conn, name, server, weblink, description string, numusers int) {
	err := c.Exec(`
		insert into channels
			(name, server, weblink, description, numusers, approved, checked, lastcheck, errmsg)
		values
			(?, ?, ?, ?, ?, ?, 0, datetime(), '');
	`, name, server, weblink, description, numusers, !conf.Approval)
	if err != nil {
		panic(fmt.Errorf("Failed to add channel to database: %s.\n", err.Error()))
	}
}
