-- Schema snapshot for testmoon_execution. Models and compatibility migrations are owned by the service DAO.
-- New databases only; existing installations are upgraded by service startup migrations.

CREATE TABLE IF NOT EXISTS `budgets` (
  `org_id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`org_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `consumed_events` (
  `id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `outboxes` (
  `id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` bigint NOT NULL DEFAULT '0',
  `run_id` char(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `payload` text COLLATE utf8mb4_unicode_ci,
  `published` tinyint(1) DEFAULT NULL,
  `lease_until` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_outboxes_run_id` (`run_id`),
  KEY `idx_outboxes_deleted_at` (`deleted_at`),
  KEY `idx_outboxes_published` (`published`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `runners` (
  `id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` bigint NOT NULL DEFAULT '0',
  `org_id` char(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `project_id` char(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `name` longtext COLLATE utf8mb4_unicode_ci,
  `public` tinyint(1) DEFAULT NULL,
  `revoked` tinyint(1) DEFAULT NULL,
  `enrollment_hash` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `credential_hash` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `registered` tinyint(1) DEFAULT NULL,
  `last_seen` datetime(3) DEFAULT NULL,
  `concurrency` bigint DEFAULT NULL,
  `version` longtext COLLATE utf8mb4_unicode_ci,
  `capabilities` longtext COLLATE utf8mb4_unicode_ci,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_runners_enrollment_hash` (`enrollment_hash`),
  KEY `idx_runners_deleted_at` (`deleted_at`),
  KEY `idx_runners_org_id` (`org_id`),
  KEY `idx_runners_project_id` (`project_id`),
  KEY `idx_runners_credential_hash` (`credential_hash`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `runs` (
  `id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` bigint NOT NULL DEFAULT '0',
  `org_id` char(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `project_id` char(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `user_id` longtext COLLATE utf8mb4_unicode_ci,
  `task_id` longtext COLLATE utf8mb4_unicode_ci,
  `name` longtext COLLATE utf8mb4_unicode_ci,
  `state` varchar(24) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `target_runner_id` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `node_id` char(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `request_hash` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `fingerprint` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `lease_hash` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `snapshot` longtext COLLATE utf8mb4_unicode_ci,
  `result` longtext COLLATE utf8mb4_unicode_ci,
  `timeout_seconds` bigint DEFAULT NULL,
  `queue_deadline` datetime(3) DEFAULT NULL,
  `lease_until` datetime(3) DEFAULT NULL,
  `deadline` datetime(3) DEFAULT NULL,
  `started_at` datetime(3) DEFAULT NULL,
  `finished_at` datetime(3) DEFAULT NULL,
  `error` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_runs_request_hash` (`request_hash`),
  KEY `idx_runs_deleted_at` (`deleted_at`),
  KEY `run_scope` (`org_id`,`project_id`),
  KEY `idx_runs_state` (`state`),
  KEY `idx_runs_target_runner_id` (`target_runner_id`),
  KEY `idx_runs_node_id` (`node_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `tasks` (
  `id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` bigint NOT NULL DEFAULT '0',
  `org_id` char(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `project_id` char(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `name` longtext COLLATE utf8mb4_unicode_ci,
  `revision` bigint DEFAULT NULL,
  `document` longtext COLLATE utf8mb4_unicode_ci,
  `items_json` json DEFAULT NULL,
  `environment_id` char(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `runner_id` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `timeout_seconds` bigint DEFAULT NULL,
  `config_version` bigint NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_tasks_deleted_at` (`deleted_at`),
  KEY `idx_tasks_org_id` (`org_id`),
  KEY `idx_tasks_project_id` (`project_id`),
  KEY `idx_tasks_environment_id` (`environment_id`),
  KEY `idx_tasks_runner_id` (`runner_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
