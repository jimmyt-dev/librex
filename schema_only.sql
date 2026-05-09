/*M!999999\- enable the sandbox mode */ 
-- MariaDB dump 10.19-11.4.5-MariaDB, for Linux (x86_64)
--
-- Host: localhost    Database: booklore
-- ------------------------------------------------------
-- Server version	11.4.5-MariaDB-log

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*M!100616 SET @OLD_NOTE_VERBOSITY=@@NOTE_VERBOSITY, NOTE_VERBOSITY=0 */;

--
-- Table structure for table `annotations`
--

DROP TABLE IF EXISTS `annotations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `annotations` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) NOT NULL,
  `book_id` bigint(20) NOT NULL,
  `cfi` varchar(1000) NOT NULL,
  `text` varchar(5000) NOT NULL,
  `color` varchar(20) DEFAULT NULL,
  `style` varchar(50) DEFAULT NULL,
  `note` varchar(5000) DEFAULT NULL,
  `chapter_title` varchar(500) DEFAULT NULL,
  `version` bigint(20) NOT NULL DEFAULT 0,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_annotation_user_book_cfi` (`user_id`,`book_id`,`cfi`) USING HASH,
  KEY `idx_annotations_user_id` (`user_id`),
  KEY `idx_annotations_book_id` (`book_id`),
  KEY `idx_annotations_user_book` (`user_id`,`book_id`),
  KEY `idx_annotations_user_created` (`user_id`,`created_at`),
  CONSTRAINT `fk_annotations_book_id` FOREIGN KEY (`book_id`) REFERENCES `book` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_annotations_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `app_migration`
--

DROP TABLE IF EXISTS `app_migration`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `app_migration` (
  `migration_key` varchar(100) NOT NULL COMMENT 'Unique identifier for the migration',
  `executed_at` timestamp NOT NULL COMMENT 'When the migration was executed',
  `description` text DEFAULT NULL COMMENT 'Optional description of what the migration did',
  PRIMARY KEY (`migration_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='Tracks one-time application-level data migrations';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `app_settings`
--

DROP TABLE IF EXISTS `app_settings`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `app_settings` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `val` text DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_app_settings_name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=27 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `audit_log`
--

DROP TABLE IF EXISTS `audit_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `audit_log` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) DEFAULT NULL,
  `username` varchar(255) NOT NULL,
  `action` varchar(100) NOT NULL,
  `entity_type` varchar(100) DEFAULT NULL,
  `entity_id` bigint(20) DEFAULT NULL,
  `description` varchar(1024) NOT NULL,
  `ip_address` varchar(45) DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT current_timestamp(),
  `country_code` char(2) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_audit_log_created_at` (`created_at`),
  KEY `idx_audit_log_user_id` (`user_id`),
  KEY `idx_audit_log_action` (`action`)
) ENGINE=InnoDB AUTO_INCREMENT=4343 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `author`
--

DROP TABLE IF EXISTS `author`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `author` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) DEFAULT NULL,
  `description` text DEFAULT NULL,
  `asin` varchar(20) DEFAULT NULL,
  `name_locked` tinyint(1) NOT NULL DEFAULT 0,
  `description_locked` tinyint(1) NOT NULL DEFAULT 0,
  `asin_locked` tinyint(1) NOT NULL DEFAULT 0,
  `photo_locked` tinyint(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_name` (`name`),
  KEY `idx_author_asin` (`asin`)
) ENGINE=InnoDB AUTO_INCREMENT=35 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `book`
--

DROP TABLE IF EXISTS `book`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `book` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `library_id` bigint(20) NOT NULL,
  `library_path_id` bigint(20) DEFAULT NULL,
  `added_on` timestamp NULL DEFAULT current_timestamp(),
  `similar_books_json` text DEFAULT NULL,
  `metadata_match_score` float DEFAULT NULL,
  `read_status` varchar(20) DEFAULT 'UNREAD',
  `deleted` tinyint(1) DEFAULT 0,
  `deleted_at` timestamp NULL DEFAULT NULL,
  `metadata_updated_at` timestamp NULL DEFAULT NULL,
  `metadata_for_write_updated_at` timestamp NULL DEFAULT NULL,
  `book_cover_hash` varchar(20) DEFAULT NULL,
  `is_physical` tinyint(1) NOT NULL DEFAULT 0,
  `audiobook_cover_hash` varchar(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_library_path_id` (`library_path_id`),
  KEY `idx_library_id` (`library_id`),
  KEY `idx_book_deleted` (`deleted`),
  KEY `idx_book_deleted_at` (`deleted_at`),
  KEY `idx_book_is_physical` (`is_physical`),
  CONSTRAINT `fk_library` FOREIGN KEY (`library_id`) REFERENCES `library` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_library_path_id` FOREIGN KEY (`library_path_id`) REFERENCES `library_path` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=133 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `book_award`
--

DROP TABLE IF EXISTS `book_award`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `book_award` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `book_id` bigint(20) NOT NULL,
  `name` varchar(255) NOT NULL,
  `awarded_at` timestamp NOT NULL,
  `category` varchar(255) NOT NULL,
  `designation` varchar(255) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_book_award` (`book_id`,`name`,`category`,`awarded_at`),
  CONSTRAINT `fk_book_awards_book` FOREIGN KEY (`book_id`) REFERENCES `book` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `book_file`
--

DROP TABLE IF EXISTS `book_file`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `book_file` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `book_id` bigint(20) NOT NULL,
  `file_name` varchar(1000) NOT NULL,
  `file_sub_path` varchar(512) NOT NULL,
  `file_size_kb` bigint(20) DEFAULT NULL,
  `initial_hash` varchar(128) DEFAULT NULL,
  `current_hash` varchar(128) DEFAULT NULL,
  `description` text DEFAULT NULL,
  `added_on` timestamp NULL DEFAULT current_timestamp(),
  `is_book` tinyint(1) DEFAULT 0,
  `book_type` varchar(32) DEFAULT NULL,
  `archive_type` varchar(255) DEFAULT NULL,
  `alt_format_current_hash` varchar(128) GENERATED ALWAYS AS (case when `is_book` = 1 then `current_hash` end) STORED,
  `is_folder_based` tinyint(1) NOT NULL DEFAULT 0,
  `duration_seconds` bigint(20) DEFAULT NULL,
  `bitrate` int(11) DEFAULT NULL,
  `sample_rate` int(11) DEFAULT NULL,
  `channels` int(11) DEFAULT NULL,
  `codec` varchar(50) DEFAULT NULL,
  `chapter_count` int(11) DEFAULT NULL,
  `chapters_json` text DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_book_additional_file_book_id` (`book_id`),
  KEY `idx_book_file_current_hash_alt_format` (`alt_format_current_hash`),
  CONSTRAINT `fk_book_file_book` FOREIGN KEY (`book_id`) REFERENCES `book` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=136 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `book_marks`
--

DROP TABLE IF EXISTS `book_marks`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `book_marks` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) NOT NULL,
  `book_id` bigint(20) NOT NULL,
  `cfi` varchar(1000) DEFAULT NULL,
  `title` varchar(255) DEFAULT NULL,
  `created_at` datetime NOT NULL,
  `color` varchar(7) DEFAULT NULL,
  `notes` varchar(2000) DEFAULT NULL,
  `priority` int(11) DEFAULT NULL,
  `updated_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `version` bigint(20) NOT NULL DEFAULT 1,
  `position_ms` bigint(20) DEFAULT NULL,
  `track_index` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_user_book_cfi` (`user_id`,`book_id`,`cfi`) USING HASH,
  KEY `idx_book_marks_user_id` (`user_id`),
  KEY `idx_book_marks_book_id` (`book_id`),
  KEY `idx_bookmark_book_user_priority` (`book_id`,`user_id`,`priority`,`created_at`),
  KEY `idx_book_marks_user_created` (`user_id`,`created_at`),
  CONSTRAINT `fk_book_marks_book_id` FOREIGN KEY (`book_id`) REFERENCES `book` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_book_marks_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `book_metadata`
