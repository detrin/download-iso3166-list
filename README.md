# download-iso3166-list
This is an unofficial CLI tool for fetching the ISO 3166 country list from iso.org. It uses `go` and `chromedp`.

## Why?
ISO 3166 list changes over time, there are multiple unoficiall lists on the internet and github. There is however only one true surce for ISO 3166 and that is www.iso.org. The list of countries in this repository is checked on weekly basis. 

## Usage
Use directly `countries.json` form this repository. 
```
curl https://raw.githubusercontent.com/detrin/download-iso3166-list/main/countries.json
```
If you wan to fetch the newest list use go
```
go run main.go --mode normal --timeout 60
```
or use docker container
```
docker build -t download-iso3166-list .
docker run --rm -it download-iso3166-list --mode normal --timeout 60
```
The `countries.json` list was generated with 
```
docker run --rm -it my-scraper-app -m slow -t 60 | jq 'sort_by(.Numeric)' > countries.json
```

CLI options
```
Usage:
  main [OPTIONS]

Application Options:
  -m, --mode=        Set the run mode (options: fast, normal, slow) (default: normal)
  -s, --show-window  Show browser window for debugging
  -v, --version      Show version information
  -t, --timeout=     Timeout for the entire scraping session in seconds (default: 60)
  -h, --help         Show help message

Help Options:
  -h, --help         Show this help message
```