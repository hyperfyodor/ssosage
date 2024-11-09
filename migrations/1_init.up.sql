create table if not exists clients (
    id integer primary key,
    name text not null unique,
    password_hash blob not null
);

create index if not exists idx_client_name on clients (name);

create table if not exists apps (
    id integer primary key,
    name text not null unique,
    secret text not null,
    roles text not null
);

create index if not exists idx_app_name on apps (name);