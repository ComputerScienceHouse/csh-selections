from flask import Flask, render_template, redirect, url_for, flash
import csh_ldap
from flask_sqlalchemy import SQLAlchemy
from flask_pyoidc.flask_pyoidc import OIDCAuthentication
import os
from sqlalchemy import ForeignKey
from flask import request

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
app.secret_key = 'some_secret'

class intro_members(db.Model):
    id = db.Column(db.Integer())
    Social = db.Column(db.Integer())
    Technical = db.Column(db.Integer())
    Creativity = db.Column(db.Integer())
    Activity_Level = db.Column(db.Integer())
    Versatility = db.Column(db.Integer())
    Leadership = db.Column(db.Integer())
    Motivation = db.Column(db.Integer())
    Overall_Feeling = db.Column(db.Integer())
    Application = db.Column(db.String(2000),primary_key = True)
    Team = db.Column(db.Integer())
    Gender = db.Column(db.String(10))
    User_Reviewed = db.Column(db.String(50), ForeignKey("selections_users.username"), primary_key = True)
    def __init__(self, id, Social, Technical, Creativity, Activity_Level, Versatility, Leadership, Motivation, Overall_Feeling, Gender, Application, Team, User_Reviewed):
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
        self.Gender = Gender
        
class selections_users(db.Model):
    username = db.Column(db.String(50), primary_key = True)
    team = db.Column(db.Integer()) 
    def __init__(self, username, team):
        self.username = username
        self.tream = team

information = None

@app.route("/")
@auth.oidc_auth
@before_request
def main(info = None):
    global information
    information = info

    db.create_all()
    user = selections_users.query.filter_by(username=info['uid'])
    if user != None:
        userTeamNumber = user[0].team
        userTeam = selections_users.query.filter_by(team=userTeamNumber)
        userApplications=intro_members.query.filter_by(Team=userTeamNumber).filter(intro_members.User_Reviewed!=user[0].username)
        if("eboard-evaluations" in info['member_info']['group_list']):
            allApplications = intro_members.query.filter_by(User_Reviewed="god")
            allUsers = selections_users.query.all()
            return render_template('index.html', info = info, teammates = userTeam, applications=userApplications, allApplications = allApplications, allUsers = allUsers)

        return render_template('index.html', info=info, teammates=userTeam, applications=userApplications)
    else:
        return "you aren't signed up for selections, leave me alone"
    
@app.route("/application/<variable>")
@auth.oidc_auth
def userroute(variable):
    global information
    if intro_members.query.filter_by(id=variable).filter_by(User_Reviewed=information['uid']).first() != None:
        flash("application already reviewed")
        return redirect(url_for("main"))

    application = intro_members.query.filter_by(id=variable).all()[0]
    return render_template("vote.html", application = application, info=information)

@app.route("/submit_intro_member", methods=["POST"])
@auth.oidc_auth
def submit_intro_member():
    id = request.form.get("id")
    member = intro_members(id =id, Social = 0, Technical=0, Creativity=0, Activity_Level = 0, Versatility= 0, Leadership = 0, Motivation = 0, Overall_Feeling = 0, Application = request.form.get("Application"), Team = request.form.get("Team"), User_Reviewed="god", Gender=request.form.get("Gender"))
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
def evals():
    return(render_template("evals.html", info = information))

@app.route("/applicationReview/<variable>")
def applicationReview(variable):
    items = intro_members.query.filter(intro_members.User_Reviewed != "god")
    team = intro_members.query.filter_by(id=variable).first().Team
    teamReviewers = selections_users.query.filter_by(team = team)
    return(render_template("applicationReview.html", info=information, members=items, review=variable, team=team, teamReviewers=teamReviewers))

@app.route("/submit/<variable>/<variable2>", methods=['POST'])
@auth.oidc_auth
def submit(variable, variable2):
    app = intro_members.query.filter_by(id=variable)[0]
    Social = request.form.get("Social")
    Technical = request.form.get("Technical")
    Creativity = request.form.get("Creativity")
    Activity_Level = request.form.get("Activity_Level")
    Versatility = request.form.get("Versatility")
    Leadership = request.form.get("Leadership")
    Motivation = request.form.get("Motivation")
    Overall_Feeling = request.form.get("Overall_Feeling")
    Application = app.Application
    Team = app.Team
    Gender = app.Gender
    User_Reviewed = selections_users.query.filter_by(username=variable2).first()
    
    if(Social != "" and Technical != "" and Creativity != "" and Activity_Level != "" and Versatility != "" and Leadership != "" and Motivation != "" and Overall_Feeling != "" and int(Social) <= 10 and int(Technical) <= 10 and int(Creativity) <= 10 and int(Activity_Level) <= 10 and int(Versatility) <= 10 and int(Leadership) <= 10 and int(Motivation)<= 10 and int(Overall_Feeling)<= 10):
        member_score = intro_members(id=variable, Social=Social, Technical=Technical, Creativity=Creativity, Activity_Level=Activity_Level, Versatility=Versatility, Leadership=Leadership, Motivation=Motivation, Overall_Feeling=Overall_Feeling,Gender=Gender, Application=Application, Team=Team, User_Reviewed=variable2)
        db.session.add(member_score)
        db.session.flush()
        db.session.commit()
        return(main())
    else:
        return("you didn't fill that out right. try again.")
    


if __name__ =="__main__":
    app.run()

application = app   
