package main

import (
	sql "code.google.com/p/go-sqlite/go1/sqlite3"
	"log"
	"time"
	"io"
)

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
			lastcheck,
			errmsg
		from
			channels
		where
			approved is not 0
		order by
			numusers desc;
	`)
	if err != nil {
		log.Fatalf("Failed to create schema: %s", err.Error())
	}
}

func dbOpen() *sql.Conn {
	c, err := sql.Open(conf.DatabasePath)
	if err != nil {
		log.Panicf("Failed to open database file '%s': %s.\n", conf.DatabasePath, err.Error())
	}

	return c
}

type dbChannel struct {
	name string
	server string
	weblink string
	description string
	numusers int
	approved bool
	lastcheck time.Time
	errmsg string
}

func dbGetChannel(c *sql.Conn, name, server string) *dbChannel {
	stmt, err := c.Query(`
		select
			weblink,
			description,
			numusers,
			lastcheck,
			errmsg
		from
			channels
		where
			name is ? and server is ?;
	`, name, server)
	defer stmt.Close()

	if err == io.EOF {
		return nil
	}

	if err != nil {
		log.Panicf("Unable to retrieve channel '%s@%s'.\n", name, server)
	}

	var (
		weblink string
		description string
		numusers int
		lastcheck string
		errmsg string
	)

	stmt.Scan(&weblink, &description, &numusers, &lastcheck, &errmsg)
	t, _ := time.Parse("2006-01-02 15:04:05", lastcheck)
	ch := &dbChannel{ name, server, weblink, description, numusers, true, t, errmsg }
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
		log.Panicf("Failed to update channel '%s@%s': %s.\n", originalName, originalServer, err.Error())
	}
}

func dbGetApprovedChannels(c *sql.Conn, off, len int) ([]dbChannel, int) {
	stmt, err := c.Query(`
		select
			name,
			server,
			weblink, 
			description,
			numusers,
			lastcheck,
			errmsg,
			(select count(*) from channels_approved)
		from
			channels_approved
		limit
			?
		offset
			?;
	`, len, off);

	if err == io.EOF {
		return []dbChannel{}, 0
	}

	if err != nil {
		log.Panicf("Failed to query channels from database: %s.\n", err.Error())
	}

	defer stmt.Close()

	channels := make([]dbChannel, 0, len)

	var (
		name string
		server string
		weblink string
		description string
		numusers int
		lastcheck string
		errmsg string
		numchannels int
	)

	for err = nil; err == nil; err = stmt.Next() {
		stmt.Scan(&name, &server, &weblink, &description, &numusers, &lastcheck, &errmsg, &numchannels)
		t, _ := time.Parse("2006-01-02 15:04:05", lastcheck)
		ch := dbChannel{ name, server, weblink, description, numusers, true, t, errmsg }
		channels = append(channels, ch)
	}

	return channels, numchannels
}

func dbAddChannel(c *sql.Conn, name, server, weblink, description string, numusers int) {
	err := c.Exec(`
		insert into channels
			(name, server, weblink, description, numusers, approved, lastcheck, errmsg)
		values
			(?, ?, ?, ?, ?, ?, datetime(), '');
	`, name, server, weblink, description, numusers, !conf.Approval)
	if err != nil {
		log.Panicf("Failed to add channel to database: %s.\n", err.Error())
	}
}
