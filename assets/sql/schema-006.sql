-- +migrate Up

-- Identical to previous iteration except for "on update cascade"

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
		on update cascade
);

insert into channel_status_new
select
	*
from
	channel_status;

drop table channel_status;
alter table channel_status_new rename to channel_status;

drop trigger server_to_network;

-- identical to previous iteration except for 'before'
create trigger server_to_network
before insert on server for each row
begin update channel set network = NEW.network where network = NEW.server; end;

-- +migrate Down

-- TODO

