create type payment_status as enum ('Authorized', 'Declined');

create table if not exists payments(
    id varchar(36) primary key,
    order_id varchar(36) not null, 
    transaction_id varchar(255) unique,
    amount bigint not null,
    status payment_status not null,
    created_at timestamp with time zone default current_timestamp
);

create index idx_payments_order_id on payments(order_id);