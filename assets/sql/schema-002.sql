-- +migrate Up

create table servers (
	server text not null primary key,
	network text not null
);

drop view channels_approved;
drop view channels_unapproved;
drop view channel_latest;

create view channels_all as
select
	name,
	(case when network is null then server else network end) as server,
	weblink,
	description,
	numusers,
	approved,
	checked,
	lastcheck,
	errmsg,
	submitdate,
	approvedate
from
	channels left join servers using (server);

create view channels_approved as
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
	submitdate,
	approvedate
from
	channels_all
where
	approved is not 0
order by
	numusers desc, checked desc;

create view channel_latest as
select
	name,
	server,
	weblink,
	description,
	numusers,
	approved,
	checked,
	lastcheck,
	checked,
	errmsg,
	submitdate,
	approvedate
from
	channels_all
where
	approved is not 0
order by
	approvedate desc;

create view channels_unapproved as
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
	submitdate,
	approvedate
from
	channels_all
where
	approved is 0
order by
	submitdate desc;

-- +migrate Down

drop view channels_approved;
drop view channels_unapproved;
drop view channel_latest;
drop view channel_all;

drop table servers;

create view channels_approved as
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
	submitdate,
	approvedate
from
	channels
where
	approved is not 0
order by
	numusers desc, checked desc;

create view channel_latest as
select
	name,
	server,
	weblink,
	description,
	numusers,
	approved,
	checked,
	lastcheck,
	checked,
	errmsg,
	submitdate,
	approvedate
from
	channels
where
	approved is not 0
order by
	approvedate desc;

create view channels_unapproved as
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
	submitdate,
	approvedate
from
	channels
where
	approved is 0
order by
	submitdate desc;

