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
	"html/template"
)

var dbLock *sync.Mutex = &sync.Mutex{}

const dbTimeFmt string = "2006-01-02 15:04:05"

type dbConn struct {
	*sql.DB
}

func dbInit() {
	migrations := &migrate.FileMigrationSource{
		Dir: conf.AssetsPath + "/sql",
	}

	c, err := dbOpen()
	if err != nil {
		log.Fatalf("Failed to open database:\n%s\n", err.Error())
	}
	defer c.Close()

	n, err := migrate.Exec(c.DB, "sqlite3", migrations, migrate.Up)
	log.Printf("Applied %d migrations.\n", n);
	if err != nil {
		log.Fatalf("Error on applying migrations:\n%s\n", err.Error())
	}
}

func dbOpen() (*dbConn, error) {
	dbLock.Lock()
	c, err := sql.Open("sqlite3", conf.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to open database file '%s': %s.\n", conf.DatabasePath, err.Error())
	}
	dbLock.Unlock()

	_, err = c.Exec("pragma foreign_keys = true;")
	if err != nil {
		return nil, err
	}

	return &dbConn{c}, nil
}

// Should more or less match the columns of the channel_all view
type dbChannel struct {
	channel
	channel_name string
	network string
	weblink string
	description string
	submit_time time.Time
	approved bool
	approve_time time.Time
	numusers int
	topic string
	checked bool
	check_time time.Time
	errmsg string
}

func (ch *dbChannel) Name() string { return ch.channel_name }
func (ch *dbChannel) Network() string { return ch.network }
func (ch *dbChannel) Weblink() string { return ch.weblink }
func (ch *dbChannel) Description() string { return ch.description }
func (ch *dbChannel) SubmitTime() time.Time { return ch.submit_time }
func (ch *dbChannel) Approved() bool { return ch.approved }
func (ch *dbChannel) ApproveTime() time.Time { return ch.approve_time }
func (ch *dbChannel) NumberOfUsers() int { return ch.numusers }
func (ch *dbChannel) Topic() string { return ch.topic }
func (ch *dbChannel) Checked() bool { return ch.checked }
func (ch *dbChannel) CheckTime() time.Time { return ch.check_time }
func (ch *dbChannel) Error() string { return ch.errmsg }

func (ch *dbChannel) Status() string {
	str, _ := channelStatusString(ch)
	return str
}

func dbScanChannel(scan interface{ Scan(dest ...interface{}) error }) (*dbChannel, error) {
	var (
		ch dbChannel
		submit_time string
		approve_time string
		numusers sql.NullInt64
		topic sql.NullString
		check_time sql.NullString
		errmsg sql.NullString
	)

	err := scan.Scan(
		&ch.channel_name,
		&ch.network,
		&ch.weblink,
		&ch.description,
		&submit_time,
		&ch.approved,
		&approve_time,
		&numusers,
		&topic,
		&check_time,
		&errmsg,
	)

	if err != nil {
		return nil, err
	}

	ch.submit_time, err = time.Parse(dbTimeFmt, submit_time)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse submission time: %s", err.Error())
	}

	if approve_time != "" {
		ch.approve_time, err = time.Parse(dbTimeFmt, approve_time)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse approval time: %s", err.Error())
		}
	}

	if !check_time.Valid {
		ch.checked = false
	} else {
		ch.checked = true
		ch.check_time, err = time.Parse(dbTimeFmt, check_time.String)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse status time: %s", err.Error())
		}

		ch.numusers = int(numusers.Int64)
		ch.topic = topic.String
		ch.errmsg = errmsg.String
	}

	return &ch, nil
}

func (c *dbConn) GetChannel(name, network string) (channel, error) {
	row := c.QueryRow(`
		select
			*
		from
			channel_all_server_combinations
		where
			channel_name is ? and network is ?;
	`, name, network)

	ch, err := dbScanChannel(row)
	if err != nil {
		return nil, fmt.Errorf("Failed to get channel %s@%s: %s", name, network, err.Error())
	}

	return ch, nil
}

