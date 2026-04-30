# Selections
But this time, in Go! (and written to the specs that ResLife has so graciously forced upon us)

## How to run Selections
Copy .env.example to .env, and fill out the variables as needed.
```
podman build . --tag=selections
podman run --rm -p 8080:8080 --env-file .env selections
podman run --rm -p 5432:5432 -v ./volume:/var/lib/postgresql/data -e POSTGRES_USER=selections -e POSTGRES_PASSWORD=selectionspassword -e POSTGRES_DB=selections docker.io/postgres:16
``` 

### Features Implemented
- [x] Autofill for users to add to the Session
- [x] Automatic Team Creation
  - [x] Team count can be lowered, but maximum team count is # of E-Board members
  - [x] At least one E-Board member per team
- [x] Session teams can be rotated before starting, but not during
- [x] People can be removed from a team
- [x] People can be removed from a team once starting (if they need to leave)
- [x] Applications can be added
- [x] PDF Reader for applications (well, it uses browser PDF reading)
- [x] Rubric on the side
- [x] Quickly exportable application score information
- [x] Application scores are viewable on admin page
- [x] Selection stop button deletes applications, memberships, teams, and session_attendances
- [x] Selection stop button requires confirmation
- [x] JSON error responses are displayed somehow
- [x] ResLife has the ability to login

## Further information
none yet