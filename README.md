# reimager for Reddit

Download images from reddit.

```bash
go mod tidy

go run main.go pics oldschoolcool thewaywewere MilitaryPorn EarthPorn
```

```bash
docker run --restart unless-stopped \
  -v ${PWD}/subreddits:/data/subreddits \
  ghcr.io/cwpearson/reimager --every 1800 --out-dir /data/subreddits \
    pics \
    EarthPorn \
    oldschoolcool \
    thewaywewere
```

1. Register for API access
2. Register an application: https://ssl.reddit.com/prefs/apps