func (c *dbConn) EditChannel(originalName, originalServer string, name, network, weblink, description string) error {
	_, err := c.Exec(`
		update channel
		set
			channel_name = ?,
			network = ?,
			weblink = ?,
			description = ?
		where
			channel_name is ? and network is ?;
	`, name, network, weblink, description, originalName, originalServer)
	if err != nil {
		return fmt.Errorf("Failed to update channel %s@%s: %s", originalName, originalServer, err.Error())
	}
	return nil
}

func (c *dbConn) UpdateStatus(name, network string, numusers int, topic, query_method, errmsg string) error {
	_, err := c.Exec(`
		insert into channel_status
			(channel_name, network, numusers, topic, query_method, errmsg, status_time)
		values
			(?, ?, ?, ?, ?, ?, datetime());
	`, name, network, numusers, topic, query_method, errmsg)
	if err != nil {
		return fmt.Errorf("Failed to store channel status: %s", err.Error())
	}

	return nil
}

func (c *dbConn) DeleteChannel(name, network string) error {
	_, err := c.Exec(`
		delete from channel
		where
			channel_name = ? and network = ?;
	`, name, network)
	if err != nil {
		return fmt.Errorf("Failed to delete channel '%s@%s': %s", name, network, err.Error())
	}
	return nil
}

func (c *dbConn) ApproveChannel(name, network string) error {
	_, err := c.Exec(`
		update channel
		set
			approved = 1,
			approve_time = datetime()
		where
			channel_name = ? and network = ?;
	`, name, network)
	if err != nil {
		return fmt.Errorf("Failed to approve channel '%s@%s': %s", name, network, err.Error())
	}
	return nil
}

func (c *dbConn) GetApprovedChannels(off, len int) ([]channel, error) {
	return c.GetChannels(off, len, "approved")
}

func (c *dbConn) GetUnapprovedChannels(off, len int) ([]channel, error) {
	return c.GetChannels(off, len, "unapproved")
}

func (c *dbConn) GetLatestChannels(off, len int) ([]channel, error) {
	return c.GetChannels(off, len, "latest")
}

