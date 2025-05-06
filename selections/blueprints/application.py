from collections import defaultdict
from zipfile import BadZipFile

from flask import render_template, redirect, url_for, flash, request

from selections.utils import before_request, assign_pending_applicants
from selections import app, auth, s3
from selections.models import Applicant, Criteria, db, Members, Submission


@app.route('/application/<app_id>')
@auth.oidc_auth
@before_request
def get_application(app_id, info=None):
    applicant_info = Applicant.query.filter_by(id=app_id).first()
    member = Members.query.filter_by(username=info['uid']).first()
    is_evals = '/eboard-evaluations' in info['group_list']
    is_rtp = '/active_rtp' in info['group_list']
    if not member and not (is_rtp or is_evals):
        return redirect(url_for('main'))

    reviewed = Submission.query.filter_by(
        id=app_id).filter_by(member=info['uid']).first()
    if reviewed:
        flash('You already reviewed that application!')
        return redirect(url_for('main'))

    return render_template(
        'vote.html',
        application=applicant_info,
        pdf_url='/application/content'+app_id,
        info=info,
        fields=fields)

@app.route('/application/content/<app_id>')
@auth.oidc_auth
def get_application_pdf(app_id):
    applicant_info = Applicant.query.filter_by(id=app_id).first()
    resp = s3.get_object(Bucket=app.config['S3_BUCKET_NAME'], Key='/'+applicant_info.rit_id+'.pdf')
    pdfdata = resp.read()
    return pdfdata


@app.route('/application', methods=['POST'])
@auth.oidc_auth
def create_application():
    applicant_rit_id = request.form.get('rit_id')
    applicant = Applicant(
        team=request.form.get('team'),
        gender=request.form.get('gender'),
        rit_id=applicant_rit_id,
    )
    pdf = request.files['file']
    pdf.save('/tmp/'+pdf.filename)
    s3.upload_file('/tmp/'+pdf.filename, app.config['S3_BUCKET_NAME'], applicant_rit_id+'.pdf')
    db.session.add(applicant)
    db.session.flush()
    db.session.commit()
    return get_application_creation()


@app.route('/application/import', methods=['POST'])
@auth.oidc_auth
#@before_request
def import_application():
    word_file = request.files['file']
    if not word_file:
        return 'No file', 400

    gender = {'M': 'Male',
              'F': 'Female'}

    unparsed_applications = defaultdict(list)
    applications = {}

    old_apps = [app.id for app in Applicant.query.all()]

    document = ""
    iteration = 0

    for paragraph in document.paragraphs:
        if 'Entry' not in paragraph.text:
            unparsed_applications[iteration].append(paragraph.text[1:])
        else:
            iteration += 1

    for array in unparsed_applications:
        app_info = unparsed_applications[array][0].split('\t')
        app_rit_id = app_info[0]
        app_gender = gender[app_info[1]]
        app_text = app_info[2]
        if app_rit_id in old_apps:
            # If the application is already in the DB, skip it.
            continue

        for line in unparsed_applications[array][1:]:
            if line[-1:] == ' ':
                app_text += line
            else:
                app_text += '\n{}'.format(line)

        applications[app_rit_id] = [app_gender, app_text]
        new_app = Applicant(
            rit_id=app_rit_id,
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
    is_evals = '/eboard-evaluations' in info['group_list']
    is_rtp = '/active_rtp' in info['group_list']
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
    flash("You can't delete applications.")
    redirect(url_for('main'))


@app.route('/application/create')
@auth.oidc_auth
@before_request
def get_application_creation(info=None):
    is_evals = '/eboard-evaluations' in info['group_list']
    is_rtp = '/active_rtp' in info['group_list']
    if is_evals or is_rtp:
        return render_template('create.html', info=info)
    else:
        flash("You aren't allowed to see that page!")
        return redirect(url_for('main'))


@app.route('/application/<app_id>', methods=['POST'])
@auth.oidc_auth
@before_request
def submit_application(app_id, info=None):
    member = Members.query.filter_by(username=info['uid']).first()
    if not member:
        flash("You can't score applications.")
        return redirect(url_for('main'))

    fields = [{
        'value': request.form.get(crit.name),
        'weight': crit.weight,
        'max': crit.max_score,
        'min': crit.min_score} for crit in Criteria.query.filter_by(medium='Paper').all()]
    applicant_info = Applicant.query.filter_by(id=app_id).first()
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
    return render_template(
        'review_app.html',
        info=info,
        application=applicant_info,
        scores=scores,
        pdf_url='/application/content'+app_id,
        evaluated=evaluated)


@app.route('/application/phone/<app_id>', methods=['GET'])
@auth.oidc_auth
@before_request
def get_phone_application(app_id, info=None):
    applicant_info = Applicant.query.filter_by(id=app_id).first()
    pdf_url = s3.generate_presigned_url('get_object', Params={'Bucket': app.config['S3_BUCKET_NAME'], 'Key': applicant_info.rit_id+'.pdf'}, ExpiresIn=30)
    pdf_url = pdf_url.replace("s3.csh", "assets.csh")
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
            pdf_url=pdf_url)


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
