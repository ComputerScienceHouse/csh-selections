# Preparing selections for a new year
There are a couple steps required to get selections ready for a new selections committee.
To do this, you'll either need to be an RTP, or:
- have access to the selections okd project; and
- have write access to mysql in the selections folder; and
- be evals

## Bring down selections
Go into OKD and set the replicas on the prod deployment to 0. This will ensure selections doesn't do anything weird while you're setting up the db.

## Back up and prep Mysql
The following steps will create an archive of the current selections db, and prep the db for a new year
1. Rename `selections_skyler` to `selections_<last year>`
2. Copy the structure of the table back to `selections_skyler`
3. (optional) Reset the auto increments to 1
4. Copy the criteria

## Bring selections back up
Go back and edit the deployment and restore the number of replicas

## Setup teams
Head to selections and start creating teams
