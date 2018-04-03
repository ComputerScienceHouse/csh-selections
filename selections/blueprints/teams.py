from flask import render_template, redirect, url_for, flash, request

from selections.utils import before_request
from selections import app, auth, db
from selections.models import *


@app.route("/teams")
@auth.oidc_auth
@before_request
def get_teams(info=None):

    team_numbers = set([member.team for member in members.query.all()])

    if None in team_numbers:
        team_numbers.remove(None)

    teams = {}
    for team in team_numbers:
        teams[team] = [
            member.username for member in members.query.filter_by(team=team)]

    return render_template(
        'teams.html',
        info=info,
        team_numbers=sorted(team_numbers),
        teams=teams)


@app.route("/teams", methods=["POST"])
@auth.oidc_auth
@before_request
def create_team(info=None):
    is_evals = "eboard-evaluations" in info['member_info']['group_list']
    is_rtp = "rtp" in info['member_info']['group_list']

    if not is_evals and not is_rtp:
        return "Not Evals or an RTP"

    team_number = request.form.get("number")
    new_members = request.form.get("members")

    usernames = []

    if "," in new_members:
        usernames = new_members.replace(" ", "").split(",")
    else:
        usernames.append(new_members)

    for new_member in usernames:
        member_data = members.query.filter_by(username=new_member).first()
        if member_data:
            member_data.team = team_number
        else:
            person = members(username=new_member, team=team_number)
            print(person.username)
            db.session.add(person)

    db.session.commit()

    return redirect("/teams", 302)


@app.route("/teams/<team_id>", methods=["POST"])
@auth.oidc_auth
@before_request
def add_to_team(team_id, info=None):
    is_evals = "eboard-evaluations" in info['member_info']['group_list']
    is_rtp = "rtp" in info['member_info']['group_list']

    if not is_evals and not is_rtp:
        return "Not Evals or an RTP"

    form_input = request.form.get("username")
    usernames = []

    if "," in form_input:
        usernames = form_input.replace(" ", "").split(",")
    else:
        usernames.append(form_input)

    for new_member in usernames:
        member_data = members.query.filter_by(username=new_member).first()
        if member_data:
            member_data.team = team_id
        else:
            person = members(username=new_member, team=team_id)
            print(person.username)
            db.session.add(person)

    db.session.commit()

    return redirect("/teams", 302)


@app.route("/teams/remove/<username>", methods=["GET"])
@auth.oidc_auth
@before_request
def remove_from_team(username, info=None):
    is_evals = "eboard-evaluations" in info['member_info']['group_list']
    is_rtp = "rtp" in info['member_info']['group_list']

    if not is_evals and not is_rtp:
        return "Not Evals or an RTP"

    member = members.query.filter_by(username=username).first()
    member.team = None
    db.session.commit()
    return redirect("/teams", 302)
