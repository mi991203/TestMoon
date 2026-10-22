-- Schema snapshot for testmoon_mock. Models and compatibility migrations are owned by the service DAO.
-- New databases only; existing installations are upgraded by service startup migrations.

CREATE TABLE IF NOT EXISTS `endpoints` (
  `id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` bigint NOT NULL DEFAULT '0',
  `org_id` char(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `project_id` char(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `name` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_endpoints_deleted_at` (`deleted_at`),
  KEY `idx_endpoints_org_id` (`org_id`),
  KEY `idx_endpoints_project_id` (`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `rules` (
  `id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` bigint NOT NULL DEFAULT '0',
  `org_id` char(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `project_id` char(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `endpoint_id` char(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `name` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `method` varchar(12) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `path` varchar(1024) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `priority` bigint DEFAULT NULL,
  `document` longtext COLLATE utf8mb4_unicode_ci,
  `status` bigint DEFAULT NULL,
  `delay_ms` bigint DEFAULT NULL,
  `headers_cipher` longtext COLLATE utf8mb4_unicode_ci,
  `body_cipher` longtext COLLATE utf8mb4_unicode_ci,
  `conditions_cipher` longtext COLLATE utf8mb4_unicode_ci,
  `definition_version` bigint NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_rules_deleted_at` (`deleted_at`),
  KEY `idx_rules_org_id` (`org_id`),
  KEY `idx_rules_project_id` (`project_id`),
  KEY `idx_rules_endpoint_id` (`endpoint_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
