--
-- PostgreSQL database dump
--

\restrict WE1EiWrSZCpUO8dNFpjBjZLooVGi4ebO6DXW1355NIhE22xRXwbkPrMej5qrKZn

-- Dumped from database version 16.11 (Ubuntu 16.11-0ubuntu0.24.04.1)
-- Dumped by pg_dump version 16.11 (Ubuntu 16.11-0ubuntu0.24.04.1)

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
-- Name: beatmaps_id_seq; Type: SEQUENCE; Schema: public; Owner: advance
--

CREATE SEQUENCE public.beatmaps_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.beatmaps_id_seq OWNER TO advance;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: beatmaps; Type: TABLE; Schema: public; Owner: advance
--

CREATE TABLE public.beatmaps (
    id integer DEFAULT nextval('public.beatmaps_id_seq'::regclass) NOT NULL,
    beatmap_id integer NOT NULL,
    beatmapset_id integer NOT NULL,
    play_count integer DEFAULT 1 NOT NULL,
    pass_count integer DEFAULT 1 NOT NULL,
    title character varying(255) NOT NULL,
    artist character varying(255) NOT NULL,
    creator character varying(255) NOT NULL,
    creator_id integer NOT NULL,
    version character varying(255) NOT NULL,
    length integer NOT NULL,
    max_combo integer NOT NULL,
    ranked integer NOT NULL,
    last_update integer NOT NULL,
    added integer NOT NULL
);


ALTER TABLE public.beatmaps OWNER TO advance;

--
-- Name: mappers; Type: TABLE; Schema: public; Owner: advance
--

CREATE TABLE public.mappers (
    id integer NOT NULL,
    user_id integer NOT NULL,
    ranked_beatmaps integer,
    loved_beatmaps integer,
    graveyard_beatmaps integer,
    guest_beatmaps integer DEFAULT 0
);


ALTER TABLE public.mappers OWNER TO advance;

--
-- Name: mappers_id_seq; Type: SEQUENCE; Schema: public; Owner: advance
--

CREATE SEQUENCE public.mappers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    MAXVALUE 2147483647
    CACHE 1;


ALTER SEQUENCE public.mappers_id_seq OWNER TO advance;

--
-- Name: mappers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: advance
--

ALTER SEQUENCE public.mappers_id_seq OWNED BY public.mappers.id;


--
-- Name: scores; Type: TABLE; Schema: public; Owner: advance
--

CREATE TABLE public.scores (
    id integer NOT NULL,
    user_id integer NOT NULL,
    beatmap integer,
    score_id bigint,
    score integer,
    accuracy real NOT NULL,
    max_combo integer,
    count_50 integer,
    count_100 integer,
    count_300 integer,
    count_miss integer,
    fc boolean NOT NULL,
    mods text[],
    "time" integer,
    rank character varying(3),
    passed boolean NOT NULL,
    pp real DEFAULT 0,
    mode smallint NOT NULL,
    added integer
);


ALTER TABLE public.scores OWNER TO advance;

--
-- Name: scores_id_seq; Type: SEQUENCE; Schema: public; Owner: advance
--

CREATE SEQUENCE public.scores_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.scores_id_seq OWNER TO advance;

--
-- Name: scores_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: advance
--

ALTER SEQUENCE public.scores_id_seq OWNED BY public.scores.id;


--
-- Name: scores_go; Type: TABLE; Schema: public; Owner: advance
--

