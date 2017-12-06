# uWSGI Prometheus Exporter

Fork of https://github.com/micktwomey/uwsgi_exporter

Reads the uWSGI stats socket and exports metrics.

# Usage

`-uwsgi-stats-adddress` handles:

- `file:///path/to/file.json`
- `fileglob:///path/to/files*.json`
- `unix:///path/to/file.sock`
- `unixglob:///path/to/files*.sock`

## UNIX Socket

Start uwsgi with something like:

```
uwsgi --stats /tmp/uwsgi.sock --http :9090 --wsgi-file uwsgi_app.py
```

Then start the exporter:

```
uwsgi_exporter -listen-address localhost:9131 -uwsgi-stats-address unix:///tmp/uwsgi.sock

```

### Using Docker

```
docker run -v /tmp/uwsgi_stats.sock:/tmp/uwsgi_stats.sock -p 9131:9131 -ti amitsaha/uwsgi_exporter
```
