from flask import render_template, redirect, url_for, flash, request
from selections.utils import before_request
from selections import app, auth
from selections.models import *

@app.route("/application/<app_id>")
@auth.oidc_auth
@before_request
def get_application(app_id, info=None):
    reviewed = submission.query.filter_by(id=app_id).filter_by(member=info['uid']).first()
    if reviewed:
        flash("You already reviewed that application!")
        return redirect(url_for("main"))

    applicant_info = applicant.query.filter_by(id=app_id).first()
    split_body = applicant_info.body.split("\n")
    fields = criteria.query.filter_by(medium="Paper").all()
    return render_template(
        "vote.html",
        application = applicant_info,
        split_body = split_body,
        info=info,
        fields = fields)


@app.route("/application", methods=["POST"])
@auth.oidc_auth
@before_request
def create_application(info=None):
    id = request.form.get("id")
    member = applicant(
        id = id,
        body = request.form.get("application"),
        team = request.form.get("team"),
        gender = request.form.get("gender"))
    db.session.add(member)
    db.session.flush()
    db.session.commit()
    return(get_application_creation())


@app.route("/application/delete/<app_id>", methods=["GET"])
@auth.oidc_auth
@before_request
def delete_application(app_id, info=None):
    is_evals = "eboard-evaluations" in info['member_info']['group_list']
    is_rtp = "rtp" in info['member_info']['group_list']
    if is_evals or is_rtp:
        scores = submission.query.filter_by(application=app_id).all()
        applicant_info = applicant.query.filter_by(id=app_id).first()
        for score in scores:
            db.session.delete(score)
            db.session.flush()
            db.session.commit()
        db.session.delete(applicant_info)
        db.session.flush()
        db.session.commit()
        return redirect("/", 302)

@app.route("/application/create")
@auth.oidc_auth
@before_request
def get_application_creation(info=None):
    is_evals = "eboard-evaluations" in info['member_info']['group_list']
    is_rtp = "rtp" in info['member_info']['group_list']
    if is_evals or is_rtp:
        return(render_template("admin.html", info=info))
    else:
        flash("You aren't allowed to see that page!")
        return redirect(url_for("main"))

@app.route("/logout")
@auth.oidc_logout
def logout():
    return redirect("/", 302)


@app.route("/application/<app_id>", methods=['POST'])
@auth.oidc_auth
@before_request
def submit_application(app_id, info=None):
    print(request.form)
    fields = [{
        "value": request.form.get(crit.name),
        "weight": crit.weight,
        "max": crit.max_score,
        "min": crit.min_score} for crit in criteria.query.filter_by(medium="Paper").all()]
    print(fields)
    applicant_info = applicant.query.filter_by(id=app_id).first()
    member = members.query.filter_by(username=info['uid']).first()
    submissions = [sub.member for sub in submission.query.filter_by(application=app_id).all()]

    if info['uid'] in submissions:
        flash("You have already reviewed this application!")
        return redirect(url_for("main"))

    if applicant_info.team != member.team:
        flash("You are not on the correct team to review that application!")
        return redirect(url_for("main"))
    
    for field in fields:
        if not field["min"] <= int(field["value"]) <= field["max"]:
            flash("Please make sure that the data you submitted is valid!")
            return redirect(url_for("main"))

    total_score = 0
    for field in fields:
        total_score += (int(field["value"]) * field["weight"])

    member_score = submission(application=app_id, member=member.username, medium="Paper", score=total_score)
    db.session.add(member_score)
    db.session.flush()
    db.session.commit()
    flash("Thanks for evaluating application #{}!".format(app_id))
    return redirect("/", 302)


@app.route("/application/review/<app_id>", methods=['GET'])
@auth.oidc_auth
@before_request
def review_application(app_id, info=None):
    is_evals = "eboard-evaluations" in info['member_info']['group_list']
    is_rtp = "rtp" in info['member_info']['group_list']
    if is_evals or is_rtp:
        applicant_info = applicant.query.filter_by(id=app_id).first()
        scores = submission.query.filter_by(application=app_id).all()
        split_body = applicant_info.body.split("\n")
        return render_template(
            'review_app.html',
            info=info,
            application = applicant_info,
            scores = scores,
            split_body = split_body)
    else:
        flash("You aren't allowed to see that page!")
        return redirect(url_for("main"))
