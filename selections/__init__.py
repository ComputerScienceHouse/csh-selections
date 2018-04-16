import os
from collections import defaultdict

import csh_ldap
from flask import Flask
from flask_migrate import Migrate
from flask_pyoidc.flask_pyoidc import OIDCAuthentication
from flask_sqlalchemy import SQLAlchemy

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
_ldap = csh_ldap.CSHLDAP(
    app.config['LDAP_BIND_DN'], app.config['LDAP_BIND_PASS'])


# Initalize the SQLAlchemy object and add models.
# Make sure that you run the migrate task before running.
db = SQLAlchemy(app)
from selections.models import *

migrate = Migrate(app, db)

# Load Applications Blueprint
from selections.blueprints.application import *
from selections.blueprints.teams import *

from selections.utils import before_request, get_member_info


@app.route("/")
@auth.oidc_auth
@before_request
def main(info=None):
    is_evals = "eboard-evaluations" in info['member_info']['group_list']
    is_rtp = "rtp" in info['member_info']['group_list']
    member = members.query.filter_by(username=info['uid']).first()

    all_applications = applicant.query.all()
    all_users = [u.username for u in members.query.all()]

    averages = {}
    reviewers = defaultdict(list)
    for application in all_applications:
        score_sum = 0
        results = submission.query.filter_by(
            application=application.id,
            medium="Paper").all()
        phone_r = submission.query.filter_by(
            application=application.id,
            medium="Phone").first()
        for result in results:
            score_sum += int(result.score)
            reviewers[application.id].append(result.member)
            reviewers[application.id] = sorted(reviewers[application.id])
        if len(results) != 0:
            avg = int(score_sum / len(results))
            if phone_r:
                avg += phone_r.score
            averages[application.id] = avg
        else:
            averages[application.id] = 0
            reviewers[application.id] = []

    if member and member.team or is_evals or is_rtp:
        team = members.query.filter_by(team=member.team)
        reviewed_apps = [a.application for a in submission.query.filter_by(
            member=info['uid']).all()]
        applications = [{
            "id": a.id,
            "gender": a.gender,
            "reviewed": a.id in reviewed_apps,
            "review_count": submission.query.filter_by(application=a.id).count()} for a in applicant.query.filter_by(team=member.team).all()]

        return render_template(
            'index.html',
            info=info,
            teammates=team,
            applications=applications,
            reviewed_apps=reviewed_apps,
            all_applications=all_applications,
            all_users=all_users,
            averages=averages,
            reviewers=reviewers)


if __name__ == "__main__":
    app.run()

application = app
