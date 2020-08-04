/*
 * create data mart tables 
 */

drop schema if exists "mart" cascade;
create schema "mart";

set search_path = "mart";

drop table if exists date_dimension;
create table date_dimension (
	date_key date primary key, -- date format yyyyMMdd
	-- date parts
	the_date date,
	the_year smallint, --
	quarter_number smallint, -- 1-4
	quarter_name char(2),	
	month_number smallint, -- 1-12
	month_name varchar(20),
	week_of_year smallint, -- 1-52
	weekday_number smallint, -- 1-7
	weekday_name varchar(10),
	day_of_month smallint, -- 1-31
	day_of_year smallint, -- 1-368  
	weekend boolean
);

drop table if exists time_dimension;
create table time_dimension (
	time_key time with time zone primary key, -- 24 hour time	
	the_time_24 time with time zone,
	the_time_12 time with time zone,
	-- time parts
	hour_12 smallint,
	hour_24 smallint,
	the_minute smallint,
	the_second smallint
	--nanosecond
);

drop table if exists customer_fact;
create table customer_fact (
	date_key date,
	time_key time with time zone,
	nanosecond smallint, -- optional field but could be good for granularity
	n smallint
);


drop function if exists notify_customer_fact_tr_fn cascade;
create function notify_customer_fact_tr_fn() returns trigger as $tr$
declare 
	_resp json;
	chan_res text;
  begin
	select jsonb_build_array(jsonb_build_object( 
		'time_stamp', new.date_key + new.time_key, 
		'n', new.n )) into _resp;
	  
    --notify "customer", _resp::text;
    select pg_notify('customer', _resp::text) into chan_res;
    return new;
  end;
$tr$ language plpgsql;

create trigger notify_customer_fact_tr after insert on customer_fact
  for each row execute function notify_customer_fact_tr_fn();


drop table if exists order_fact;
create table order_fact (
	date_key date,
	time_key time with time zone,
	nanosecond smallint, -- optional field but could be good for granularity
	product varchar(255),
	n smallint,
	revenue decimal(50, 5)
);

drop function if exists notify_order_fact_tr_fn cascade;
create function notify_order_fact_tr_fn() returns trigger as $tr$
declare 
	_resp json;
	chan_res text;
  begin
	select jsonb_build_array(jsonb_build_object( 
		'time_stamp', new.date_key + new.time_key, 
		'order_count', 1,
		'n', new.n,
		'revenue', new.revenue )) into _resp;
	  
    select pg_notify('order', _resp::text) into chan_res;
    return new;
  end;
$tr$ language plpgsql;

create trigger notify_order_fact_tr after insert on order_fact
  for each row execute function notify_order_fact_tr_fn();


/*
 *  create store tables 
 */

set search_path = "public";

drop table if exists customer cascade;
create table customer (
  customer_id serial primary key,
  name_last varchar(255),
  name_first varchar(255),
  street_address text,
  city varchar(255),
  state varchar(255),
  created_at timestamptz default now(),
  updated_at timestamptz default now(),
  deleted_at timestamptz
);

drop function if exists notify_customer_tr_fn cascade;
create function notify_customer_tr_fn() returns trigger as $tr$
--declare 
--	_date_key bigint;
--	_time_key bigint;
  begin
	--_date_key := (date_part('year', new.created_at) * 10000) + (date_part('month', new.created_at) * 100) + date_part('day', new.created_at);
	--_time_key := (date_part('hour', new.created_at) * 10000) + (date_part('minute', new.created_at) * 100) + date_part('second', new.created_at);

	insert into "mart"."customer_fact" (date_key, time_key, nanosecond, n) values (new.created_at::date, new.created_at::time, null, 1);
	  
    return new;
  end;
$tr$ language plpgsql;

create trigger notify_customer_tr after insert on customer
  for each row execute function notify_customer_tr_fn();


drop table if exists "order" cascade;
create table "order" (
  order_id serial primary key,
  customer_id bigint not null,
  product varchar(255),
  quantity int,
  unit_price money,
  sales_price money,
  created_at timestamptz default now(),
  updated_at timestamptz default now(),
  deleted_at timestamptz,
  constraint order_customer_id_fk foreign key (customer_id) references customer (customer_id)
);

drop function if exists notify_order_tr_fn cascade;
create function notify_order_tr_fn() returns trigger as $tr$
--declare
--	_date_key bigint;
--	_time_key bigint;
  begin
--	_date_key := (date_part('year', new.created_at) * 10000) + (date_part('month', new.created_at) * 100) + date_part('day', new.created_at);
--	_time_key := (date_part('hour', new.created_at) * 10000) + (date_part('minute', new.created_at) * 100) + date_part('second', new.created_at);

	new.sales_price := new.quantity * new.unit_price;

	insert into "mart"."order_fact" (date_key, time_key, nanosecond, product, n, revenue) values (new.created_at::date, cast(new.created_at as time with time zone), null, new.product, new.quantity, new.sales_price::decimal(50, 5));
	
    return new;
  end;
$tr$ language plpgsql;


create trigger notify_order_tr after insert on "order"
  for each row execute function notify_order_tr_fn();