--

DROP TABLE IF EXISTS `book_metadata`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `book_metadata` (
  `book_id` bigint(20) NOT NULL,
  `title` varchar(1000) DEFAULT NULL,
  `subtitle` varchar(1000) DEFAULT NULL,
  `publisher` varchar(1000) DEFAULT NULL,
  `published_date` date DEFAULT NULL,
  `description` text DEFAULT NULL,
  `isbn_13` varchar(64) DEFAULT NULL,
  `isbn_10` varchar(64) DEFAULT NULL,
  `page_count` int(11) DEFAULT NULL,
  `thumbnail` varchar(2000) DEFAULT NULL,
  `language` varchar(255) DEFAULT NULL,
  `rating` float DEFAULT NULL,
  `review_count` int(11) DEFAULT NULL,
  `cover` varchar(2000) DEFAULT NULL,
  `cover_updated_on` timestamp NULL DEFAULT NULL,
  `series_name` varchar(1000) DEFAULT NULL,
  `series_number` float DEFAULT NULL,
  `series_total` int(11) DEFAULT NULL,
  `all_fields_locked` tinyint(1) DEFAULT 0,
  `title_locked` tinyint(1) DEFAULT 0,
  `authors_locked` tinyint(1) DEFAULT 0,
  `categories_locked` tinyint(1) DEFAULT 0,
  `subtitle_locked` tinyint(1) DEFAULT 0,
  `publisher_locked` tinyint(1) DEFAULT 0,
  `published_date_locked` tinyint(1) DEFAULT 0,
  `description_locked` tinyint(1) DEFAULT 0,
  `isbn_13_locked` tinyint(1) DEFAULT 0,
  `isbn_10_locked` tinyint(1) DEFAULT 0,
  `page_count_locked` tinyint(1) DEFAULT 0,
  `thumbnail_locked` tinyint(1) DEFAULT 0,
  `language_locked` tinyint(1) DEFAULT 0,
  `cover_locked` tinyint(1) DEFAULT 0,
  `rating_locked` tinyint(1) DEFAULT 0,
  `review_count_locked` tinyint(1) DEFAULT 0,
  `series_name_locked` tinyint(1) DEFAULT 0,
  `series_number_locked` tinyint(1) DEFAULT 0,
  `series_total_locked` tinyint(1) DEFAULT 0,
  `amazon_rating` float DEFAULT NULL,
  `amazon_review_count` int(11) DEFAULT NULL,
  `goodreads_rating` float DEFAULT NULL,
  `goodreads_review_count` int(11) DEFAULT NULL,
  `amazon_rating_locked` tinyint(1) DEFAULT 0,
  `amazon_review_count_locked` tinyint(1) DEFAULT 0,
  `goodreads_rating_locked` tinyint(1) DEFAULT 0,
  `goodreads_review_count_locked` tinyint(1) DEFAULT 0,
  `asin` varchar(20) DEFAULT NULL,
  `asin_locked` tinyint(1) DEFAULT 0,
  `hardcover_rating` float DEFAULT NULL,
  `hardcover_review_count` int(11) DEFAULT NULL,
  `hardcover_rating_locked` tinyint(1) DEFAULT 0,
  `hardcover_review_count_locked` tinyint(1) DEFAULT 0,
  `goodreads_id` varchar(100) DEFAULT NULL,
  `hardcover_id` varchar(512) DEFAULT NULL,
  `google_id` varchar(100) DEFAULT NULL,
  `goodreads_id_locked` tinyint(1) DEFAULT 0,
  `hardcover_id_locked` tinyint(1) DEFAULT 0,
  `google_id_locked` tinyint(1) DEFAULT 0,
  `comicvine_id` varchar(100) DEFAULT NULL,
  `comicvine_id_locked` tinyint(1) DEFAULT 0,
  `reviews_locked` tinyint(1) DEFAULT 0,
  `moods_locked` tinyint(1) DEFAULT 0,
  `tags_locked` tinyint(1) DEFAULT 0,
  `embedding_vector` text DEFAULT NULL,
  `embedding_updated_at` datetime DEFAULT NULL,
  `search_text` text DEFAULT NULL,
  `hardcover_book_id` varchar(100) DEFAULT NULL,
  `hardcover_book_id_locked` tinyint(1) DEFAULT 0,
  `lubimyczytac_id` varchar(100) DEFAULT NULL,
  `lubimyczytac_rating` float DEFAULT NULL,
  `lubimyczytac_id_locked` tinyint(1) DEFAULT 0,
  `lubimyczytac_rating_locked` tinyint(1) DEFAULT 0,
  `ranobedb_id` varchar(100) DEFAULT NULL,
  `ranobedb_rating` float DEFAULT NULL,
  `ranobedb_id_locked` tinyint(1) DEFAULT 0,
  `ranobedb_rating_locked` tinyint(1) DEFAULT 0,
  `audiobook_cover_updated_on` timestamp NULL DEFAULT NULL,
  `audible_id` varchar(100) DEFAULT NULL,
  `audible_rating` float DEFAULT NULL,
  `audible_review_count` int(11) DEFAULT NULL,
  `audible_id_locked` tinyint(1) DEFAULT 0,
  `audible_rating_locked` tinyint(1) DEFAULT 0,
  `audible_review_count_locked` tinyint(1) DEFAULT 0,
  `audiobook_cover_locked` tinyint(1) DEFAULT 0,
  `narrator_locked` tinyint(1) DEFAULT 0,
  `abridged_locked` tinyint(1) DEFAULT 0,
  `narrator` varchar(500) DEFAULT NULL,
  `abridged` tinyint(1) DEFAULT NULL,
  `age_rating` int(11) DEFAULT NULL,
  `content_rating` varchar(20) DEFAULT NULL,
  `age_rating_locked` tinyint(1) NOT NULL DEFAULT 0,
  `content_rating_locked` tinyint(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (`book_id`),
  CONSTRAINT `fk_book_metadata` FOREIGN KEY (`book_id`) REFERENCES `book` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `book_metadata_author_mapping`
--

DROP TABLE IF EXISTS `book_metadata_author_mapping`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `book_metadata_author_mapping` (
  `book_id` bigint(20) NOT NULL,
  `author_id` bigint(20) NOT NULL,
  PRIMARY KEY (`book_id`,`author_id`),
  KEY `idx_book_metadata_id` (`book_id`),
  KEY `idx_author_id` (`author_id`),
  CONSTRAINT `fk_book_metadata_author_mapping_author` FOREIGN KEY (`author_id`) REFERENCES `author` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_book_metadata_author_mapping_book` FOREIGN KEY (`book_id`) REFERENCES `book_metadata` (`book_id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `book_metadata_category_mapping`
--

DROP TABLE IF EXISTS `book_metadata_category_mapping`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `book_metadata_category_mapping` (
  `book_id` bigint(20) NOT NULL,
  `category_id` bigint(20) NOT NULL,
  PRIMARY KEY (`book_id`,`category_id`),
  KEY `fk_book_metadata_category_mapping_category` (`category_id`),
  CONSTRAINT `fk_book_metadata_category_mapping_book` FOREIGN KEY (`book_id`) REFERENCES `book_metadata` (`book_id`) ON DELETE CASCADE,
  CONSTRAINT `fk_book_metadata_category_mapping_category` FOREIGN KEY (`category_id`) REFERENCES `category` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `book_metadata_mood_mapping`
--

DROP TABLE IF EXISTS `book_metadata_mood_mapping`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `book_metadata_mood_mapping` (
  `book_id` bigint(20) NOT NULL,
  `mood_id` bigint(20) NOT NULL,
  PRIMARY KEY (`book_id`,`mood_id`),
  KEY `fk_book_metadata_mood_mapping_mood` (`mood_id`),
  CONSTRAINT `fk_book_metadata_mood_mapping_book` FOREIGN KEY (`book_id`) REFERENCES `book_metadata` (`book_id`) ON DELETE CASCADE,
  CONSTRAINT `fk_book_metadata_mood_mapping_mood` FOREIGN KEY (`mood_id`) REFERENCES `mood` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `book_metadata_tag_mapping`
--

DROP TABLE IF EXISTS `book_metadata_tag_mapping`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `book_metadata_tag_mapping` (
  `book_id` bigint(20) NOT NULL,
  `tag_id` bigint(20) NOT NULL,
  PRIMARY KEY (`book_id`,`tag_id`),
  KEY `fk_book_metadata_tag_mapping_tag` (`tag_id`),
  CONSTRAINT `fk_book_metadata_tag_mapping_book` FOREIGN KEY (`book_id`) REFERENCES `book_metadata` (`book_id`) ON DELETE CASCADE,
  CONSTRAINT `fk_book_metadata_tag_mapping_tag` FOREIGN KEY (`tag_id`) REFERENCES `tag` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `book_notes`
--

DROP TABLE IF EXISTS `book_notes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `book_notes` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) NOT NULL,
  `book_id` bigint(20) NOT NULL,
  `title` varchar(255) DEFAULT NULL,
  `content` text NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  PRIMARY KEY (`id`),
  KEY `idx_book_notes_user_id` (`user_id`),
  KEY `idx_book_notes_book_id` (`book_id`),
  CONSTRAINT `fk_book_notes_book_id` FOREIGN KEY (`book_id`) REFERENCES `book` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_book_notes_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `book_notes_v2`
--

DROP TABLE IF EXISTS `book_notes_v2`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `book_notes_v2` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) NOT NULL,
  `book_id` bigint(20) NOT NULL,
  `cfi` varchar(1000) NOT NULL,
  `selected_text` varchar(5000) DEFAULT NULL,
  `note_content` text NOT NULL,
  `color` varchar(20) DEFAULT NULL,
  `chapter_title` varchar(500) DEFAULT NULL,
  `version` bigint(20) NOT NULL DEFAULT 0,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_book_notes_v2_user_book_cfi` (`user_id`,`book_id`,`cfi`) USING HASH,
  KEY `idx_book_notes_v2_user_id` (`user_id`),
  KEY `idx_book_notes_v2_book_id` (`book_id`),
  KEY `idx_book_notes_v2_user_book` (`user_id`,`book_id`),
  KEY `idx_book_notes_v2_user_created` (`user_id`,`created_at`),
  CONSTRAINT `fk_book_notes_v2_book_id` FOREIGN KEY (`book_id`) REFERENCES `book` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_book_notes_v2_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `book_shelf_mapping`
--

DROP TABLE IF EXISTS `book_shelf_mapping`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `book_shelf_mapping` (
  `book_id` bigint(20) NOT NULL,
  `shelf_id` bigint(20) NOT NULL,
  PRIMARY KEY (`book_id`,`shelf_id`),
  KEY `fk_book_shelf_mapping_shelf` (`shelf_id`),
  CONSTRAINT `fk_book_shelf_mapping_book` FOREIGN KEY (`book_id`) REFERENCES `book` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_book_shelf_mapping_shelf` FOREIGN KEY (`shelf_id`) REFERENCES `shelf` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `bookdrop_file`
--

DROP TABLE IF EXISTS `bookdrop_file`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `bookdrop_file` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `file_path` text NOT NULL,
  `file_name` varchar(512) NOT NULL,
  `file_size` bigint(20) DEFAULT NULL,
  `status` varchar(20) NOT NULL DEFAULT 'PENDING_REVIEW',
  `original_metadata` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL CHECK (json_valid(`original_metadata`)),
  `fetched_metadata` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL CHECK (json_valid(`fetched_metadata`)),
  `created_at` timestamp NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_file_path` (`file_path`(255))
) ENGINE=InnoDB AUTO_INCREMENT=124 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `category`
--

DROP TABLE IF EXISTS `category`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `category` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=40 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cbx_viewer_preference`
--

DROP TABLE IF EXISTS `cbx_viewer_preference`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `cbx_viewer_preference` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) NOT NULL,
  `book_id` bigint(20) NOT NULL,
  `spread` varchar(16) DEFAULT NULL,
  `view_mode` varchar(16) DEFAULT NULL,
  `fit_mode` varchar(16) DEFAULT NULL,
  `scroll_mode` varchar(16) DEFAULT NULL,
  `background_color` varchar(16) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_id` (`user_id`,`book_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `comic_character`
--

DROP TABLE IF EXISTS `comic_character`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `comic_character` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `comic_creator`
--

DROP TABLE IF EXISTS `comic_creator`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `comic_creator` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `comic_location`
--

DROP TABLE IF EXISTS `comic_location`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `comic_location` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `comic_metadata`
--

DROP TABLE IF EXISTS `comic_metadata`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `comic_metadata` (
  `book_id` bigint(20) NOT NULL,
  `issue_number` varchar(50) DEFAULT NULL,
  `volume_name` varchar(255) DEFAULT NULL,
  `volume_number` int(11) DEFAULT NULL,
  `story_arc` varchar(255) DEFAULT NULL,
  `story_arc_number` int(11) DEFAULT NULL,
  `alternate_series` varchar(255) DEFAULT NULL,
  `alternate_issue` varchar(50) DEFAULT NULL,
  `imprint` varchar(255) DEFAULT NULL,
  `format` varchar(50) DEFAULT NULL,
  `black_and_white` tinyint(1) DEFAULT 0,
  `manga` tinyint(1) DEFAULT 0,
  `reading_direction` varchar(10) DEFAULT 'ltr',
  `web_link` varchar(1000) DEFAULT NULL,
  `notes` text DEFAULT NULL,
  `issue_number_locked` tinyint(1) DEFAULT 0,
  `volume_name_locked` tinyint(1) DEFAULT 0,
  `volume_number_locked` tinyint(1) DEFAULT 0,
  `story_arc_locked` tinyint(1) DEFAULT 0,
  `creators_locked` tinyint(1) DEFAULT 0,
  `characters_locked` tinyint(1) DEFAULT 0,
  `teams_locked` tinyint(1) DEFAULT 0,
  `locations_locked` tinyint(1) DEFAULT 0,
  `imprint_locked` tinyint(1) DEFAULT 0,
  `format_locked` tinyint(1) DEFAULT 0,
  `black_and_white_locked` tinyint(1) DEFAULT 0,
  `manga_locked` tinyint(1) DEFAULT 0,
  `reading_direction_locked` tinyint(1) DEFAULT 0,
  `web_link_locked` tinyint(1) DEFAULT 0,
  `notes_locked` tinyint(1) DEFAULT 0,
  `story_arc_number_locked` tinyint(1) DEFAULT 0,
  `alternate_series_locked` tinyint(1) DEFAULT 0,
  `alternate_issue_locked` tinyint(1) DEFAULT 0,
  `pencillers_locked` tinyint(1) DEFAULT 0,
  `inkers_locked` tinyint(1) DEFAULT 0,
  `colorists_locked` tinyint(1) DEFAULT 0,
  `letterers_locked` tinyint(1) DEFAULT 0,
  `cover_artists_locked` tinyint(1) DEFAULT 0,
  `editors_locked` tinyint(1) DEFAULT 0,
  PRIMARY KEY (`book_id`),
  KEY `idx_comic_metadata_story_arc` (`story_arc`),
  KEY `idx_comic_metadata_volume_name` (`volume_name`),
  CONSTRAINT `fk_comic_metadata_book` FOREIGN KEY (`book_id`) REFERENCES `book_metadata` (`book_id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `comic_metadata_character_mapping`
--

DROP TABLE IF EXISTS `comic_metadata_character_mapping`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `comic_metadata_character_mapping` (
  `book_id` bigint(20) NOT NULL,
  `character_id` bigint(20) NOT NULL,
  PRIMARY KEY (`book_id`,`character_id`),
  KEY `fk_comic_char_mapping_char` (`character_id`),
  CONSTRAINT `fk_comic_char_mapping_book` FOREIGN KEY (`book_id`) REFERENCES `comic_metadata` (`book_id`) ON DELETE CASCADE,
  CONSTRAINT `fk_comic_char_mapping_char` FOREIGN KEY (`character_id`) REFERENCES `comic_character` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `comic_metadata_creator_mapping`
--

DROP TABLE IF EXISTS `comic_metadata_creator_mapping`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `comic_metadata_creator_mapping` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `book_id` bigint(20) NOT NULL,
  `creator_id` bigint(20) NOT NULL,
  `role` varchar(20) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_comic_creator_mapping_creator` (`creator_id`),
  KEY `idx_comic_creator_mapping_role` (`role`),
  KEY `idx_comic_creator_mapping_book` (`book_id`),
  CONSTRAINT `fk_comic_creator_mapping_book` FOREIGN KEY (`book_id`) REFERENCES `comic_metadata` (`book_id`) ON DELETE CASCADE,
  CONSTRAINT `fk_comic_creator_mapping_creator` FOREIGN KEY (`creator_id`) REFERENCES `comic_creator` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `comic_metadata_location_mapping`
--

DROP TABLE IF EXISTS `comic_metadata_location_mapping`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `comic_metadata_location_mapping` (
  `book_id` bigint(20) NOT NULL,
  `location_id` bigint(20) NOT NULL,
  PRIMARY KEY (`book_id`,`location_id`),
  KEY `fk_comic_loc_mapping_loc` (`location_id`),
  CONSTRAINT `fk_comic_loc_mapping_book` FOREIGN KEY (`book_id`) REFERENCES `comic_metadata` (`book_id`) ON DELETE CASCADE,
  CONSTRAINT `fk_comic_loc_mapping_loc` FOREIGN KEY (`location_id`) REFERENCES `comic_location` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `comic_metadata_team_mapping`
--

DROP TABLE IF EXISTS `comic_metadata_team_mapping`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `comic_metadata_team_mapping` (
  `book_id` bigint(20) NOT NULL,
  `team_id` bigint(20) NOT NULL,
  PRIMARY KEY (`book_id`,`team_id`),
  KEY `fk_comic_team_mapping_team` (`team_id`),
  CONSTRAINT `fk_comic_team_mapping_book` FOREIGN KEY (`book_id`) REFERENCES `comic_metadata` (`book_id`) ON DELETE CASCADE,
  CONSTRAINT `fk_comic_team_mapping_team` FOREIGN KEY (`team_id`) REFERENCES `comic_team` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `comic_team`
--

DROP TABLE IF EXISTS `comic_team`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `comic_team` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `custom_font`
--

DROP TABLE IF EXISTS `custom_font`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `custom_font` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) NOT NULL,
  `font_name` varchar(255) NOT NULL,
  `file_name` varchar(255) NOT NULL,
  `original_file_name` varchar(255) NOT NULL,
  `format` varchar(10) NOT NULL,
  `file_size` bigint(20) NOT NULL,
  `uploaded_at` timestamp NOT NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`id`),
  UNIQUE KEY `file_name` (`file_name`),
  KEY `idx_user_id` (`user_id`),
  CONSTRAINT `custom_font_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `ebook_viewer_preference`
--

DROP TABLE IF EXISTS `ebook_viewer_preference`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `ebook_viewer_preference` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) NOT NULL,
  `book_id` bigint(20) NOT NULL,
  `font_family` varchar(128) DEFAULT NULL,
  `font_size` int(11) DEFAULT NULL,
  `gap` float DEFAULT NULL,
  `hyphenate` tinyint(1) DEFAULT NULL,
  `is_dark` tinyint(1) DEFAULT NULL,
  `justify` tinyint(1) DEFAULT NULL,
  `line_height` float DEFAULT NULL,
  `max_block_size` int(11) DEFAULT NULL,
  `max_column_count` int(11) DEFAULT NULL,
  `max_inline_size` int(11) DEFAULT NULL,
  `theme` varchar(64) DEFAULT NULL,
  `flow` varchar(32) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_id` (`user_id`,`book_id`),
  KEY `fk_ebook_viewer_preference_book` (`book_id`),
  CONSTRAINT `fk_ebook_viewer_preference_book` FOREIGN KEY (`book_id`) REFERENCES `book` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_ebook_viewer_preference_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `email_provider`
--

DROP TABLE IF EXISTS `email_provider`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `email_provider` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `host` varchar(255) NOT NULL,
  `port` int(11) NOT NULL,
  `username` varchar(255) NOT NULL,
  `password` varchar(255) NOT NULL,
  `from_address` varchar(255) DEFAULT NULL,
  `auth` tinyint(1) NOT NULL,
  `start_tls` tinyint(1) NOT NULL,
  `is_default` tinyint(1) NOT NULL DEFAULT 0,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`,`host`,`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `email_provider_v2`
--

DROP TABLE IF EXISTS `email_provider_v2`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `email_provider_v2` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) NOT NULL,
  `name` varchar(255) NOT NULL,
  `host` varchar(255) NOT NULL,
  `port` int(11) NOT NULL,
  `username` varchar(255) NOT NULL,
  `password` varchar(255) NOT NULL,
  `from_address` varchar(255) DEFAULT NULL,
  `auth` tinyint(1) NOT NULL,
  `start_tls` tinyint(1) NOT NULL,
  `is_default` tinyint(1) NOT NULL DEFAULT 0,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `shared` tinyint(1) DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_id` (`user_id`,`name`),
  KEY `idx_email_provider_v2_user_id` (`user_id`),
  CONSTRAINT `fk_email_provider_v2_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `email_recipient`
--

DROP TABLE IF EXISTS `email_recipient`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `email_recipient` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `email` varchar(255) NOT NULL,
  `name` varchar(255) NOT NULL,
  `is_default` tinyint(1) NOT NULL DEFAULT 0,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`id`),
  UNIQUE KEY `email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `email_recipient_v2`
--

DROP TABLE IF EXISTS `email_recipient_v2`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `email_recipient_v2` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) NOT NULL,
  `email` varchar(255) NOT NULL,
  `name` varchar(255) NOT NULL,
  `is_default` tinyint(1) NOT NULL DEFAULT 0,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_id` (`user_id`,`email`),
  KEY `idx_email_recipient_v2_user_id` (`user_id`),
  CONSTRAINT `fk_email_recipient_v2_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `epub_viewer_preference`
--

DROP TABLE IF EXISTS `epub_viewer_preference`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `epub_viewer_preference` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) NOT NULL,
  `book_id` bigint(20) NOT NULL,
  `theme` varchar(255) DEFAULT NULL,
  `font` varchar(255) DEFAULT NULL,
  `font_size` int(11) DEFAULT NULL,
  `flow` varchar(32) DEFAULT NULL,
  `letter_spacing` float DEFAULT NULL,
  `line_height` float DEFAULT NULL,
  `spread` varchar(20) DEFAULT 'double',
  `custom_font_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_id` (`user_id`,`book_id`),
  KEY `custom_font_id` (`custom_font_id`),
  CONSTRAINT `epub_viewer_preference_ibfk_1` FOREIGN KEY (`custom_font_id`) REFERENCES `custom_font` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `flyway_schema_history`
--

DROP TABLE IF EXISTS `flyway_schema_history`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `flyway_schema_history` (
  `installed_rank` int(11) NOT NULL,
  `version` varchar(50) DEFAULT NULL,
  `description` varchar(200) NOT NULL,
  `type` varchar(20) NOT NULL,
  `script` varchar(1000) NOT NULL,
  `checksum` int(11) DEFAULT NULL,
  `installed_by` varchar(100) NOT NULL,
  `installed_on` timestamp NOT NULL DEFAULT current_timestamp(),
  `execution_time` int(11) NOT NULL,
  `success` tinyint(1) NOT NULL,
  PRIMARY KEY (`installed_rank`),
  KEY `flyway_schema_history_s_idx` (`success`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `jwt_secret`
--

DROP TABLE IF EXISTS `jwt_secret`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `jwt_secret` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `secret` text NOT NULL,
  `created_at` timestamp NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `kobo_library_snapshot`
--

DROP TABLE IF EXISTS `kobo_library_snapshot`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `kobo_library_snapshot` (
  `id` varchar(36) NOT NULL,
  `user_id` bigint(20) NOT NULL,
  `created_date` timestamp NOT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_snapshot_user` (`user_id`),
  CONSTRAINT `fk_snapshot_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `kobo_library_snapshot_book`
--

DROP TABLE IF EXISTS `kobo_library_snapshot_book`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `kobo_library_snapshot_book` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `snapshot_id` varchar(36) NOT NULL,
  `book_id` bigint(20) NOT NULL,
  `synced` tinyint(1) NOT NULL DEFAULT 0,
  `file_hash` varchar(255) DEFAULT NULL,
  `metadata_updated_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_snapshot_book` (`snapshot_id`,`book_id`),
  CONSTRAINT `fk_snapshot_book` FOREIGN KEY (`snapshot_id`) REFERENCES `kobo_library_snapshot` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `kobo_reading_state`
--

DROP TABLE IF EXISTS `kobo_reading_state`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `kobo_reading_state` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `entitlement_id` varchar(255) NOT NULL,
  `created` varchar(255) DEFAULT NULL,
  `last_modified` varchar(255) DEFAULT NULL,
  `priority_timestamp` varchar(255) DEFAULT NULL,
  `current_bookmark_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL CHECK (json_valid(`current_bookmark_json`)),
  `statistics_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL CHECK (json_valid(`statistics_json`)),
  `status_info_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL CHECK (json_valid(`status_info_json`)),
  `last_modified_string` varchar(255) DEFAULT NULL,
  `user_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_kobo_reading_state_user_entitlement` (`user_id`,`entitlement_id`),
  KEY `idx_kobo_reading_state_entitlement_id` (`entitlement_id`),
  CONSTRAINT `fk_kobo_reading_state_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `kobo_removed_books_tracking`
--

DROP TABLE IF EXISTS `kobo_removed_books_tracking`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `kobo_removed_books_tracking` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `snapshot_id` varchar(36) NOT NULL,
  `user_id` bigint(20) NOT NULL,
  `book_id_synced` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_snapshot_user_book` (`snapshot_id`,`user_id`,`book_id_synced`),
  CONSTRAINT `fk_removed_snapshot` FOREIGN KEY (`snapshot_id`) REFERENCES `kobo_library_snapshot` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `kobo_user_settings`
--

DROP TABLE IF EXISTS `kobo_user_settings`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `kobo_user_settings` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) NOT NULL,
  `token` varchar(2048) NOT NULL,
  `sync_enabled` tinyint(1) NOT NULL DEFAULT 1,
  `progress_mark_as_reading_threshold` float DEFAULT 1,
  `progress_mark_as_finished_threshold` float DEFAULT 99,
  `auto_add_to_shelf` tinyint(1) NOT NULL DEFAULT 0,
  `hardcover_api_key` varchar(2048) DEFAULT NULL,
  `hardcover_sync_enabled` tinyint(1) DEFAULT 0,
  `two_way_progress_sync` tinyint(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_id` (`user_id`),
  CONSTRAINT `fk_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `koreader_user`
--

DROP TABLE IF EXISTS `koreader_user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `koreader_user` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `username` varchar(100) NOT NULL,
  `password` varchar(255) NOT NULL,
  `password_md5` varchar(255) NOT NULL,
  `sync_enabled` tinyint(1) DEFAULT NULL,
  `booklore_user_id` bigint(20) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  `sync_with_booklore_reader` tinyint(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `username` (`username`),
  KEY `fk_booklore_user` (`booklore_user_id`),
  CONSTRAINT `fk_booklore_user` FOREIGN KEY (`booklore_user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `library`
--

DROP TABLE IF EXISTS `library`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `library` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `sort` varchar(255) DEFAULT NULL,
  `icon` varchar(64) DEFAULT NULL,
  `watch` tinyint(1) NOT NULL DEFAULT 0,
  `file_naming_pattern` varchar(1000) DEFAULT NULL,
  `icon_type` varchar(255) DEFAULT NULL,
  `format_priority` text DEFAULT NULL,
  `organization_mode` varchar(50) NOT NULL DEFAULT 'AUTO_DETECT',
  `allowed_formats` text DEFAULT NULL,
  `metadata_source` varchar(20) DEFAULT 'EMBEDDED',
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `library_path`
--

DROP TABLE IF EXISTS `library_path`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `library_path` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `path` text DEFAULT NULL,
  `library_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_library_path` (`library_id`),
  CONSTRAINT `fk_library_path` FOREIGN KEY (`library_id`) REFERENCES `library` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `magic_shelf`
--

DROP TABLE IF EXISTS `magic_shelf`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `magic_shelf` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) NOT NULL,
  `name` varchar(255) NOT NULL,
  `icon` varchar(64) DEFAULT NULL,
  `filter_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL CHECK (json_valid(`filter_json`)),
  `created_at` timestamp NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  `is_public` tinyint(1) NOT NULL DEFAULT 0,
  `icon_type` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_user_name` (`user_id`,`name`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `metadata_fetch_jobs`
--

DROP TABLE IF EXISTS `metadata_fetch_jobs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `metadata_fetch_jobs` (
  `task_id` varchar(100) NOT NULL,
  `user_id` bigint(20) DEFAULT NULL,
  `status` varchar(20) NOT NULL DEFAULT 'pending',
  `status_message` text DEFAULT NULL,
  `started_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `completed_at` timestamp NULL DEFAULT NULL,
  `total_books_count` int(11) DEFAULT NULL,
  `completed_books` int(11) DEFAULT 0,
  PRIMARY KEY (`task_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `metadata_fetch_proposals`
--

DROP TABLE IF EXISTS `metadata_fetch_proposals`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `metadata_fetch_proposals` (
  `proposal_id` bigint(20) NOT NULL AUTO_INCREMENT,
  `task_id` varchar(100) NOT NULL,
  `book_id` bigint(20) NOT NULL,
  `fetched_at` timestamp NULL DEFAULT current_timestamp(),
  `reviewed_at` timestamp NULL DEFAULT NULL,
  `reviewer_user_id` bigint(20) DEFAULT NULL,
  `status` varchar(30) NOT NULL DEFAULT 'pending',
  `metadata_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL CHECK (json_valid(`metadata_json`)),
  PRIMARY KEY (`proposal_id`),
  KEY `idx_metadata_proposal_task_id` (`task_id`),
  KEY `idx_metadata_proposal_book_id` (`book_id`),
  KEY `idx_metadata_proposal_status` (`status`),
  CONSTRAINT `fk_metadata_fetch_task` FOREIGN KEY (`task_id`) REFERENCES `metadata_fetch_jobs` (`task_id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `mood`
--

DROP TABLE IF EXISTS `mood`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `mood` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `new_pdf_viewer_preference`
--

DROP TABLE IF EXISTS `new_pdf_viewer_preference`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `new_pdf_viewer_preference` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) NOT NULL,
  `book_id` bigint(20) NOT NULL,
  `spread` varchar(16) DEFAULT NULL,
  `view_mode` varchar(16) DEFAULT NULL,
  `fit_mode` varchar(16) DEFAULT NULL,
  `scroll_mode` varchar(16) DEFAULT NULL,
  `background_color` varchar(16) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_id` (`user_id`,`book_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `opds_user`
--

DROP TABLE IF EXISTS `opds_user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `opds_user` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `username` varchar(255) NOT NULL,
  `password` varchar(255) NOT NULL,
  `created_at` timestamp NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`id`),
  UNIQUE KEY `username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `opds_user_v2`
--

DROP TABLE IF EXISTS `opds_user_v2`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `opds_user_v2` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) NOT NULL,
  `username` varchar(100) NOT NULL,
  `password_hash` varchar(255) NOT NULL,
  `created_at` timestamp NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  `sort_order` varchar(20) NOT NULL DEFAULT 'RECENT',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_username` (`username`),
  KEY `fk_opds_user` (`user_id`),
  CONSTRAINT `fk_opds_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pdf_annotations`
--

DROP TABLE IF EXISTS `pdf_annotations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `pdf_annotations` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) NOT NULL,
  `book_id` bigint(20) NOT NULL,
  `data` longtext NOT NULL,
  `version` bigint(20) NOT NULL DEFAULT 0,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_id` (`user_id`,`book_id`),
  KEY `idx_pdf_annotations_user_id` (`user_id`),
  KEY `idx_pdf_annotations_book_id` (`book_id`),
  CONSTRAINT `fk_pdf_annotations_book` FOREIGN KEY (`book_id`) REFERENCES `book` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_pdf_annotations_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pdf_viewer_preference`
--

DROP TABLE IF EXISTS `pdf_viewer_preference`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `pdf_viewer_preference` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) NOT NULL,
  `book_id` bigint(20) NOT NULL,
  `zoom` varchar(64) DEFAULT NULL,
  `spread` varchar(64) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_id` (`user_id`,`book_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `public_book_review`
--

DROP TABLE IF EXISTS `public_book_review`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `public_book_review` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `metadata_provider` varchar(255) NOT NULL,
  `book_id` bigint(20) NOT NULL,
  `reviewer_name` varchar(512) DEFAULT NULL,
  `title` varchar(512) DEFAULT NULL,
  `rating` float DEFAULT NULL,
  `date` timestamp NULL DEFAULT NULL,
  `body` text DEFAULT NULL,
  `country` varchar(255) DEFAULT NULL,
  `spoiler` tinyint(1) DEFAULT NULL,
  `followers_count` int(11) DEFAULT NULL,
  `text_reviews_count` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `book_id` (`book_id`),
  CONSTRAINT `public_book_review_ibfk_1` FOREIGN KEY (`book_id`) REFERENCES `book_metadata` (`book_id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `reading_sessions`
--

DROP TABLE IF EXISTS `reading_sessions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `reading_sessions` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) NOT NULL,
  `book_id` bigint(20) NOT NULL,
  `book_type` varchar(10) NOT NULL,
  `start_time` datetime NOT NULL,
  `end_time` datetime NOT NULL,
  `duration_seconds` int(11) NOT NULL,
  `start_progress` float DEFAULT NULL,
  `end_progress` float DEFAULT NULL,
  `progress_delta` float DEFAULT NULL,
  `start_location` varchar(500) DEFAULT NULL,
  `end_location` varchar(500) DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT current_timestamp(),
  `duration_formatted` varchar(50) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_reading_session_user_time` (`user_id`,`start_time` DESC),
  KEY `idx_reading_session_book` (`book_id`,`start_time` DESC),
  KEY `idx_reading_session_user_book` (`user_id`,`book_id`,`start_time` DESC),
  CONSTRAINT `fk_reading_session_book` FOREIGN KEY (`book_id`) REFERENCES `book` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_reading_session_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `refresh_token`
--

DROP TABLE IF EXISTS `refresh_token`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `refresh_token` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) NOT NULL,
  `token` varchar(512) NOT NULL,
  `expiry_date` datetime(6) NOT NULL,
  `revoked` tinyint(1) NOT NULL DEFAULT 0,
  `revocation_date` datetime(6) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_refresh_token` (`token`),
  KEY `fk_refresh_token_user` (`user_id`),
  CONSTRAINT `fk_refresh_token_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `shelf`
--

DROP TABLE IF EXISTS `shelf`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `shelf` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) NOT NULL,
  `name` varchar(255) NOT NULL,
  `sort` varchar(255) DEFAULT NULL,
  `icon` varchar(64) DEFAULT NULL,
  `icon_type` varchar(255) DEFAULT NULL,
  `is_public` tinyint(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_id` (`user_id`,`name`),
  CONSTRAINT `fk_shelf_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tag`
--

DROP TABLE IF EXISTS `tag`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `tag` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `task_cron_configuration`
--

DROP TABLE IF EXISTS `task_cron_configuration`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `task_cron_configuration` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `task_type` varchar(100) NOT NULL,
  `cron_expression` varchar(100) NOT NULL,
  `enabled` tinyint(1) NOT NULL DEFAULT 1,
  `created_by` bigint(20) NOT NULL DEFAULT -1,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_task_type` (`task_type`)
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tasks`
--

DROP TABLE IF EXISTS `tasks`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `tasks` (
  `id` varchar(36) NOT NULL,
  `type` varchar(50) NOT NULL,
  `status` varchar(50) NOT NULL,
  `user_id` bigint(20) NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime DEFAULT NULL,
  `completed_at` datetime DEFAULT NULL,
  `progress_percentage` int(11) DEFAULT NULL,
  `message` text DEFAULT NULL,
  `errorDetails` text DEFAULT NULL,
  `task_options` text DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_tasks_user_id` (`user_id`),
  KEY `idx_tasks_type` (`type`),
  KEY `idx_tasks_status` (`status`),
  KEY `idx_tasks_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `user_book_file_progress`
--

DROP TABLE IF EXISTS `user_book_file_progress`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_book_file_progress` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) NOT NULL,
  `book_file_id` bigint(20) NOT NULL,
  `position_data` varchar(1000) DEFAULT NULL,
  `position_href` varchar(1000) DEFAULT NULL,
  `progress_percent` float DEFAULT NULL,
  `last_read_time` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_book_file` (`user_id`,`book_file_id`),
  KEY `fk_ubfp_book_file` (`book_file_id`),
  KEY `idx_ubfp_user_book_file` (`user_id`,`book_file_id`),
  CONSTRAINT `fk_ubfp_book_file` FOREIGN KEY (`book_file_id`) REFERENCES `book_file` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_ubfp_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `user_book_progress`
--

DROP TABLE IF EXISTS `user_book_progress`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_book_progress` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) NOT NULL,
  `book_id` bigint(20) NOT NULL,
  `last_read_time` timestamp NULL DEFAULT NULL,
  `pdf_progress` int(11) DEFAULT NULL,
  `epub_progress` varchar(1000) DEFAULT NULL,
  `pdf_progress_percent` float DEFAULT NULL,
  `epub_progress_percent` float DEFAULT NULL,
  `cbx_progress` int(11) DEFAULT NULL,
  `cbx_progress_percent` float DEFAULT NULL,
  `read_status` varchar(20) DEFAULT NULL,
  `date_finished` timestamp NULL DEFAULT NULL,
  `koreader_progress` varchar(1000) DEFAULT NULL,
  `koreader_progress_percent` float DEFAULT NULL,
  `koreader_device` varchar(100) DEFAULT NULL,
  `koreader_device_id` varchar(100) DEFAULT NULL,
  `koreader_last_sync_time` timestamp NULL DEFAULT NULL,
  `kobo_progress_percent` float DEFAULT NULL,
  `kobo_location` varchar(1000) DEFAULT NULL,
  `kobo_location_type` varchar(50) DEFAULT NULL,
  `kobo_location_source` varchar(512) DEFAULT NULL,
  `kobo_progress_received_time` timestamp NULL DEFAULT NULL,
  `kobo_status_sent_time` timestamp NULL DEFAULT NULL,
  `read_status_modified_time` timestamp NULL DEFAULT NULL,
  `kobo_progress_sent_time` timestamp NULL DEFAULT NULL,
  `personal_rating` tinyint(4) DEFAULT NULL,
  `epub_progress_href` varchar(1000) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_user_book_progress` (`user_id`,`book_id`),
  KEY `idx_user_book_progress_user` (`user_id`),
  KEY `idx_user_book_progress_book` (`book_id`),
  KEY `idx_user_book_progress_date_finished` (`date_finished`),
  CONSTRAINT `fk_user_book_progress_book` FOREIGN KEY (`book_id`) REFERENCES `book` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_user_book_progress_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=36 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `user_content_restriction`
--

DROP TABLE IF EXISTS `user_content_restriction`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_content_restriction` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) NOT NULL,
  `restriction_type` varchar(20) NOT NULL,
  `mode` varchar(15) NOT NULL,
  `value` varchar(255) NOT NULL,
  `created_at` timestamp NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_restriction` (`user_id`,`restriction_type`,`value`),
  KEY `idx_ucr_user_id` (`user_id`),
  CONSTRAINT `fk_ucr_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `user_email_provider_preference`
--

DROP TABLE IF EXISTS `user_email_provider_preference`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_email_provider_preference` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) NOT NULL,
  `default_provider_id` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_user_id` (`user_id`),
  KEY `fk_user_email_preference_provider` (`default_provider_id`),
  CONSTRAINT `fk_user_email_preference_provider` FOREIGN KEY (`default_provider_id`) REFERENCES `email_provider_v2` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_user_email_preference_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `user_library_mapping`
--

DROP TABLE IF EXISTS `user_library_mapping`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_library_mapping` (
  `user_id` bigint(20) NOT NULL,
  `library_id` bigint(20) NOT NULL,
  PRIMARY KEY (`user_id`,`library_id`),
  KEY `fk_user_library_mapping_library` (`library_id`),
  CONSTRAINT `fk_user_library_mapping_library` FOREIGN KEY (`library_id`) REFERENCES `library` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_user_library_mapping_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `user_permissions`
--

DROP TABLE IF EXISTS `user_permissions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_permissions` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) NOT NULL,
  `permission_upload` tinyint(1) NOT NULL DEFAULT 0,
  `permission_download` tinyint(1) NOT NULL DEFAULT 0,
  `permission_edit_metadata` tinyint(1) NOT NULL DEFAULT 0,
  `permission_manipulate_library` tinyint(1) NOT NULL DEFAULT 0,
  `permission_admin` tinyint(1) NOT NULL DEFAULT 0,
  `permission_email_book` tinyint(1) NOT NULL DEFAULT 1,
  `permission_delete_book` tinyint(1) NOT NULL DEFAULT 0,
  `permission_sync_koreader` tinyint(1) NOT NULL DEFAULT 0,
  `permission_sync_kobo` tinyint(1) NOT NULL DEFAULT 0,
  `permission_access_opds` tinyint(1) NOT NULL DEFAULT 0,
  `permission_manage_metadata_config` tinyint(1) NOT NULL DEFAULT 0,
  `permission_access_bookdrop` tinyint(1) NOT NULL DEFAULT 0,
  `permission_access_library_stats` tinyint(1) NOT NULL DEFAULT 0,
  `permission_access_user_stats` tinyint(1) NOT NULL DEFAULT 0,
  `permission_access_task_manager` tinyint(1) NOT NULL DEFAULT 0,
  `permission_manage_global_preferences` tinyint(1) NOT NULL DEFAULT 0,
  `permission_manage_icons` tinyint(1) NOT NULL DEFAULT 0,
  `permission_demo_user` tinyint(1) NOT NULL DEFAULT 0,
  `permission_bulk_auto_fetch_metadata` tinyint(1) NOT NULL DEFAULT 0,
  `permission_bulk_custom_fetch_metadata` tinyint(1) NOT NULL DEFAULT 0,
  `permission_bulk_edit_metadata` tinyint(1) NOT NULL DEFAULT 0,
  `permission_bulk_regenerate_cover` tinyint(1) NOT NULL DEFAULT 0,
  `permission_move_organize_files` tinyint(1) NOT NULL DEFAULT 0,
  `permission_bulk_lock_unlock_metadata` tinyint(1) NOT NULL DEFAULT 0,
  `permission_bulk_reset_booklore_read_progress` tinyint(1) NOT NULL DEFAULT 0,
  `permission_bulk_reset_koreader_read_progress` tinyint(1) NOT NULL DEFAULT 0,
  `permission_bulk_reset_book_read_status` tinyint(1) NOT NULL DEFAULT 0,
  `permission_manage_fonts` tinyint(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_user_permissions_user` (`user_id`),
  CONSTRAINT `fk_user_permissions_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `user_settings`
--

DROP TABLE IF EXISTS `user_settings`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_settings` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) NOT NULL,
  `setting_key` varchar(100) NOT NULL,
  `setting_value` text NOT NULL,
  `created_at` timestamp NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_id` (`user_id`,`setting_key`),
  CONSTRAINT `user_settings_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=15 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `users`
--

DROP TABLE IF EXISTS `users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `users` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `username` varchar(255) NOT NULL,
  `password_hash` varchar(255) NOT NULL,
  `is_default_password` tinyint(1) NOT NULL DEFAULT 1,
  `name` varchar(255) NOT NULL,
  `email` varchar(255) DEFAULT NULL,
  `book_preferences` text DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `provisioning_method` varchar(50) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `username` (`username`),
  UNIQUE KEY `email` (`email`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*M!100616 SET NOTE_VERBOSITY=@OLD_NOTE_VERBOSITY */;

-- Dump completed on 2026-03-24 14:39:44
