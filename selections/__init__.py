from flask import Flask, render_template, redirect, url_for, flash, request

from flask_pyoidc.flask_pyoidc import OIDCAuthentication
from flask_sqlalchemy import SQLAlchemy
from flask_migrate import Migrate

import os
import csh_ldap

# Create the initial Flask Object
app = Flask(__name__)

# Check if deployed on OpenShift, if so use environment.
if os.path.exists(os.path.join(os.getcwd(), "config.py")):
    app.config.from_pyfile(os.path.join(os.getcwd(), "config.py"))
else:
    app.config.from_pyfile(os.path.join(os.getcwd(), "config.env.py"))

auth = OIDCAuthentication(app, issuer=app.config["OIDC_ISSUER"],
                                  client_registration_info=app.config["OIDC_CLIENT_CONFIG"])

# Create a connection to CSH LDAP
_ldap = csh_ldap.CSHLDAP(app.config['LDAP_BIND_DN'], app.config['LDAP_BIND_PASS'])

from selections.utils import before_request, get_member_info, process_image

# Initalize the SQLAlchemy object and add models.
# Make sure that you run the migrate task before running.
db = SQLAlchemy(app)
from selections.models import *
migrate = Migrate(app, db)

@app.route("/")
@auth.oidc_auth
@before_request
def main(info = None):
    member = members.query.filter_by(username=info['uid']).first()
    if member != None:
        team = members.query.filter_by(team=member.team)
        reviewed_apps = [a.application for a in submission.query.filter_by(member=info['uid']).all()]
        applications = [{
            "id": a.id,
            "gender": a.gender,
            "reviewed": a.id in reviewed_apps} for a in applicant.query.filter_by(team=member.team).all()]
        
        if("eboard-evaluations" in info['member_info']['group_list']):
            all_applications = applicant.query.all()
            all_users = members.query.all()
            return render_template(
                'index.html',
                info = info,
                teammates = team,
                applications = applications,
                reviewed_apps = reviewed_apps,
                all_applications = all_applications,
                all_users = all_users)

        return render_template(
            'index.html',
            info=info,
            reviewed_apps = reviewed_apps,
            teammates=team,
            applications=applications)
    else:
        return "you aren't signed up for selections, leave me alone"
    
@app.route("/application/<app_id>")
@auth.oidc_auth
@before_request
def userroute(app_id, info=None):
    reviewed = submission.query.filter_by(id=app_id).filter_by(member=info['uid']).first()
    if reviewed:
        flash("You already reviewed that application!")
        return redirect(url_for("main"))

    applicant_info = applicant.query.filter_by(id=app_id).first()
    fields = criteria.query.filter_by(medium="Paper").all()
    return render_template(
        "vote.html",
        application = applicant_info,
        info=info,
        fields = fields)

@app.route("/submit_intro_member", methods=["POST"])
@auth.oidc_auth
@before_request
def submit_intro_member(info=None):
    id = request.form.get("id")
    print(request.form)
    member = applicant(
        id = id,
        body = request.form.get("application"),
        team = request.form.get("team"),
        gender = request.form.get("gender"))
    db.session.add(member)
    db.session.flush()
    db.session.commit()
    return(evals())

@app.route("/logout")
@auth.oidc_logout
def logout():
    return redirect("/", 302)

@app.route("/evals")
@auth.oidc_auth
@before_request
def evals(info=None):
    return(render_template("evals.html", info=info))

@app.route("/applicationReview/<app_id>")
@auth.oidc_auth
@before_request
def application_review(app_id, info=None):
    items = intro_members.query.filter_by(id=app_id).filter(intro_members.User_Reviewed != "god")
    team = intro_members.query.filter_by(id=app_id).first().Team
    teamReviewers = selections_users.query.filter_by(team = team)
    return(render_template("applicationReview.html", info=info, members=items, review=app_id, team=team, teamReviewers=teamReviewers))

@app.route("/submit/<app_id>", methods=['POST'])
@auth.oidc_auth
@before_request
def submit(app_id, info=None):
    print(request.form)
    fields = [{
        "value": request.form.get(crit.name),
        "weight": crit.weight,
        "max": crit.max_score,
        "min": crit.min_score} for crit in criteria.query.filter_by(medium="Paper").all()]
    print(fields)
    applicant_info = applicant.query.filter_by(id=app_id).first()
    member = members.query.filter_by(username=info['uid']).first()

    if applicant_info.team != member.team:
        return("You are not allowed to review this application!")
    
    for field in fields:
        if not field["min"] <= int(field["value"]) <= field["max"]:
            return("Please fill out the form correctly!")

    total_score = 0
    for field in fields:
        total_score += (int(field["value"]) * field["weight"])

    member_score = submission(application=app_id, member=member.username, medium="Paper", score=total_score)
    db.session.add(member_score)
    db.session.flush()
    db.session.commit()
    return(main())
    


if __name__ =="__main__":
    app.run()

application = app   
