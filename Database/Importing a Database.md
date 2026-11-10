1. Obtain a database `something.gz`
2. Extract it `something.sql`
3. Log in to mysql
```cmd
mysql -u root -p
```
4. Create a database 
```sql
CREATE DATABASE <any_name> CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

5. Use the database using the `use` keyward `use any_name`
6. Now export it using the source 

```sql
SOURCE <sql_file_path>
```