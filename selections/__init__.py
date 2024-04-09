# pylint: disable=wrong-import-position
import os
from collections import defaultdict

from flask import Flask
from flask_migrate import Migrate
from flask_pyoidc.flask_pyoidc import OIDCAuthentication
from flask_sqlalchemy import SQLAlchemy
import sentry_sdk
from sentry_sdk.integrations.flask import FlaskIntegration
from sentry_sdk.integrations.sqlalchemy import SqlalchemyIntegration

# Create the initial Flask Object
app = Flask(__name__)

# Check if deployed on OpenShift, if so use environment.
if os.path.exists(os.path.join(os.getcwd(), 'config.py')):
    app.config.from_pyfile(os.path.join(os.getcwd(), 'config.py'))
else:
    app.config.from_pyfile(os.path.join(os.getcwd(), 'config.env.py'))

auth = OIDCAuthentication(app, issuer=app.config['OIDC_ISSUER'],
                          client_registration_info=app.config['OIDC_CLIENT_CONFIG'])

# Sentry
sentry_sdk.init(
    dsn=app.config['SENTRY_DSN'],
    integrations=[FlaskIntegration(), SqlalchemyIntegration()],
)

# Initalize the SQLAlchemy object and add models.
# Make sure that you run the migrate task before running.
db = SQLAlchemy(app)
from selections.models import *

migrate = Migrate(app, db)

# Load Applications Blueprint
from selections.blueprints.application import *
from selections.blueprints.teams import *

from selections.utils import before_request


@app.route('/')
@auth.oidc_auth
@before_request
def main(info=None):
    print(info)
    print(info['group_list'])
    print('/active_rtp' in info['group_list'])
    is_evals = '/eboard-evaluations' in info['group_list']
    is_rtp = '/active_rtp' in info['group_list']
    member = Members.query.filter_by(username=info['uid']).first()

    all_applications = Applicant.query.all()
    all_users = [u.username for u in Members.query.all()]

    averages = {}
    reviewers = defaultdict(list)
    evaluated = {}
    for applicant in all_applications:
        score_sum = 0
        results = Submission.query.filter_by(
            application=applicant.id,
            medium='Paper').all()
        phone_r = Submission.query.filter_by(
            application=applicant.id,
            medium='Phone').first()
        for result in results:
            score_sum += int(result.score)
            reviewers[applicant.id].append(result.member)
            reviewers[applicant.id] = sorted(reviewers[applicant.id])
        if len(results) != 0:
            avg = int(score_sum / len(results))
            if phone_r:
                avg += phone_r.score
            averages[applicant.id] = avg
        else:
            averages[applicant.id] = 0
            reviewers[applicant.id] = []
        evaluated[applicant.id] = bool(Submission.query.filter_by(application=applicant.id, medium='Phone').all())

    if member and member.team:
        team = Members.query.filter_by(team=member.team)
        reviewed_apps = [a.application for a in Submission.query.filter_by(
            member=info['uid']).all()]
        applications = [
                {
                    'id': a.id,
                    'gender': a.gender,
                    'reviewed': a.id in reviewed_apps,
                    'interview': a.phone_int,
                    'review_count': Submission.query.filter_by(application=a.id).count(),
                    'rit_id': a.rit_id,
                    } for a in Applicant.query.filter_by(team=member.team).all()
                ]

        return render_template(
            'index.html',
            info=info,
            teammates=team,
            applications=applications,
            reviewed_apps=reviewed_apps,
            all_applications=all_applications,
            all_users=all_users,
            averages=averages,
            evaluated=evaluated,
            reviewers=reviewers)
    elif is_evals or is_rtp:
        all_users.append(info['uid'])
        return render_template(
            'index.html',
            info=info,
            all_applications=all_applications,
            all_users=all_users,
            averages=averages,
            evaluated=evaluated,
            reviewers=reviewers)
    else:
        return render_template(
            'index.html',
            info=info,
            all_users=all_users)


@app.route('/logout')
@auth.oidc_logout
def logout():
    return redirect('/', 302)


if __name__ == '__main__':
    app.run()

application = app
