create table orders (
    id serial primary key,
    user_id int not null,
    "order" varchar(512),
    count int,
    price int,
    status varchar(128) DEFAULT 'Новый',
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP
)