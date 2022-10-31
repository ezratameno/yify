-- Description: Create table movies
    CREATE TABLE IF NOT EXISTS movies (
    id  SERIAL ,
    name TEXT,
    description TEXT,
    year TEXT,
    pageUrl TEXT,
    imageUrl TEXT,
    downloadLinks Text,
    categories Text,
    PRIMARY KEY (id)
    );


