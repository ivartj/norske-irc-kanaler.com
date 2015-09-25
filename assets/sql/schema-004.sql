-- This is preparation for a larger migration

-- +migrate Up
create table channel_status (
	channel_name text not null,
	channel_server text not null,
	status_time text not null,
	mode text, -- MAY BE NULL
	numusers integer, -- MAY BE NULL
	topic text, -- MAY BE NULL
	query_method text not null,
	errmsg text not null,
	foreign key (channel_name, channel_server) references channels (name, server)
);

-- +migrate Down
drop table channel_status;

