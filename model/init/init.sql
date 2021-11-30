CREATE DATABASE  IF NOT EXISTS `cryptocompare` /*!40100 DEFAULT CHARACTER SET utf8 */;
USE `cryptocompare`;
-- MySQL dump 10.13  Distrib 5.7.36, for Linux (x86_64)
--
-- Host: localhost    Database: cryptocompare
-- ------------------------------------------------------
-- Server version	5.7.36-0ubuntu0.18.04.1

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
-- Table structure for table `data`
--

DROP TABLE IF EXISTS `data`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `data` (
  `_id` int(11) NOT NULL AUTO_INCREMENT,
  `fromSym` int(11) NOT NULL,
  `toSym` int(11) NOT NULL,
  `change24hour` double DEFAULT NULL,
  `changepct24hour` double DEFAULT NULL,
  `open24hour` double DEFAULT NULL,
  `volume24hour` double DEFAULT NULL,
  `low24hour` double DEFAULT NULL,
  `high24hour` double DEFAULT NULL,
  `price` double DEFAULT NULL,
  `supply` double DEFAULT NULL,
  `mktcap` double DEFAULT NULL,
  `lastupdate` mediumtext NOT NULL,
  `displaydataraw` text,
  PRIMARY KEY (`_id`),
  KEY `fk_tsym_idx` (`toSym`),
  KEY `fk_fsym_idx` (`fromSym`),
  CONSTRAINT `fk_fsym` FOREIGN KEY (`fromSym`) REFERENCES `fsym` (`_id`) ON DELETE NO ACTION ON UPDATE NO ACTION,
  CONSTRAINT `fk_tsym` FOREIGN KEY (`toSym`) REFERENCES `tsym` (`_id`) ON DELETE NO ACTION ON UPDATE NO ACTION
) ENGINE=InnoDB AUTO_INCREMENT=8628 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `data`
--

LOCK TABLES `data` WRITE;
/*!40000 ALTER TABLE `data` DISABLE KEYS */;
/*!40000 ALTER TABLE `data` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `fsym`
--

DROP TABLE IF EXISTS `fsym`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fsym` (
  `_id` int(11) NOT NULL AUTO_INCREMENT,
  `symbol` varchar(5) DEFAULT NULL,
  `unicode` char(1) DEFAULT NULL,
  PRIMARY KEY (`_id`)
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `fsym`
--

LOCK TABLES `fsym` WRITE;
/*!40000 ALTER TABLE `fsym` DISABLE KEYS */;
INSERT INTO `fsym` VALUES (1,'BTC','Ƀ'),(2,'XRP',NULL),(3,'ETH','Ξ'),(4,'BCH',NULL),(5,'EOS',NULL),(6,'LTC','Ł'),(7,'XMR',NULL),(8,'DASH',NULL);
/*!40000 ALTER TABLE `fsym` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `tsym`
--

DROP TABLE IF EXISTS `tsym`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tsym` (
  `_id` int(11) NOT NULL AUTO_INCREMENT,
  `symbol` varchar(5) DEFAULT NULL,
  `unicode` char(1) DEFAULT NULL,
  PRIMARY KEY (`_id`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `tsym`
--

LOCK TABLES `tsym` WRITE;
/*!40000 ALTER TABLE `tsym` DISABLE KEYS */;
INSERT INTO `tsym` VALUES (1,'USD','$'),(2,'EUR','€'),(3,'GBP','£'),(4,'JPY','¥'),(5,'RUR','₽');
/*!40000 ALTER TABLE `tsym` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2021-11-30  0:55:46