func (c *dbConn) GetChannels(off, len int, tablename string) ([]channel, error) {

	table := "channel_" + tablename

	rows, err := c.Query(`
		select
			*
		from
			` + table + `
		limit
			?
		offset
			?;
	`, len, off);

	if err == io.EOF {
		return []channel{}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("Failed to query channels from database: %s.\n", err.Error())
	}

	defer rows.Close()

	channels := []channel{}

	for rows.Next() {
		ch, err := dbScanChannel(rows)
		if err != nil {
			return nil, err
		}
		channels = append(channels, ch)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return channels, nil
}

type dbServer struct {
	server, network string
}

type dbNetwork struct {
	network string
	servers []string
}

func (c *dbConn) GetNetworks() ([]*dbNetwork, error) {
	servers, err := c.GetServers()
	if err != nil {
		return nil, err
	}

	networkMap := make(map[string][]string)
	for _, server := range servers {
		networkServers, ok := networkMap[server.network]
		if !ok {
			networkServers = []string{}
		}
		networkServers = append(networkServers, server.server)
		networkMap[server.network] = networkServers
	}

	networks := []*dbNetwork{}

	for k, v := range networkMap {
		networks = append(networks, &dbNetwork{k, v})
	}

	return networks, nil
}

func (c *dbConn) GetServers() ([]*dbServer, error) {
	rows, err := c.Query(`
		select
			server, network
		from
			server_all;
	`);

	if err == io.EOF {
		return []*dbServer{}, nil
	}

	if err != nil {
		// TODO: More descriptive error
		return nil, err
	}

	defer rows.Close()

	servers := []*dbServer{}

	for rows.Next() {
		server := &dbServer{}
		rows.Scan(&server.server, &server.network)
		servers = append(servers, server)
	}

	if rows.Err() != nil {
		// TODO: More descriptive error
		return nil, rows.Err()
	}

	return servers, nil
}

func (c *dbConn) AddServer(server, network string) error {
	_, err := c.Exec(`
		insert into server
			(server, network)
		values
			(?, ?)
	`, server, network)
	if err != nil {
		return fmt.Errorf("Failed to add server '%s', %s", server, err.Error())
	}
	return nil
}


func (c *dbConn) AddChannel(name, network, weblink, description string, approved bool) error {
	_, err := c.Exec(`
		insert into channel
			(channel_name,
			 network,
			 weblink,
			 description,
			 approved,
			 submit_time,
			 approve_time)
		values
			(?, -- name
			 (select
				(case when server_table.network is null
				 then submit.server
				 else server_table.network
				 end)
			  from
				(select ? as server) submit
				 left natural join
				 server server_table), -- network
			 ?, -- weblink
			 ?, -- description
			 ?, -- approved
			 datetime(),
			 datetime());
	`, name, network, weblink, description, approved)
	if err != nil {
		// TODO: More descriptive error
		return err
	}

	return nil
}

func (c *dbConn) IsChannelExcluded(name, network string) (bool, string, error) {
	row := c.QueryRow(`
		select
			exclude_reason
		from
			channel_excluded_all_server_combinations
		where
			channel_name is ? and network is ?;
	`, name, network)

	var exclude_reason string
	err := row.Scan(&exclude_reason)

	if err == sql.ErrNoRows {
		return false, "", nil
	} else if err != nil {
		return false, "", err
	}

	return true, exclude_reason, nil
}

func (c *dbConn) GetNumberOfChannelsUnapproved() (int, error) {
	row := c.QueryRow(`
		select
			count(*)
		from
			channel_unapproved;
	`)

	var num int
	err := row.Scan(&num)
	return num, err
}

func (c *dbConn) GetNumberOfChannelsExcluded() (int, error) {
	row := c.QueryRow(`
		select
			count(*)
		from
			channel_excluded;
	`)

	var num int
	err := row.Scan(&num)
	return num, err
}

func (c *dbConn) AddExclusion(name, network, reason string) error {
	_, err := c.Exec(`
		insert into channel_excluded
			(channel_name,
			 network,
			 exclude_reason)
		values
			(?, -- name
			 (select
				(case when server_table.network is null
				 then submit.server
				 else server_table.network
				 end)
			  from
				(select ? as server) submit
				 left natural join
				 server server_table), -- network
			 ? -- reason
			);
	`, name, network, reason)
	if err != nil {
		// TODO: More descriptive error
		return err
	}

	return nil
}

type dbExclusion struct {
	name, network, reason string
}

func (ex *dbExclusion) Name() string { return ex.name }
func (ex *dbExclusion) Network() string { return ex.network }
func (ex *dbExclusion) Reason() template.HTML { return template.HTML(ex.reason) }

func (c *dbConn) GetExclusions() ([]*dbExclusion, error) {
	rows, err := c.Query(`
		select
			channel_name, network, exclude_reason
		from
			channel_excluded;
	`);

	if err == io.EOF {
		return []*dbExclusion{}, nil
	}

	if err != nil {
		// TODO: More descriptive error
		return nil, err
	}

	defer rows.Close()

	exs := []*dbExclusion{}

	for rows.Next() {
		ex := &dbExclusion{}
		rows.Scan(&ex.name, &ex.network, &ex.reason)
		exs = append(exs, ex)
	}

	if rows.Err() != nil {
		// TODO: More descriptive error
		return nil, rows.Err()
	}

	return exs, nil
}

func (c *dbConn) DeleteExclusion(name, network string) error {
	_, err := c.Exec(`
		delete from channel_excluded
		where
			channel_name = ? and network = ?;
	`, name, network)
	if err != nil {
		return fmt.Errorf("Failed to delete exclusion '%s@%s': %s", name, network, err.Error())
	}
	return nil
}

