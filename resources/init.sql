CREATE TABLE IF NOT EXISTS dic_languages
(
    id   smallint                                          NOT NULL,
    name character varying(2) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT dic_languages_pkey PRIMARY KEY (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS dic_languages_name_unique_key ON dic_languages (lower(name));

-- INSERT INTO dic_languages
-- VALUES (0, 'en')
-- ON CONFLICT DO NOTHING;
INSERT INTO dic_languages
VALUES (1, 'id')
ON CONFLICT DO NOTHING;
INSERT INTO dic_languages
VALUES (2, 'fr')
ON CONFLICT DO NOTHING;
INSERT INTO dic_languages
VALUES (3, 'de')
ON CONFLICT DO NOTHING;
INSERT INTO dic_languages
VALUES (4, 'es')
ON CONFLICT DO NOTHING;
INSERT INTO dic_languages
VALUES (5, 'ru')
ON CONFLICT DO NOTHING;
INSERT INTO dic_languages
VALUES (6, 'zh')
ON CONFLICT DO NOTHING;
INSERT INTO dic_languages
VALUES (7, 'ko')
ON CONFLICT DO NOTHING;
INSERT INTO dic_languages
VALUES (8, 'hr')
ON CONFLICT DO NOTHING;
INSERT INTO dic_languages
VALUES (9, 'hi')
ON CONFLICT DO NOTHING;
INSERT INTO dic_languages
VALUES (10, 'uk')
ON CONFLICT DO NOTHING;
INSERT INTO dic_languages
VALUES (11, 'sr')
ON CONFLICT DO NOTHING;
INSERT INTO dic_languages
VALUES (12, 'ro')
ON CONFLICT DO NOTHING;
INSERT INTO dic_languages
VALUES (13, 'it')
ON CONFLICT DO NOTHING;
INSERT INTO dic_languages
VALUES (14, 'pt')
ON CONFLICT DO NOTHING;
INSERT INTO dic_languages
VALUES (15, 'pl')
ON CONFLICT DO NOTHING;

CREATE SEQUENCE IF NOT EXISTS translations_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 2147483647
    CACHE 1;

CREATE TABLE IF NOT EXISTS translations
(
    id            integer               NOT NULL DEFAULT nextval('translations_id_seq'::regclass),
    word_id       integer               NOT NULL,
    address       character varying(42) NOT NULL,
    language_id   smallint              NOT NULL,
    name          character varying(30) NOT NULL,
    description   character varying(150),
    req_timestamp timestamptz           NOT NULL,
    timestamp     timestamptz           NOT NULL DEFAULT CURRENT_TIMESTAMP,
    up_votes      integer               NOT NULL DEFAULT 0,
    down_votes    integer               NOT NULL DEFAULT 0,
    CONSTRAINT translations_pkey PRIMARY KEY (id),
    CONSTRAINT translations_language_id_fkey FOREIGN KEY (language_id)
        REFERENCES dic_languages (id) MATCH SIMPLE
);
CREATE UNIQUE INDEX IF NOT EXISTS translations_unique_key ON translations (word_id, lower(address), language_id);
CREATE INDEX IF NOT EXISTS translations_key ON translations (word_id, language_id, (up_votes - down_votes) desc, id);

CREATE TABLE IF NOT EXISTS votes
(
    translation_id integer               NOT NULL,
    address        character varying(42) NOT NULL,
    up             boolean               NOT NULL,
    req_timestamp  timestamptz           NOT NULL,
    timestamp      timestamptz           NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT votes_translation_id_fkey FOREIGN KEY (translation_id)
        REFERENCES translations (id) MATCH SIMPLE
);
CREATE UNIQUE INDEX IF NOT EXISTS votes_unique_key ON votes (translation_id, lower(address));
CREATE INDEX IF NOT EXISTS votes_translation_id_key ON votes (translation_id);

DO
$$
    BEGIN
        CREATE TYPE tp_submit_translation_result AS
        (
            res_code       smallint,
            translation_id bigint
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE tp_vote_result AS
        (
            res_code   smallint,
            up_votes   integer,
            down_votes integer
        );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

CREATE OR REPLACE FUNCTION submit_translation(p_address text,
                                              p_word_id integer,
                                              p_language text,
                                              p_name text,
                                              p_description text,
                                              p_req_timestamp timestamptz,
                                              p_confirmed_rate integer) RETURNS tp_submit_translation_result
    LANGUAGE 'plpgsql'
AS
$body$
DECLARE
    l_language_id   smallint;
    l_rate          smallint;
    l_id            integer;
    l_req_timestamp timestamptz;
BEGIN
    SELECT id
    INTO l_language_id
    FROM dic_languages
    WHERE lower(name) = lower(p_language);

    if l_language_id is null then
        return CAST(ROW (-1, 0) AS tp_submit_translation_result);
    end if;

    SELECT id, up_votes - down_votes, req_timestamp
    INTO l_id, l_rate, l_req_timestamp
    FROM translations
    WHERE word_id = p_word_id
      AND lower(address) = lower(p_address)
      AND language_id = l_language_id;

    if l_id is not null then
        if l_rate >= p_confirmed_rate then
            return CAST(ROW (2, 0) AS tp_submit_translation_result);
        end if;
        if l_req_timestamp >= p_req_timestamp then
            return CAST(ROW (3, 0) AS tp_submit_translation_result);
        end if;
        DELETE FROM votes WHERE translation_id = l_id;
        DELETE FROM translations WHERE id = l_id;
    end if;

    INSERT INTO translations (word_id, address, language_id, name, description, req_timestamp)
    VALUES (p_word_id, p_address, l_language_id, p_name, p_description, p_req_timestamp)
    RETURNING id INTO l_id;
    return CAST(ROW (0, l_id) AS tp_submit_translation_result);
END
$body$;

CREATE OR REPLACE FUNCTION vote(p_address text,
                                p_translation_id integer,
                                p_up boolean,
                                p_req_timestamp timestamptz) RETURNS tp_vote_result
    LANGUAGE 'plpgsql'
AS
$body$
DECLARE
    l_address        text;
    l_up             bool;
    l_up_change      smallint;
    l_down_change    smallint;
    l_req_timestamp  timestamptz;
    l_new_up_votes   integer;
    l_new_down_votes integer;
BEGIN
    SELECT address INTO l_address FROM translations WHERE id = p_translation_id;

    if l_address is null then
        return CAST(ROW (-1, 0, 0) AS tp_vote_result);
    end if;

    if l_address = p_address then
        return CAST(ROW (1, 0, 0) AS tp_vote_result);
    end if;

    SELECT up, req_timestamp
    INTO l_up, l_req_timestamp
    FROM votes
    WHERE translation_id = p_translation_id
      AND lower(address) = lower(p_address);

    if l_up is null then
        INSERT INTO votes (translation_id, address, up, req_timestamp)
        VALUES (p_translation_id, p_address, p_up, p_req_timestamp);
        if p_up then
            l_up_change = 1;
            l_down_change = 0;
        else
            l_up_change = 0;
            l_down_change = 1;
        end if;
    else
        if l_up = p_up then
            return CAST(ROW (3, 0, 0) AS tp_vote_result);
        end if;

        if l_req_timestamp >= p_req_timestamp then
            return CAST(ROW (2, 0, 0) AS tp_vote_result);
        end if;

        UPDATE votes
        SET up            = p_up,
            timestamp     = CURRENT_TIMESTAMP,
            req_timestamp = p_req_timestamp
        WHERE translation_id = p_translation_id
          AND lower(address) = lower(p_address);
        if p_up then
            l_up_change = 1;
            l_down_change = -1;
        else
            l_up_change = -1;
            l_down_change = 1;
        end if;
    end if;

    UPDATE translations
    SET up_votes   = up_votes + l_up_change,
        down_votes = down_votes + l_down_change,
        timestamp  = CURRENT_TIMESTAMP
    WHERE id = p_translation_id
    RETURNING up_votes, down_votes INTO l_new_up_votes, l_new_down_votes;

    return CAST(ROW (0, l_new_up_votes, l_new_down_votes) AS tp_vote_result);
END

$body$;