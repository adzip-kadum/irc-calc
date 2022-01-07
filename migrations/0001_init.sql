CREATE TABLE irc_calcs
(
    id      BIGSERIAL     NOT NULL PRIMARY KEY,
    channel VARCHAR(100)  NOT NULL,
    "key"   VARCHAR(100)  NOT NULL,
    by      VARCHAR(255)  NOT NULL,
    "when"  TIMESTAMP     NOT NULL,
    content VARCHAR(1024) NOT NULL
);

CREATE INDEX channel_index
    ON irc_calcs USING BTREE (channel);
CREATE INDEX key_index
    ON irc_calcs USING BTREE ("key");

---- create above / drop below ----

DROP INDEX channel_index;
DROP INDEX key_index;
DROP TABLE irc_calcs;
