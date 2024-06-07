# download-iso3166-list
This is an unofficial CLI tool for fetching the ISO 3166 country list from iso.org. It uses `go` and `chromedp`.

## Why?
ISO 3166 list changes over time, there are multiple unoficiall lists on the internet and github. There is however only one true surce for ISO 3166 and that is www.iso.org. The list of countries in this repository is checked on weekly basis. 

## Usage
Use directly `countries.json` form this repository. If you wan to fetch the newest list use go
```
go run main.go --mode normal --timeout 60
```
or use docker container
```
docker build -t download-iso3166-list .
docker run --rm -it download-iso3166-list --mode normal --timeout 60
```