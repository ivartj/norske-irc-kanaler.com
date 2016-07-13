-- +migrate Up

drop view channel_all;
create view channel_all as
select
	channel_name,
	network,
	weblink,
	description,
	submit_time,
	submit_time > datetime(datetime(), '-30 days') as new,
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

create view channel_indexed as
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
	new desc,
	numusers desc;

-- +migrate Down

