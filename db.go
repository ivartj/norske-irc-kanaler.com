package main

import (
	sql "code.google.com/p/go-sqlite/go1/sqlite3"
	"log"
	"time"
	"io"
)

func dbInit() {
	c := dbOpen(conf)
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

func dbOpen(conf *config) *sql.Conn {
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
			count(*)
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

	for ; err == nil; err = stmt.Next() {
		stmt.Scan(&name, &server, &weblink, &description, &numusers, &lastcheck, &errmsg, &numchannels)
		t, _ := time.Parse("2006-01-02 15:04:05", lastcheck)
		ch := dbChannel{ name, server, weblink, description, numusers, true, t, errmsg }
		channels = append(channels, ch)
	}

	return channels, numchannels
}

