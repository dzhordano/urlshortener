# yeah-yeah url shortener

Basically, it is somewhat of a URL shortener.

## run 

**prerequisites**:  

- Docker (docker compose)
- Go 1.25.4
- Make (optional)

`make run` or `docker compose up --remove-orphans -d`

http docs are in `/api` dir

#### todos and possible improvements

- `_test` postfix in test files (increasing some tech debt here)
- http server configuration (and impl https i guess)
- graceful shutdown (improve timeouts and health checks?)
- postgres setup (and redis?)
- possibility to run profiler
- cors issue with swagger ui (specifically the 'redirect' method)
- grafana dashboards are kind of bad ;_;

#### some faq

- linter `exhaustruct` and `testpackage` in test files are suppressed (annoying ashell)
