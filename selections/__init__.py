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
_ldap = csh_ldap.CSHLDAP(app.config['LDAP_BIND_DN'], app.config['LDAP_BIND_PASS'])

from selections.utils import before_request, get_member_info, process_image

# Initalize the SQLAlchemy object and add models.
# Make sure that you run the migrate task before running.
db = SQLAlchemy(app)
from selections.models import *

migrate = Migrate(app, db)

# Load Applications Blueprint
from selections.blueprints.application import *


@app.route("/")
@auth.oidc_auth
@before_request
def main(info=None):
    all_applications = []
    all_users = []
    averages = []
    reviewers = []

    is_evals = "eboard-evaluations" in info['member_info']['group_list']
    is_rtp = "rtp" in info['member_info']['group_list']
    member = members.query.filter_by(username=info['uid']).first()

    if is_evals or is_rtp:
        all_applications = applicant.query.all()
        all_users = members.query.all()

        averages = {}
        reviewers = defaultdict(list)
        for application in all_applications:
            score_sum = 0
            results = submission.query.filter_by(application=application.id).all()
            print(results)
            for result in results:
                score_sum += int(result.score)
                reviewers[application.id].append(result.member)
                reviewers[application.id] = sorted(reviewers[application.id])
            if len(results) != 0:
                averages[application.id] = int(score_sum / len(results))
            else:
                averages[application.id] = "N/A"
                reviewers[application.id] = []

    if member and member.team:
        team = members.query.filter_by(team=member.team)
        reviewed_apps = [a.application for a in submission.query.filter_by(member=info['uid']).all()]
        applications = [{
            "id": a.id,
            "gender": a.gender,
            "reviewed": a.id in reviewed_apps} for a in applicant.query.filter_by(team=member.team).all()]

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

    elif is_evals or is_rtp:
        return render_template(
            'index.html',
            info=info,
            all_applications=all_applications,
            all_users=all_users,
            averages=averages,
            reviewers=reviewers)


if __name__ == "__main__":
    app.run()

application = app
