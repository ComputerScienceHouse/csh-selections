# Selections
But this time, in Go! (and written to the specs that ResLife has so graciously forced upon us)

## How to run Selections
Copy .env.example to .env, and fill out the variables as needed.
```
podman build . --tag=selections
podman run --rm -p 8080:8080 --env-file .env selections
``` 

### Features Implemented
- Autofill for users to add to the Session

### Features Planned
- [x] Automatic Team Creation
  - [x] Team count can be lowered, but maximum team count is # of E-Board members
  - [x] At least one E-Board member per team
- [x] Session teams can be rotated before starting, but not during
- [x] People can be removed from a team
- [x] People can be removed from a team once starting (if they need to leave)
- [x] Applications can be added
- [x] PDF Reader for applications (well, it uses browser PDF reading)
- [x] Rubric on the side
- [ ] Quickly exportable application score information
- [ ] Application scores are viewable on admin page
- [x] Selection stop button deletes applications, memberships, teams, and session_attendances
- [x] Selection stop button requires confirmation
- [ ] JSON error responses are displayed somehow

## Further information
none yet