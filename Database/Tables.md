---
id: Tables
aliases: []
tags:
  - coding
  - database
dg-publish: true
---
# Tables

1. Create A table

```sql
create table <table_name>

```

2. Delete a table 

```sql 
DROP TABLE <table_name>;

```


```sql
CREATE TABLE Customers (
    CustomerID INT PRIMARY KEY,
    FirstName VARCHAR(50) NOT NULL,
    LastName VARCHAR(50) NOT NULL,
    Address VARCHAR(255),
    City VARCHAR(255)
);
```

