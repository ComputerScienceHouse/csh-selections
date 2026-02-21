# Selections
But this time, in Go! (and written to the specs that ResLife has so graciously forced upon us)

## How to run Selections
Copy .env.example to .env, and fill out the variables as needed.
```
podman build . --tag=selections
podman run --rm -p 8080:8080 --env-file .env selections
``` 

## Further information
none yet