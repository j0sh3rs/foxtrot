# Foxtrot

A simple golang script to help boost Firefox's numbers on analytics.usa.gov

## What it do

The script is simple:

1. Downlod the list of websites being tracked by the analytics.usa.gov website
2. It will randomly select 10 of them and begin sending https requests against them using the Firefox User Agent.
3. After an hour, it will select 10 new random websites from the list, and repeat.

## Why

Firefox is doing great work for user privacy and anti-tracking advocacy. It was [recently highlighted](https://www.brycewray.com/posts/2023/11/firefox-brink/?utm_source=tldrnewsletter) and cross posted on [hacker news](https://news.ycombinator.com/item?id=38531104) that Firefox is close to being dropped from federal website support and development efforts (according to the USWDS) standards.

Given the importance of Firefox to the community, we need to improve the percentage share of traffic hitting the US Government websites in order to ensure Firefox remains competitive and a priority for developers.

## Usage

```bash
git clone https://github.com/j0sh3rs/foxtrot.git
cd foxtrot
go get
go run ./main.go
```

## Arguments

The script supports a few input arguments to change concurrency, delay between queries, and to customize the User-Agent.

An example of a custom run using all three arguments is:
`go run . --concurrency 5 --delay 30 --user-agent "Custom User Agent/1.0"`

## Roadmap

The script's purpose is simple, but to enable better adoption I will focus on:

- [X] Enabling flags for the number of websites to choose from the list, time between queries (currently random between 1 and 45 seconds), and how long to wait between choosing new sites (1h)
- [X] DNS Caching -- Reduce total load on upstream DNS when using higher concurrency
- [ ] Helm Chart + Container for easy adoption/deployment onto kubernetes clusters