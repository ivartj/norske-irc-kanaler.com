-- +migrate Up

create table channel_excluded (
	channel_name text not null,
	network text not null,
	exclude_reason text not null,

	primary key(channel_name, network)
);

drop view channel_approved;
create view channel_approved as
select
	*
from
	channel_all
where
	approved is not 0
except
select
	channel_all.*
from
	channel_all
	natural join
	channel_excluded
order by
	numusers desc;

drop view server_all;
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
	network as server,
	network
from
	(select
		network
	 from
		channel
	 union
	 select
		network
	 from
		channel_excluded
	 except
		select
			server
		from
			server);

create view channel_excluded_all_server_combinations as
select
	channel_name,
	s.server as network,
	exclude_reason
from
	channel_excluded c
	left join
	server_all s
	on (c.network = s.network);

-- +migrate Down

-- TODO

