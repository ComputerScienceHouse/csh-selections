from flask import Flask, render_template, flash
import csh_ldap
from flask_sqlalchemy import SQLAlchemy
from flask_pyoidc.flask_pyoidc import OIDCAuthentication
import os
from sqlalchemy import ForeignKey

app = Flask(__name__)
db = SQLAlchemy(app)


if os.path.exists(os.path.join(os.getcwd(), "config.py")):
    app.config.from_pyfile(os.path.join(os.getcwd(), "config.py"))
else:
    app.config.from_pyfile(os.path.join(os.getcwd(), "config.env.py"))

auth = OIDCAuthentication(app, issuer=app.config["OIDC_ISSUER"],
                                  client_registration_info=app.config["OIDC_CLIENT_CONFIG"])

_ldap = csh_ldap.CSHLDAP(app.config['LDAP_BIND_DN'], app.config['LDAP_BIND_PASS'])

from selections.utils import before_request, get_member_info, process_image
db = SQLAlchemy(app)

class intro_members(db.Model):
    id = db.Column(db.Integer(), primary_key=True)
    Social = db.Column(db.Integer())
    Technical = db.Column(db.Integer())
    Creativity = db.Column(db.Integer())
    Activity_Level = db.Column(db.Integer())
    Versatility = db.Column(db.Integer())
    Leadership = db.Column(db.Integer())
    Motivation = db.Column(db.Integer())
    Overall_Feeling = db.Column(db.Integer())
    Application = db.Column(db.String(2000))
    Team = db.Column(db.Integer())
    User_Reviewed = db.Column(db.String(50), ForeignKey("selections_users.username"))
    def __init__(id, Social, Technical, Creativity, Activity_Level, Versatility, Leadership, Motivation, Overall_Feeling, Application, Team, User_Reviewed):
        self.id = id
        self.Social = Social
        self.Technical = Technical
        self.Creativity = Creativity
        self.Activity_Level = Activity_Level
        self.Versatility = Versatility
        self.Leadership = Leadership
        self.Motivation = Motivation
        self.Overall_Feeling = Overall_Feeling
        self.Application = Application
        self.Team = Team
        self.User_Reviewed = User_Reviewed
        
class selections_users(db.Model):
    username = db.Column(db.String(50), primary_key = True)
    team = db.Column(db.Integer()) 
    def __init__(username, team):
        self.username = username
        self.tream = team

@app.route("/")
@auth.oidc_auth
@before_request
def main(info = None):
    db.create_all()
    if ("eboard-evaluations" in info['member_info']['group_list']):
        print('you skyler')

    user = selections_users.query.filter_by(username=info['uid'])
    if user != None:
        userTeamNumber = (user[0].team)
        userTeam = selections_users.query.filter_by(team=userTeamNumber)
        userApplications = intro_members.query.filter_by(Team=userTeamNumber)
        return render_template('index.html', info=info, teammates=userTeam, applications=userApplications)
    else:
        return "you aren't signed up for selections, leave me alone"
    

@app.route("/logout")
@auth.oidc_logout
def logout():
    return redirect("/", 302)


if __name__ =="__main__":
    app.run()

application = app
