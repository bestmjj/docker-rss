## About

**docker-rss** is a server written in [Go](https://go.dev/) which notifies the image updates using an RSS feed at `/feed`.

There are quite a lot of image updater projects such as [watchtower](https://github.com/containrrr/watchtower/), [duin](https://github.com/crazy-max/duin) and a few more.

Unfortunately, none of them provide an RSS feed which I mainly use for keeping track of all the updates to the software that are critical for my homelab and workflow. Therefore I made this tiny program.

## Deploy

A ready-to-use `docker-compose.yaml` is available with only one environment variable to worry about: `UPDATE_SCHEDULE`.

`UPDATE_SCHEDULE` is your regular cron expression which can be adjusted accordingly.

Start the container:

```
docker compose up -d
```

After the server starts, add it as a feed to your favorite RSS reader. Add the `/feed` at the end of the URL, that's where the feeds are published.

## TODO

- support different registries
- properly detect images which are local only
- write tests
- ...

## License

MIT. See `LICENSE` for more details.
