create database if not exists balance_service;

use balance_service;

create table users_info (
	user_id int unsigned not null auto_increment,
    first_name varchar(30),
    last_name varchar(30),
    reg_date date not null,
    
    primary key(user_id)
);

create table users_balance (
	user_id int unsigned not null,
    balance double not null,
    
    foreign key (user_id) references users_info (user_id),
    primary key (user_id)
);

create table transactions (
	tr_id int unsigned not null auto_increment,
    user_id int unsigned not null,
    tr_value double not null,
    info set('replenishment', 'withdrawal', 'transfer'),
    tr_date datetime not null,
    
    foreign key (user_id) references users_info (user_id), 
    primary key (tr_id, user_id)
);