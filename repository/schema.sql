--
-- PostgreSQL database dump
--

-- Dumped from database version 13.2 (Debian 13.2-1.pgdg100+1)
-- Dumped by pg_dump version 13.2 (Debian 13.2-1.pgdg100+1)

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
-- Name: tiger; Type: SCHEMA; Schema: -; Owner: root
--

CREATE SCHEMA tiger;


ALTER SCHEMA tiger OWNER TO root;

--
-- Name: tiger_data; Type: SCHEMA; Schema: -; Owner: root
--

CREATE SCHEMA tiger_data;


ALTER SCHEMA tiger_data OWNER TO root;

--
-- Name: topology; Type: SCHEMA; Schema: -; Owner: root
--

CREATE SCHEMA topology;


ALTER SCHEMA topology OWNER TO root;

--
-- Name: SCHEMA topology; Type: COMMENT; Schema: -; Owner: root
--

COMMENT ON SCHEMA topology IS 'PostGIS Topology schema';


--
-- Name: fuzzystrmatch; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS fuzzystrmatch WITH SCHEMA public;


--
-- Name: EXTENSION fuzzystrmatch; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION fuzzystrmatch IS 'determine similarities and distance between strings';


--
-- Name: postgis; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS postgis WITH SCHEMA public;


--
-- Name: EXTENSION postgis; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION postgis IS 'PostGIS geometry and geography spatial types and functions';


--
-- Name: postgis_tiger_geocoder; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS postgis_tiger_geocoder WITH SCHEMA tiger;


--
-- Name: EXTENSION postgis_tiger_geocoder; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION postgis_tiger_geocoder IS 'PostGIS tiger geocoder and reverse geocoder';


--
-- Name: postgis_topology; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS postgis_topology WITH SCHEMA topology;


--
-- Name: EXTENSION postgis_topology; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION postgis_topology IS 'PostGIS topology spatial types and functions';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: irc_calcs; Type: TABLE; Schema: public; Owner: root
--

CREATE TABLE public.irc_calcs (
    id bigint NOT NULL,
    channel character varying(100) NOT NULL,
    key character varying(100) NOT NULL,
    by character varying(255) NOT NULL,
    "when" timestamp without time zone NOT NULL,
    content character varying(1024) NOT NULL
);


ALTER TABLE public.irc_calcs OWNER TO root;

--
-- Name: irc_calcs_id_seq; Type: SEQUENCE; Schema: public; Owner: root
--

CREATE SEQUENCE public.irc_calcs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.irc_calcs_id_seq OWNER TO root;

--
-- Name: irc_calcs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: root
--

ALTER SEQUENCE public.irc_calcs_id_seq OWNED BY public.irc_calcs.id;


--
-- Name: migrations; Type: TABLE; Schema: public; Owner: root
--

CREATE TABLE public.migrations (
    version integer NOT NULL
);


ALTER TABLE public.migrations OWNER TO root;

--
-- Name: irc_calcs id; Type: DEFAULT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.irc_calcs ALTER COLUMN id SET DEFAULT nextval('public.irc_calcs_id_seq'::regclass);


--
-- Name: irc_calcs irc_calcs_pkey; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.irc_calcs
    ADD CONSTRAINT irc_calcs_pkey PRIMARY KEY (id);


--
-- Name: channel_index; Type: INDEX; Schema: public; Owner: root
--

CREATE INDEX channel_index ON public.irc_calcs USING btree (channel);


--
-- Name: key_index; Type: INDEX; Schema: public; Owner: root
--

CREATE INDEX key_index ON public.irc_calcs USING btree (key);


--
-- PostgreSQL database dump complete
--

