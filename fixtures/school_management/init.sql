-- Run as an administrator against the existing MySQL instance.
-- No platform tables or existing schemas are changed.
CREATE DATABASE IF NOT EXISTS testmoon_school_management
  CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
GRANT ALL ON testmoon_school_management.* TO 'testmoon'@'%';
