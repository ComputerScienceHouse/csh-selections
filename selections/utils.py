import subprocess
from functools import wraps
from itertools import zip_longest
from flask import session
from math import ceil

from selections import _ldap, db
from selections.ldap import *
from selections.models import applicant, members


def before_request(func):
    @wraps(func)
    def wrapped_function(*args, **kwargs):
        git_revision = subprocess.check_output(['git', 'rev-parse', '--short', 'HEAD']).decode('utf-8').rstrip()
        uuid = str(session["userinfo"].get("sub", ""))
        uid = str(session["userinfo"].get("preferred_username", ""))
        user_obj = _ldap.get_member(uid, uid=True)
        info = {
            "git_revision": git_revision,
            "uuid": uuid,
            "uid": uid,
            "user_obj": user_obj,
            "member_info": get_member_info(uid)
        }
        kwargs["info"] = info
        return func(*args, **kwargs)

    return wrapped_function


def get_member_info(uid):
    account = ldap_get_member(uid)

    if ldap_is_active(account):
        alumInfo = None
    else:
        alumInfo = parse_alum_name(account.gecos)

    member_info = {
        "user_obj": account,
        "group_list": ldap_get_groups(account),
        "uid": account.uid,
        "name": account.cn,
        "alumInfo": alumInfo,
        "active": ldap_is_active(account),
        "onfloor": ldap_is_onfloor(account),
        "room": ldap_get_roomnumber(account),
        "hp": account.housingPoints,
        "plex": account.plex,
        "rn": ldap_get_roomnumber(account)
    }
    return member_info


def assign_pending_applicants():
    pending = applicant.query.filter_by(team=-1).all()
    teams = set([member.team for member in members.query.all()])
    apps_per_team = ceil(len(pending)/len(teams))
    
    div_apps = list(zip_longest(*(iter(pending),) * apps_per_team))

    app_group = 0
    for team in teams:
        for app_data in div_apps[app_group]:
            if app_data:
                app_data.team = int(team)
        app_group += 1
    db.session.flush()
    db.session.commit()
