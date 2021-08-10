-- +migrate Up

-- Adding 'on delete cascade' to channel_status table.

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
		on delete cascade
);

insert into channel_status_new
select
        *
from
        channel_status;

pragma legacy_alter_table = 1;
drop table channel_status;
alter table channel_status_new rename to channel_status;
pragma legacy_alter_table = 0;


-- +migrate Down

-- TODO

