#github-usercheck
Simple tool to check https://github.com/ for a given set of usernames (separated by \n) then prints out available usernames.

##Usage
`./github-usercheck <<< ausernamethatnoonehas`

`./github-usercheck -path names.txt -workers 4 >> results.txt`

## Options
```
-path string
    The filepath of the names
-sleep int
    Sleep duration between each workers task. (Millisecond) (default 100)
-workers int
    How many workers to run concurrently. (More workers are faster but more prone to rate limiting or bandwith issues) (default 2)
```
