-- +migrate Up

create table if not exists channels (
	name text not null,
	server text not null,
	weblink text not null,
	description text not null,
	numusers integer not null,
	approved integer not null,
	checked integer not null,
	lastcheck text not null,
	errmsg text not null,
	submitdate text not null,
	approvedate text not null,
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

create view if not exists channel_latest as
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

create view if not exists channels_unapproved as
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

-- +migrate Down
drop view channels_unapproved;
drop view channel_latest;
drop view channels_approved;
drop table channels;
