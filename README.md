# omgili-importer

Simple tool to import (Omgili)[http://omgili.com/] data feed files to a redis DB.

## Tool Flow

1. Download all zip files from http://bitly.com/nuvi-plz.
1. Extract the xml files from each zip file.
1. Publish the content of each xml file to a redis list called "NEWS_XML".

**PS** The whole flow is idempotent. So running multiple times will not duplicate the data.
