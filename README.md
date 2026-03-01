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
- Automatic Team Creation
  - Team count can be lowered, but maximum team count is # of E-Board members
  - At least one E-Board member per team
- Session teams can be rotated before starting, but not during
- People can be removed from a team once starting (if they need to leave)
- Applications are added
- PDF Reader for applications
- Rubric on the side
- Quickly exportable rubric

## Further information
none yet