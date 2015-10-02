-- +migrate Up

create view channel_status_last_no_error as
select
	channel_name,
	network,
	status_time,
	numusers,
	topic,
	query_method,
	errmsg
from
	channel_status
	natural join
	(select
		channel_name,
		network,
		max(status_time) as status_time
	 from
		channel_status
	 where
		errmsg is ''
	 group by
		channel_name,
		network);

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
	natural join
	channel_status_last_no_error
order by
	(case when numusers is not null then numusers else 0 end) desc;

-- +migrate Down

-- TODO

