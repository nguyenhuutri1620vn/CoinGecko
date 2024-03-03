PLEASE CREATE TALBLE market_price_historis 
-- Table: public.market_price_histories

DROP TABLE IF EXISTS public.market_price_histories;

CREATE TABLE IF NOT EXISTS public.market_price_histories
(
    id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( CYCLE INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    symbol character varying COLLATE pg_catalog."default" NOT NULL,
    low double precision NOT NULL,
    high double precision NOT NULL,
    open double precision NOT NULL,
    close double precision NOT NULL,
    change double precision DEFAULT 0,
    CONSTRAINT market_price_histories_pkey PRIMARY KEY (id)
)

TABLESPACE pg_default;

ALTER TABLE IF EXISTS public.market_price_histories
    OWNER to postgres;
