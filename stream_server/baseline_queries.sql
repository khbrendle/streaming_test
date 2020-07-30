set search_path = 'mart';




select * from customer_fact cf2 limit 10;

-- new customers over time
@set groupMinute = 2
select ts n, sum(n) y
from (
	select 
		date_key + make_time(
			extract(hour from date_key + time_key)::int,
			cast(floor(extract(minute from date_key + time_key) / ${groupMinute}) * ${groupMinute} as int),
			0
		) ts
		,n
	from customer_fact cf2
) t
group by ts
order by ts;

-- customer count by hour
select dd.day_of_year , td.hour_24, td.the_minute , sum(cf.n) customer_count
from customer_fact cf 
join date_dimension dd 
	on cf.date_key = dd.date_key 
join time_dimension td 
	on cf.time_key = td.time_key 
group by dd.day_of_year , td.hour_24, td.the_minute
order by dd.day_of_year , td.hour_24, td.the_minute
;


-- order count by hour
select dd.day_of_year , td.hour_24, td.the_minute , sum(cf.n) order_count, sum(cf.revenue) total_revenue
from order_fact cf 
join date_dimension dd 
	on cf.date_key = dd.date_key 
join time_dimension td 
	on cf.time_key = td.time_key 
group by dd.day_of_year , td.hour_24, td.the_minute
order by dd.day_of_year , td.hour_24, td.the_minute
;