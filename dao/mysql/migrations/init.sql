-- +migrate Up
CREATE TABLE `users`
(
    `usr_id`           int(11)          NOT NULL AUTO_INCREMENT,
    `usr_tg_id`        bigint(20)       NOT NULL,
    `usr_lang`         enum ('en','cn') NOT NULL DEFAULT 'en',
    `usr_name`         varchar(255)     NOT NULL,
    `usr_username`     varchar(255)     NOT NULL,
    `usr_mute`         tinyint(4)       NOT NULL DEFAULT '0',
    `usr_step`         varchar(255)     NOT NULL DEFAULT '',
    `usr_min_threshold` decimal(30, 10)  NOT NULL DEFAULT '0.0000000000',
    `usr_max_threshold` decimal(30, 10)  NOT NULL DEFAULT '0.0000000000',
    `usr_created_at`   timestamp        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`usr_id`),
    UNIQUE KEY `users_usr_tg_id_uindex` (`usr_tg_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;

CREATE TABLE `addresses`
(
    `adr_id`         int(11)     NOT NULL AUTO_INCREMENT,
    `adr_address`    varchar(35) NOT NULL,
    `adr_created_at` timestamp   NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`adr_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;

CREATE TABLE `users_addresses`
(
    `usr_id`    int(11)                      NOT NULL,
    `adr_id`    int(11)                      NOT NULL,
    `usa_alias` varchar(255)                 NOT NULL,
    `usa_type`  enum ('account','validator') NOT NULL,
    UNIQUE KEY `users_addresses_usr_id_adr_id_uindex` (`usr_id`, `adr_id`),
    KEY `users_addresses_addresses_adr_id_fk` (`adr_id`),
    CONSTRAINT `users_addresses_addresses_adr_id_fk` FOREIGN KEY (`adr_id`) REFERENCES `addresses` (`adr_id`),
    CONSTRAINT `users_addresses_users_usr_id_fk` FOREIGN KEY (`usr_id`) REFERENCES `users` (`usr_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;

CREATE TABLE `states`
(
    `stt_title` varchar(255) NOT NULL,
    `stt_value` text         NOT NULL,
    PRIMARY KEY (`stt_title`),
    UNIQUE KEY `states_stt_title_uindex` (`stt_title`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;

INSERT INTO states (stt_title, stt_value) VALUES ('current_height', 1);

-- +migrate Down
drop table users_addresses;
drop table addresses;
drop table users;
