-- +migrate Up
drop view channel_all_server_combinations;
create view channel_all_server_combinations as
select
	channel_name,
	s.server as network, -- only difference
	weblink,
	description,
	submit_time,
	new,
	approved,
	approve_time,
	numusers,
	topic,
	check_time,
	errmsg
from
	channel_all c
	left join
	server_all s
	on (c.network = s.network);

-- +migrate Down

