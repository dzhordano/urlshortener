# yeah-yeah url shortener

Basically, it is somewhat of a URL shortener.

## run

**prerequisites**:

- Docker (docker compose)
- Go 1.25.4
- Make (optional)

`make run` or `docker compose up --remove-orphans -d`

http docs are in `/api` dir

### some obvious improvements

- making so shortened url info is shown to its creator only (just 'admin' api key header rn)
- more tests
- better lifecycle and app configuration (better configs, timeouts, shutdown, etc.)
- adding profiling
- redirect method via swagger doesn't work
