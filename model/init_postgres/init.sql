create database if not exists cryptocompare;

drop table if exists data;
create table data
(
    _id             bigint default unique_rowid() not null primary key,
    fromsym         bigint                        not null,
    tosym           bigint                        not null,
    change24hour    double precision,
    changepct24hour double precision,
    open24hour      double precision,
    volume24hour    double precision,
    low24hour       double precision,
    high24hour      double precision,
    price           double precision,
    supply          double precision,
    mktcap          double precision,
    lastupdate      text                          not null,
    displaydataraw  text
);


drop table if exists symbols;
create table symbols
(
    _id bigint default unique_rowid() not null
        constraint symbols_pk
        primary key,
    symbol varchar(64) default ''::STRING not null
        constraint symbols_symbol_uindex
        unique,
    unicode char
);

insert into symbols(symbol, unicode)
values ('USDT','₮'),
       ('BTC','₿'),
       ('ETH','⟠'),
       ('USD','$'),
       ('XRP','✕'),
       ('LTC','Ł'),
       ('EUR','€'),
       ('GBP','£'),
       ('JPY','¥');