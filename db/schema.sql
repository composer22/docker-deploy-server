CREATE DATABASE  IF NOT EXISTS `docker_deploy_server` /*!40100 DEFAULT CHARACTER SET utf8 */;
USE `docker_deploy_server`;
-- MySQL dump 10.13  Distrib 5.6.24, for osx10.8 (x86_64)

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `auth_tokens`
--

DROP TABLE IF EXISTS `auth_tokens`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `auth_tokens` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'Primary key for each entry in the table.',
  `token` varchar(255) NOT NULL COMMENT 'Oauth or bearer token assigned to this user/service.',
  `created_at` datetime NOT NULL COMMENT 'Create date for this row.',
  `updated_at` datetime NOT NULL COMMENT 'Last update date for this row.',
  `name` varchar(255) DEFAULT NULL COMMENT 'Name of the user or service that has been granted authority.',
  `notes` text COMMENT 'General comments.',
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_UNIQUE` (`id`),
  UNIQUE KEY `token_UNIQUE` (`token`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `environments`
--

DROP TABLE IF EXISTS `environments`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `environments` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'Primary key for each entry in the table.',
  `name` varchar(255) NOT NULL COMMENT 'Name of the environment.',
  `created_at` datetime NOT NULL COMMENT 'Create date for this row.',
  `updated_at` datetime NOT NULL COMMENT 'Last update date for this row.',
  `notes` text COMMENT 'General comments.',
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_UNIQUE` (`id`),
  UNIQUE KEY `name_UNIQUE` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `environments`
--

DROP TABLE IF EXISTS `auth_tokens_environments`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `auth_tokens_environments` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'Primary key for each entry in the table.',
  `auth_token_id` int(11) NOT NULL COMMENT 'Foreign key for an auth token.',
  `environment_id` int(11) NOT NULL COMMENT 'Foreign key for an environment.',
  `created_at` datetime NOT NULL COMMENT 'Create date for this row.',
  `updated_at` datetime NOT NULL COMMENT 'Last update date for this row.',
  `notes` text COMMENT 'General comments.',
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_UNIQUE` (`id`),
  UNIQUE KEY `token_env_UNIQUE` (`auth_token_id`,`environment_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `deploys`
--

DROP TABLE IF EXISTS `deploys`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
/* note status: Started = 1, Failed = 2, Success = 3 */;
CREATE TABLE `deploys` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'Unique identifier for each row.',
  `deploy_id` varchar(255) NOT NULL COMMENT 'UUID assigned to this deployment.',
  `environment` varchar(255) NOT NULL COMMENT 'Environment deployed, for example: dev, stage, qa, prod.',
  `image_name` varchar(255) NOT NULL COMMENT 'Repository name being deployed, for example acme-video-mobile',
  `image_tag` varchar(255) NOT NULL COMMENT 'Version of the service being deployed e.g. 1.0.0-131, latest',
  `status` int(11) NOT NULL DEFAULT '1' COMMENT 'Current status of the deploy: Queued, Started, Failed, Success.',
  `message` varchar(255) DEFAULT NULL COMMENT 'A short status message.',
  `log` text COMMENT 'Complete set of log messages from the deploy.',
  `updated_at` datetime NOT NULL COMMENT 'The update date and time of the deploy.',
  `created_at` datetime NOT NULL COMMENT 'The create date and time of the deploy.',
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_UNIQUE` (`id`),
  UNIQUE KEY `key_UNIQUE` (`deploy_id`)
) ENGINE=InnoDB AUTO_INCREMENT=31 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

