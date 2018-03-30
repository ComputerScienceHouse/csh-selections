from flask import Flask, render_template, flash
import csh_ldap
from flask_sqlalchemy import SQLAlchemy
from flask_pyoidc.flask_pyoidc import OIDCAuthentication
import os

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

app.secret_key = "listen, it's real secret"

@app.route("/")
@auth.oidc_auth
@before_request
def main():
    return render_template('index.html')

if __name__ =="__main__":
    app.run()

application = app
