### Prod Tools

#### Entering a Container

##### Database

* Log into the psql service directly:

```
docker exec -it -e PGPASSWORD=${POSTGRES_PASSWORD} socialpredict-postgres-container psql -U ${POSTGRES_USER} -d socialpredict_db
```

