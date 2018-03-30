from flask import Flask, render_template, flash
import csh_ldap
from flask_sqlalchemy import SQLAlchemy
import os

app = Flask(__name__)
db = SQLAlchemy(app)

app.config.from_pyfile(os.path.join(os.getcwd(), "config.env.py"))

app.secret_key = "listen, it's real secret"

@app.route("/")
def main():
    return render_template('index.html')

if __name__ =="__main__":
    app.run()

application = app
