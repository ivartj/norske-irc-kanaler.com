-- +migrate Up

-- CHANNEL <- CHANNELS

drop view servers_all;
drop view channels_all;
drop view channel_latest;
drop view channels_approved;
drop view channels_unapproved;

create table channel (
	channel_name text not null,
	network text not null,
	weblink text not null,
	description text not null,
	submit_time text not null,
	approved integer not null,
	approve_time text not null,

	primary key(channel_name, network)
);

insert into channel
select
	name as channel_name,
	server as network,
	weblink,
	description,
	submitdate as submit_time,
	approved,
	approvedate as approve_time
from
	channels;

drop table channels;

-- CHANNEL_STATUS <- CHANNEL_STATUS

create table channel_status_new (
	channel_name text not null,
	network text not null,
	status_time text not null,
	numusers integer not null,
	topic text not null,
	query_method text not null,
	errmsg text not null,
	foreign key(channel_name, network)
		references
		channel(channel_name, network)
);

-- Note: Error if topic or numusers is null
insert into channel_status_new
select
	channel_name,
	channel_server as network,
	status_time,
	numusers,
	topic,
	query_method,
	errmsg
from
	channel_status;

drop table channel_status;

alter table channel_status_new rename to channel_status;

-- SERVER <- SERVERS

alter table servers rename to server;

-- NEW VIEWS

create view channel_all as
select
	channel_name,
	network,
	weblink,
	description,
	submit_time,
	approved,
	approve_time,

	(select
		numusers
	 from
		channel_status
	 where
		channel_name = c.channel_name
		and network = c.network
	 order by
		status_time desc
	 limit 1) as numusers,

	(select
		topic
	 from
		channel_status
	 where
		channel_name = c.channel_name
		and network = c.network
	 order by
		status_time desc
	 limit 1) as topic,

	(select
		status_time
	 from
		channel_status
	 where
		channel_name = c.channel_name
		and network = c.network
	 order by
		status_time desc
	 limit 1) as check_time,

	(select
		errmsg
	 from
		channel_status
	 where
		channel_name = c.channel_name
		and network = c.network
	 order by
		status_time desc
	 limit 1) as errmsg
from
	channel c
order by
	(case when numusers is not null then numusers else 0 end) desc;

create view server_all as
select
	server,
	network
from
	server
union
select
	network as server,
	network
from
	server
union
select distinct
	server,
	server as network
from
	(select
		network as server
	 from
		channel
	 except
	 	select
			server
		from
			server);

create view channel_approved as
select
	*
from
	channel_all
where
	approved is not 0;

create view channel_unapproved as
select
	*
from
	channel_all
where
	approved is 0;

create trigger server_to_network
insert on server for each row
begin update channel set network = NEW.network where network = NEW.server; end;

-- +migrate Down

-- TODO

