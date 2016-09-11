# omgili-importer

Simple tool to import [Omgili](http://omgili.com/) data feed files to a redis DB.

## Tool Flow

1. Download all zip files from http://bitly.com/nuvi-plz.
1. Extract the xml files from each zip file.
1. Publish the content of each xml file to a redis list called "NEWS_XML".

**The whole flow is idempotent. So running multiple times will not duplicate the data.**

## Usage

### Args

* **--disable-download:** Disable downloads. Useful to run over pre fetched zip files
* **--redis-address [string]:** Redis address (default "localhost:6379")
* **--redis-database [int]:** Redis database (default 0)
* **--redis-password [string]:** Redis password
* **--url [string]:** URL for feed list (default "http://bitly.com/nuvi-plz")

## Development

### Installing dependencies

```bash
go get -v ./...
```

### Running

```bash
go run main.go
```

### Testing

Integration enabled will fetch data from the internet:

```bash
go test -v --integration
```

Only unit tests not depending in external resources:

```bash
go test -v
```
