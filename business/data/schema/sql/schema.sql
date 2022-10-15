-- Description: Create table movies
CREATE TABLE movies (
id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
name VARCHAR(100) NOT NULL,
description TEXT ,
year TEXT,
pageUrl TEXT,
imageUrl TEXT,
downloadLinks Text,
categories Text
);