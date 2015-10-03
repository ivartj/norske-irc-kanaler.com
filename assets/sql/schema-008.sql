-- +migrate Up

-- Identical to previous iteration except for 'natural *left* join'.

drop view channel_all;
create view channel_all as
select
	channel_name,
	network,
	weblink,
	description,
	submit_time,
	approved,
	approve_time,
	numusers,
	topic,
	status_time as check_time,
	errmsg
from
	channel
	natural left join
	channel_status_last_no_error
order by
	(case when numusers is not null then numusers else 0 end) desc;

create view channel_all_server_combinations as
select
	channel_name,
	s.server as network, -- only difference
	weblink,
	description,
	submit_time,
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
-- TODO

