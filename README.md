## About

**docker-rss** is a server written in [Go](https://go.dev/) which notifies the image updates using an RSS feed at `/feed`.

There are quite a lot of image updater projects such as [watchtower](https://github.com/containrrr/watchtower/), [duin](https://github.com/crazy-max/duin) and a few more.

Unfortunately, none of them provide an RSS feed which I mainly use for keeping track of all the updates to the software that are critical for my homelab and workflow. Therefore I made this tiny program.

## Deploy

Currently, the way it detects the updates is by first mounting the `/var/run/docker.sock` socket on the `docker-rss` container which will then detect all the running containers and thereby schedule the image update scans from dockerhub.

`docker-compose.yaml`:

```YAML
services:
  docker-rss:
    image: vector450/docker-rss:latest
    container_name: docker-rss
    ports:
      - "8083:8083"
    environment:
      - UPDATE_SCHEDULE=* * * * *
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
```

`UPDATE_SCHEDULE` is your regular cron expression which can be adjusted accordingly.

Start the container:

```
docker compose up -d
```

After the server starts, add it as a feed to your favorite RSS reader. Add the `/feed` at the end of the URL, that's where the feeds are published.

## License

MIT. See `LICENSE` for more details.
