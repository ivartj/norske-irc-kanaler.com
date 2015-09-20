-- +migrate Up

create view servers_all as
select distinct
	(case when network is null then server else network end) as server
from
	(select distinct server from channels) left join servers using (server);

-- +migrate Down

drop view servers_all;

