from collections import defaultdict
from zipfile import BadZipFile

import docx
from flask import render_template, redirect, url_for, flash, request

from selections.utils import before_request, assign_pending_applicants
from selections import app, auth
from selections.models import Applicant, Criteria, db, Members, Submission


@app.route('/application/<app_id>')
@auth.oidc_auth
@before_request
def get_application(app_id, info=None):
    reviewed = Submission.query.filter_by(
        id=app_id).filter_by(member=info['uid']).first()
    if reviewed:
        flash('You already reviewed that application!')
        return redirect(url_for('main'))

    applicant_info = Applicant.query.filter_by(id=app_id).first()
    split_body = applicant_info.body.split('\n')
    fields = Criteria.query.filter_by(medium='Paper').all()
    return render_template(
        'vote.html',
        application=applicant_info,
        split_body=split_body,
        info=info,
        fields=fields)


@app.route('/application', methods=['POST'])
@auth.oidc_auth
@before_request
def create_application():
    applicant_id = request.form.get('id')
    applicant = Applicant(
        id=applicant_id,
        body=request.form.get('application'),
        team=request.form.get('team'),
        gender=request.form.get('gender'))
    db.session.add(applicant)
    db.session.flush()
    db.session.commit()
    return get_application_creation()


@app.route('/application/import', methods=['POST'])
@auth.oidc_auth
@before_request
def import_application():
    word_file = request.files['file']
    if not word_file:
        return 'No file', 400

    gender = {'M': 'Male',
              'F': 'Female'}

    unparsed_applications = defaultdict(list)
    applications = {}

    old_apps = [int(app.id) for app in Applicant.query.all()]

    try:
        document = docx.Document(word_file)
    except BadZipFile:
        return 'Not a valid Word file!'

    iteration = 0

    for paragraph in document.paragraphs:
        if 'Entry' not in paragraph.text:
            unparsed_applications[iteration].append(paragraph.text[1:])
        else:
            iteration += 1

    for array in unparsed_applications:
        app_info = unparsed_applications[array][0].split('\t')
        app_id = app_info[0]
        app_gender = gender[app_info[1]]
        app_text = app_info[2]
        if int(app_id) in old_apps:
            # If the application is already in the DB, skip it.
            continue

        for line in unparsed_applications[array][1:]:
            if line[-1:] == ' ':
                app_text += line
            else:
                app_text += '\n{}'.format(line)

        applications[app_id] = [app_gender, app_text]
        new_app = Applicant(
            id=app_id,
            body=app_text,
            team=-1,
            gender=app_gender)
        db.session.add(new_app)
        db.session.flush()
        db.session.commit()

    assign_pending_applicants()

    return get_application_creation()


@app.route('/application/delete/<app_id>', methods=['GET'])
@auth.oidc_auth
@before_request
def delete_application(app_id, info=None):
    is_evals = 'eboard-evaluations' in info['member_info']['group_list']
    is_rtp = 'rtp' in info['member_info']['group_list']
    if is_evals or is_rtp:
        scores = Submission.query.filter_by(application=app_id).all()
        applicant_info = Applicant.query.filter_by(id=app_id).first()
        for score in scores:
            db.session.delete(score)
            db.session.flush()
            db.session.commit()
        db.session.delete(applicant_info)
        db.session.flush()
        db.session.commit()
        return redirect('/', 302)


@app.route('/application/create')
@auth.oidc_auth
@before_request
def get_application_creation(info=None):
    is_evals = 'eboard-evaluations' in info['member_info']['group_list']
    is_rtp = 'rtp' in info['member_info']['group_list']
    if is_evals or is_rtp:
        return render_template('create.html', info=info)
    else:
        flash("You aren't allowed to see that page!")
        return redirect(url_for('main'))


@app.route('/logout')
@auth.oidc_logout
def logout():
    return redirect('/', 302)


@app.route('/application/<app_id>', methods=['POST'])
@auth.oidc_auth
@before_request
def submit_application(app_id, info=None):
    fields = [{
        'value': request.form.get(crit.name),
        'weight': crit.weight,
        'max': crit.max_score,
        'min': crit.min_score} for crit in Criteria.query.filter_by(medium='Paper').all()]
    applicant_info = Applicant.query.filter_by(id=app_id).first()
    member = Members.query.filter_by(username=info['uid']).first()
    submissions = [sub.member for sub in Submission.query.filter_by(
        application=app_id).all()]

    if info['uid'] in submissions:
        flash('You have already reviewed this application!')
        return redirect(url_for('main'))

    if applicant_info.team != member.team:
        flash('You are not on the correct team to review that application!')
        return redirect(url_for('main'))

    for field in fields:
        if not field['min'] <= int(field['value']) <= field['max']:
            flash('Please make sure that the data you submitted is valid!')
            return redirect(url_for('main'))

    total_score = 0
    for field in fields:
        total_score += (int(field['value']) * field['weight'])

    member_score = Submission(
        application=app_id, member=member.username, medium='Paper', score=total_score)
    db.session.add(member_score)
    db.session.flush()
    db.session.commit()
    flash('Thanks for evaluating application #{}!'.format(app_id))
    return redirect('/', 302)


@app.route('/application/review/<app_id>', methods=['GET'])
@auth.oidc_auth
@before_request
def review_application(app_id, info=None):
    applicant_info = Applicant.query.filter_by(id=app_id).first()
    evaluated = bool(Submission.query.filter_by(application=app_id, medium='Phone').all())
    scores = Submission.query.filter_by(application=app_id).all()
    split_body = applicant_info.body.split('\n')
    return render_template(
        'review_app.html',
        info=info,
        application=applicant_info,
        scores=scores,
        split_body=split_body,
        evaluated=evaluated)


@app.route('/application/phone/<app_id>', methods=['GET'])
@auth.oidc_auth
@before_request
def get_phone_application(app_id, info=None):
    applicant_info = Applicant.query.filter_by(id=app_id).first()
    split_body = applicant_info.body.split('\n')
    scores = [subs.score for subs in Submission.query.filter_by(application=app_id).all()]
    total = 0
    if scores:
        for score in scores:
            total += score

        total = total / len(scores)

    return render_template(
            'phone.html',
            info=info,
            app_score=total,
            application=applicant_info,
            split_body=split_body)


@app.route('/application/phone/<app_id>', methods=['POST'])
@auth.oidc_auth
@before_request
def promote_application(app_id, info=None):
    score = request.form.get('score')
    new_submit = Submission(
        application=app_id,
        member=info['uid'],
        medium='Phone',
        score=score)
    db.session.add(new_submit)
    db.session.flush()
    db.session.commit()
    return redirect('/', 302)
