CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

create table users (
	id uuid default uuid_generate_v4() primary key,
	username text not null,
	email text not null,
	password text not null,
	registered timestamptz,
	updated timestamptz 
);

CREATE TABLE problems {
    
}