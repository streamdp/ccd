--
-- PostgreSQL database dump
--

-- Dumped from database version 14.6
-- Dumped by pg_dump version 14.6 (Ubuntu 14.6-0ubuntu0.22.04.1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

DROP DATABASE IF EXISTS cryptocompare;
--
-- Name: cryptocompare; Type: DATABASE; Schema: -; Owner: postgres
--

CREATE DATABASE cryptocompare WITH TEMPLATE = template0 ENCODING = 'UTF8' LOCALE = 'en_US.utf8';


ALTER DATABASE cryptocompare OWNER TO postgres;

\connect cryptocompare

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: cryptocompare; Type: DATABASE PROPERTIES; Schema: -; Owner: postgres
--

ALTER DATABASE cryptocompare SET search_path TO 'public', 'cryptocompare';


\connect cryptocompare

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: cryptocompare; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA cryptocompare;


ALTER SCHEMA cryptocompare OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: data; Type: TABLE; Schema: cryptocompare; Owner: postgres
--

CREATE TABLE cryptocompare.data (
    _id bigint NOT NULL,
    fromsym bigint NOT NULL,
    tosym bigint NOT NULL,
    change24hour double precision,
    changepct24hour double precision,
    open24hour double precision,
    volume24hour double precision,
    low24hour double precision,
    high24hour double precision,
    price double precision,
    supply double precision,
    mktcap double precision,
    lastupdate text NOT NULL,
    displaydataraw text
);


ALTER TABLE cryptocompare.data OWNER TO postgres;

--
-- Name: data__id_seq; Type: SEQUENCE; Schema: cryptocompare; Owner: postgres
--

CREATE SEQUENCE cryptocompare.data__id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE cryptocompare.data__id_seq OWNER TO postgres;

--
-- Name: data__id_seq; Type: SEQUENCE OWNED BY; Schema: cryptocompare; Owner: postgres
--

ALTER SEQUENCE cryptocompare.data__id_seq OWNED BY cryptocompare.data._id;


--
-- Name: fsym; Type: TABLE; Schema: cryptocompare; Owner: postgres
--

CREATE TABLE cryptocompare.fsym (
    _id bigint NOT NULL,
    symbol character varying(5),
    unicode character(1)
);


ALTER TABLE cryptocompare.fsym OWNER TO postgres;

--
-- Name: fsym__id_seq; Type: SEQUENCE; Schema: cryptocompare; Owner: postgres
--

CREATE SEQUENCE cryptocompare.fsym__id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE cryptocompare.fsym__id_seq OWNER TO postgres;

--
-- Name: fsym__id_seq; Type: SEQUENCE OWNED BY; Schema: cryptocompare; Owner: postgres
--

ALTER SEQUENCE cryptocompare.fsym__id_seq OWNED BY cryptocompare.fsym._id;


--
-- Name: tsym; Type: TABLE; Schema: cryptocompare; Owner: postgres
--

CREATE TABLE cryptocompare.tsym (
    _id bigint NOT NULL,
    symbol character varying(5),
    unicode character(1)
);


ALTER TABLE cryptocompare.tsym OWNER TO postgres;

--
-- Name: tsym__id_seq; Type: SEQUENCE; Schema: cryptocompare; Owner: postgres
--

CREATE SEQUENCE cryptocompare.tsym__id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE cryptocompare.tsym__id_seq OWNER TO postgres;

--
-- Name: tsym__id_seq; Type: SEQUENCE OWNED BY; Schema: cryptocompare; Owner: postgres
--

ALTER SEQUENCE cryptocompare.tsym__id_seq OWNED BY cryptocompare.tsym._id;


--
-- Name: data _id; Type: DEFAULT; Schema: cryptocompare; Owner: postgres
--

ALTER TABLE ONLY cryptocompare.data ALTER COLUMN _id SET DEFAULT nextval('cryptocompare.data__id_seq'::regclass);


--
-- Name: fsym _id; Type: DEFAULT; Schema: cryptocompare; Owner: postgres
--

ALTER TABLE ONLY cryptocompare.fsym ALTER COLUMN _id SET DEFAULT nextval('cryptocompare.fsym__id_seq'::regclass);


--
-- Name: tsym _id; Type: DEFAULT; Schema: cryptocompare; Owner: postgres
--

ALTER TABLE ONLY cryptocompare.tsym ALTER COLUMN _id SET DEFAULT nextval('cryptocompare.tsym__id_seq'::regclass);


--
-- Data for Name: data; Type: TABLE DATA; Schema: cryptocompare; Owner: postgres
--

COPY cryptocompare.data (_id, fromsym, tosym, change24hour, changepct24hour, open24hour, volume24hour, low24hour, high24hour, price, supply, mktcap, lastupdate, displaydataraw) FROM stdin;
\.


--
-- Data for Name: fsym; Type: TABLE DATA; Schema: cryptocompare; Owner: postgres
--

COPY cryptocompare.fsym (_id, symbol, unicode) FROM stdin;
1	BTC	Ƀ
2	XRP	\N
3	ETH	Ξ
4	BCH	\N
5	EOS	\N
6	LTC	Ł
7	XMR	\N
8	DASH	\N
\.


--
-- Data for Name: tsym; Type: TABLE DATA; Schema: cryptocompare; Owner: postgres
--

COPY cryptocompare.tsym (_id, symbol, unicode) FROM stdin;
1	USD	$
2	EUR	€
3	GBP	£
4	JPY	¥
5	RUR	₽
\.


--
-- Name: data__id_seq; Type: SEQUENCE SET; Schema: cryptocompare; Owner: postgres
--

SELECT pg_catalog.setval('cryptocompare.data__id_seq', 1, true);


--
-- Name: fsym__id_seq; Type: SEQUENCE SET; Schema: cryptocompare; Owner: postgres
--

SELECT pg_catalog.setval('cryptocompare.fsym__id_seq', 8, true);


--
-- Name: tsym__id_seq; Type: SEQUENCE SET; Schema: cryptocompare; Owner: postgres
--

SELECT pg_catalog.setval('cryptocompare.tsym__id_seq', 5, true);


--
-- Name: data idx_24578_primary; Type: CONSTRAINT; Schema: cryptocompare; Owner: postgres
--

ALTER TABLE ONLY cryptocompare.data
    ADD CONSTRAINT idx_24578_primary PRIMARY KEY (_id);


--
-- Name: fsym idx_24585_primary; Type: CONSTRAINT; Schema: cryptocompare; Owner: postgres
--

ALTER TABLE ONLY cryptocompare.fsym
    ADD CONSTRAINT idx_24585_primary PRIMARY KEY (_id);


--
-- Name: tsym idx_24590_primary; Type: CONSTRAINT; Schema: cryptocompare; Owner: postgres
--

ALTER TABLE ONLY cryptocompare.tsym
    ADD CONSTRAINT idx_24590_primary PRIMARY KEY (_id);


--
-- Name: idx_24578_fk_fsym_idx; Type: INDEX; Schema: cryptocompare; Owner: postgres
--

CREATE INDEX idx_24578_fk_fsym_idx ON cryptocompare.data USING btree (fromsym);


--
-- Name: idx_24578_fk_tsym_idx; Type: INDEX; Schema: cryptocompare; Owner: postgres
--

CREATE INDEX idx_24578_fk_tsym_idx ON cryptocompare.data USING btree (tosym);


--
-- Name: data fk_fsym; Type: FK CONSTRAINT; Schema: cryptocompare; Owner: postgres
--

ALTER TABLE ONLY cryptocompare.data
    ADD CONSTRAINT fk_fsym FOREIGN KEY (fromsym) REFERENCES cryptocompare.fsym(_id);


--
-- Name: data fk_tsym; Type: FK CONSTRAINT; Schema: cryptocompare; Owner: postgres
--

ALTER TABLE ONLY cryptocompare.data
    ADD CONSTRAINT fk_tsym FOREIGN KEY (tosym) REFERENCES cryptocompare.tsym(_id);


--
-- PostgreSQL database dump complete
--