CREATE TABLE public.scores_go (
    id integer DEFAULT nextval('public.scores_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    beatmap integer,
    score_id bigint,
    score integer,
    accuracy real NOT NULL,
    max_combo integer,
    count_50 integer,
    count_100 integer,
    count_300 integer,
    count_miss integer,
    fc boolean NOT NULL,
    mods text[],
    "time" timestamp with time zone,
    rank character varying(3),
    passed boolean NOT NULL,
    pp real DEFAULT 0,
    mode smallint NOT NULL,
    added timestamp with time zone DEFAULT now()
);


ALTER TABLE public.scores_go OWNER TO advance;

--
-- Name: stats; Type: TABLE; Schema: public; Owner: advance
--

CREATE TABLE public.stats (
    id integer NOT NULL,
    user_id integer NOT NULL,
    global integer,
    country integer,
    pp integer,
    accuracy real NOT NULL,
    playcount integer NOT NULL,
    playtime integer NOT NULL,
    score bigint NOT NULL,
    hits integer NOT NULL,
    level integer NOT NULL,
    progress integer NOT NULL,
    mode smallint NOT NULL,
    "time" integer NOT NULL
);


ALTER TABLE public.stats OWNER TO advance;

--
-- Name: stats_id_seq; Type: SEQUENCE; Schema: public; Owner: advance
--

CREATE SEQUENCE public.stats_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.stats_id_seq OWNER TO advance;

--
-- Name: stats_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: advance
--

ALTER SEQUENCE public.stats_id_seq OWNED BY public.stats.id;


--
-- Name: stats_go; Type: TABLE; Schema: public; Owner: advance
--

CREATE TABLE public.stats_go (
    id integer DEFAULT nextval('public.stats_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    global integer DEFAULT 999999999,
    country integer DEFAULT 999999999,
    pp real DEFAULT 0,
    accuracy real NOT NULL,
    playcount integer NOT NULL,
    playtime integer NOT NULL,
    score bigint NOT NULL,
    hits integer NOT NULL,
    level integer NOT NULL,
    progress integer NOT NULL,
    mode smallint NOT NULL,
    "time" timestamp with time zone NOT NULL,
    replays_watched integer DEFAULT 0 NOT NULL
);


ALTER TABLE public.stats_go OWNER TO advance;

--
-- Name: users; Type: TABLE; Schema: public; Owner: advance
--

CREATE TABLE public.users (
    id integer NOT NULL,
    user_id integer NOT NULL,
    username character varying(255) NOT NULL,
    username_safe character varying(255) NOT NULL,
    country character(2) NOT NULL,
    added character varying(255) NOT NULL,
    restricted smallint NOT NULL,
    discord character varying(20) DEFAULT '0'::character varying NOT NULL
);


ALTER TABLE public.users OWNER TO advance;

--
-- Name: user_id_seq; Type: SEQUENCE; Schema: public; Owner: advance
--

CREATE SEQUENCE public.user_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.user_id_seq OWNER TO advance;

--
-- Name: user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: advance
--

ALTER SEQUENCE public.user_id_seq OWNED BY public.users.id;


--
-- Name: users_go; Type: TABLE; Schema: public; Owner: advance
--

CREATE TABLE public.users_go (
    id integer DEFAULT nextval('public.user_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    username character varying(255) DEFAULT ''::character varying NOT NULL,
    username_safe character varying(255) DEFAULT ''::character varying NOT NULL,
    country character(2) DEFAULT ''::character varying NOT NULL,
    added timestamp with time zone DEFAULT now() NOT NULL,
    restricted smallint DEFAULT 0 NOT NULL,
    discord character varying(20) DEFAULT '0'::character varying NOT NULL
);


ALTER TABLE public.users_go OWNER TO advance;

--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: advance
--

CREATE SEQUENCE public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.users_id_seq OWNER TO advance;

--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: advance
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: mappers id; Type: DEFAULT; Schema: public; Owner: advance
--

ALTER TABLE ONLY public.mappers ALTER COLUMN id SET DEFAULT nextval('public.mappers_id_seq'::regclass);


--
-- Name: scores id; Type: DEFAULT; Schema: public; Owner: advance
--

ALTER TABLE ONLY public.scores ALTER COLUMN id SET DEFAULT nextval('public.scores_id_seq'::regclass);


--
-- Name: stats id; Type: DEFAULT; Schema: public; Owner: advance
--

ALTER TABLE ONLY public.stats ALTER COLUMN id SET DEFAULT nextval('public.stats_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: advance
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: beatmaps beatmaps_pkey; Type: CONSTRAINT; Schema: public; Owner: advance
--

ALTER TABLE ONLY public.beatmaps
    ADD CONSTRAINT beatmaps_pkey PRIMARY KEY (id);


--
-- Name: mappers mappers_pkey; Type: CONSTRAINT; Schema: public; Owner: advance
--

ALTER TABLE ONLY public.mappers
    ADD CONSTRAINT mappers_pkey PRIMARY KEY (id);


--
-- Name: scores_go score_id_uni; Type: CONSTRAINT; Schema: public; Owner: advance
--

ALTER TABLE ONLY public.scores_go
    ADD CONSTRAINT score_id_uni UNIQUE (score_id);


--
-- Name: scores_go scores_go_pkey; Type: CONSTRAINT; Schema: public; Owner: advance
--

ALTER TABLE ONLY public.scores_go
    ADD CONSTRAINT scores_go_pkey PRIMARY KEY (id);


--
-- Name: scores scores_pkey; Type: CONSTRAINT; Schema: public; Owner: advance
--

ALTER TABLE ONLY public.scores
    ADD CONSTRAINT scores_pkey PRIMARY KEY (id);


--
-- Name: stats_go stats_go_pkey; Type: CONSTRAINT; Schema: public; Owner: advance
--

ALTER TABLE ONLY public.stats_go
    ADD CONSTRAINT stats_go_pkey PRIMARY KEY (id);


--
-- Name: stats stats_pkey; Type: CONSTRAINT; Schema: public; Owner: advance
--

ALTER TABLE ONLY public.stats
    ADD CONSTRAINT stats_pkey PRIMARY KEY (id);


--
-- Name: users_go user_id_uni; Type: CONSTRAINT; Schema: public; Owner: advance
--

ALTER TABLE ONLY public.users_go
    ADD CONSTRAINT user_id_uni UNIQUE (user_id);


--
-- Name: users_go users_go_pkey; Type: CONSTRAINT; Schema: public; Owner: advance
--

ALTER TABLE ONLY public.users_go
    ADD CONSTRAINT users_go_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: advance
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: idx_beatmaps_beatmap_id; Type: INDEX; Schema: public; Owner: advance
--

CREATE INDEX idx_beatmaps_beatmap_id ON public.beatmaps USING btree (beatmap_id);


--
-- Name: idx_scores_user_mode_time; Type: INDEX; Schema: public; Owner: advance
--

CREATE INDEX idx_scores_user_mode_time ON public.scores USING btree (user_id, mode, "time" DESC);


--
-- Name: idx_stats_mode; Type: INDEX; Schema: public; Owner: advance
--

CREATE INDEX idx_stats_mode ON public.stats_go USING btree (mode);


--
-- Name: idx_stats_mode_country; Type: INDEX; Schema: public; Owner: advance
--

CREATE INDEX idx_stats_mode_country ON public.stats_go USING btree (mode, country);


--
-- Name: idx_stats_mode_global; Type: INDEX; Schema: public; Owner: advance
--

CREATE INDEX idx_stats_mode_global ON public.stats_go USING btree (mode, global);


--
-- Name: idx_stats_mode_pp; Type: INDEX; Schema: public; Owner: advance
--

CREATE INDEX idx_stats_mode_pp ON public.stats_go USING btree (mode, pp DESC);


--
-- Name: idx_stats_pp; Type: INDEX; Schema: public; Owner: advance
--

CREATE INDEX idx_stats_pp ON public.stats_go USING btree (pp DESC);


--
-- Name: idx_stats_time; Type: INDEX; Schema: public; Owner: advance
--

CREATE INDEX idx_stats_time ON public.stats_go USING btree ("time" DESC);


--
-- Name: idx_stats_user; Type: INDEX; Schema: public; Owner: advance
--

CREATE INDEX idx_stats_user ON public.stats_go USING btree (user_id);


--
-- Name: idx_stats_user_mode; Type: INDEX; Schema: public; Owner: advance
--

CREATE INDEX idx_stats_user_mode ON public.stats_go USING btree (user_id, mode);


--
-- Name: idx_stats_user_mode_time; Type: INDEX; Schema: public; Owner: advance
--

CREATE INDEX idx_stats_user_mode_time ON public.stats USING btree (user_id, mode, "time" DESC);


--
-- Name: scores_go_user_id_mode_time_idx; Type: INDEX; Schema: public; Owner: advance
--

CREATE INDEX scores_go_user_id_mode_time_idx ON public.scores_go USING btree (user_id, mode, "time" DESC);


--
-- Name: stats_go_user_id_mode_time_idx; Type: INDEX; Schema: public; Owner: advance
--

CREATE INDEX stats_go_user_id_mode_time_idx ON public.stats_go USING btree (user_id, mode, "time" DESC);


--
-- PostgreSQL database dump complete
--

\unrestrict WE1EiWrSZCpUO8dNFpjBjZLooVGi4ebO6DXW1355NIhE22xRXwbkPrMej5qrKZn

