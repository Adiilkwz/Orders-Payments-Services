create type order_status as enum ('Pending', 'Paid', 'Failed', 'Cancelled');

create table if not exists orders (
    id varchar(36) primary key, 
    customer_id varchar(255) not null,
    item_name varchar(255) not null,
    amount bigint not null,
    status order_status not null, 
    created_at timestamp with time zone default current_timestamp
);

create index idx_orders_customer_id on orders(customer_id);